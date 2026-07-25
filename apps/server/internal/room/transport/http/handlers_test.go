package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	shared "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/domain"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/migrate"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/repository"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/sqlite"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/platform/config"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/platform/idgen"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/room/actor"
	roomsession "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/room/session"
)

func newTestHandler(t *testing.T) (*Handler, *chi.Mux, func()) {
	t.Helper()
	db, err := sqlite.New(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("new db: %v", err)
	}
	if err := migrate.Up(db.Writer()); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	repo := repository.New(db)
	cfg := config.Config{Environment: config.Test, PublicURL: &url.URL{Scheme: "http", Host: "localhost"}}
	h := NewHandler(repo, actor.NewRegistry(repo, nil), cfg, nil)
	r := chi.NewRouter()
	h.RegisterRoutes(r)
	return h, r, func() { db.Close() }
}

func TestCreateRoom(t *testing.T) {
	_, r, cleanup := newTestHandler(t)
	defer cleanup()

	g := idgen.Generator{}
	reqID, _ := g.RequestID()
	body, _ := json.Marshal(map[string]string{
		"displayName": "Mila",
		"mode":        "Coop",
		"difficulty":  "Medium",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/rooms", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", reqID.String())
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	room := resp["room"].(map[string]any)
	if room["state"] != string(shared.RoomLobby) {
		t.Fatalf("expected Lobby, got %v", room["state"])
	}
	cookies := rec.Result().Cookies()
	if len(cookies) == 0 || cookies[0].Name != roomsession.CookieName {
		t.Fatalf("expected room session cookie")
	}
}

func TestCreateRoomRejectsActiveSession(t *testing.T) {
	_, r, cleanup := newTestHandler(t)
	defer cleanup()

	g := idgen.Generator{}
	reqID, _ := g.RequestID()
	body, _ := json.Marshal(map[string]string{
		"displayName": "Mila",
		"mode":        "Coop",
		"difficulty":  "Medium",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/rooms", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", reqID.String())
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}

	// Reuse the cookie for a second create request.
	reqID2, _ := g.RequestID()
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/rooms", bytes.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Idempotency-Key", reqID2.String())
	for _, c := range rec.Result().Cookies() {
		req2.AddCookie(c)
	}
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusConflict {
		t.Fatalf("expected 409 ACTIVE_ROOM_SESSION_EXISTS, got %d", rec2.Code)
	}
}

func TestPreviewRoom(t *testing.T) {
	_, r, cleanup := newTestHandler(t)
	defer cleanup()

	g := idgen.Generator{}
	reqID, _ := g.RequestID()
	body, _ := json.Marshal(map[string]string{
		"displayName": "Mila",
		"mode":        "Coop",
		"difficulty":  "Medium",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/rooms", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", reqID.String())
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
	var createResp map[string]any
	json.Unmarshal(rec.Body.Bytes(), &createResp)
	code := createResp["room"].(map[string]any)["code"].(string)

	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/rooms/"+code, nil)
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec2.Code)
	}
	var preview map[string]any
	json.Unmarshal(rec2.Body.Bytes(), &preview)
	if preview["mode"] != "Coop" {
		t.Fatalf("expected mode Coop in preview")
	}
}

func TestFailedRoomLookupUsesProgressivePrivateRateKey(t *testing.T) {
	handler, _, cleanup := newTestHandler(t)
	defer cleanup()
	const remoteAddress = "203.0.113.9:1234"
	now := time.Date(2026, time.July, 25, 12, 0, 0, 0, time.UTC)

	for attempt, wantDelay := range []int{1, 2, 4, 8, 16} {
		if delay := handler.recordFailedLookup(remoteAddress, now); delay != wantDelay {
			t.Fatalf("attempt %d delay=%d, want %d", attempt+1, delay, wantDelay)
		}
	}
	if retry := handler.lookupRetryAfter(remoteAddress, now); retry < 15 {
		t.Fatalf("retry=%d, want progressive temporary block", retry)
	}
	for key := range handler.lookups {
		if strings.Contains(key, "203.0.113.9") {
			t.Fatal("rate limiter retained raw IP address")
		}
	}
}

func TestRoomCreationRateLimitSeparatesClientAddresses(t *testing.T) {
	handler, _, cleanup := newTestHandler(t)
	defer cleanup()
	now := time.Date(2026, time.July, 25, 12, 0, 0, 0, time.UTC)
	for range 10 {
		if !handler.allowRoomCreation("203.0.113.10", now) {
			t.Fatal("first client was limited too early")
		}
	}
	if handler.allowRoomCreation("203.0.113.10", now) {
		t.Fatal("first client exceeded its hourly limit")
	}
	if !handler.allowRoomCreation("203.0.113.11", now) {
		t.Fatal("second client incorrectly shared the first client's limit")
	}
}

func TestJoinRoom(t *testing.T) {
	_, r, cleanup := newTestHandler(t)
	defer cleanup()

	g := idgen.Generator{}
	reqID, _ := g.RequestID()
	body, _ := json.Marshal(map[string]string{
		"displayName": "Mila",
		"mode":        "Coop",
		"difficulty":  "Medium",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/rooms", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", reqID.String())
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
	var createResp map[string]any
	json.Unmarshal(rec.Body.Bytes(), &createResp)
	code := createResp["room"].(map[string]any)["code"].(string)

	reqID2, _ := g.RequestID()
	joinBody, _ := json.Marshal(map[string]string{"displayName": "Noah"})
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/rooms/"+code+"/join", bytes.NewReader(joinBody))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Idempotency-Key", reqID2.String())
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("expected 200 join, got %d: %s", rec2.Code, rec2.Body.String())
	}
	var joinResp map[string]any
	json.Unmarshal(rec2.Body.Bytes(), &joinResp)
	parts := joinResp["participants"].([]any)
	if len(parts) != 2 {
		t.Fatalf("expected 2 participants, got %d", len(parts))
	}
}

func TestLeaveRoom(t *testing.T) {
	_, r, cleanup := newTestHandler(t)
	defer cleanup()

	g := idgen.Generator{}
	reqID, _ := g.RequestID()
	body, _ := json.Marshal(map[string]string{
		"displayName": "Mila",
		"mode":        "Coop",
		"difficulty":  "Medium",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/rooms", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", reqID.String())
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
	var createResp map[string]any
	json.Unmarshal(rec.Body.Bytes(), &createResp)
	code := createResp["room"].(map[string]any)["code"].(string)

	reqID2, _ := g.RequestID()
	leaveBody, _ := json.Marshal(map[string]string{"intent": "leave_lobby"})
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/rooms/"+code+"/leave", bytes.NewReader(leaveBody))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Idempotency-Key", reqID2.String())
	for _, c := range rec.Result().Cookies() {
		req2.AddCookie(c)
	}
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusNoContent {
		t.Fatalf("expected 204 leave, got %d: %s", rec2.Code, rec2.Body.String())
	}
}
