package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"

	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/repository"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/platform/idgen"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/puzzle/catalog"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/room/session"
	"github.com/drilonrecica/ninefold-sudoku/contracts/generated/go/realtime"
)

func TestWebSocketRejectsMissingSession(t *testing.T) {
	instance, _, listener := newTestServer(t)
	stopped := make(chan error, 1)
	go func() { stopped <- instance.Serve(listener) }()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	wsURL := "ws://" + listener.Addr().String() + "/ws"
	conn, _, err := websocket.Dial(ctx, wsURL, &websocket.DialOptions{
		HTTPClient: &http.Client{Timeout: 2 * time.Second},
		HTTPHeader: map[string][]string{"Origin": {"http://" + listener.Addr().String()}},
	})
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	var msg realtime.ServerMessage
	if err := wsjson.Read(ctx, conn, &msg); err != nil {
		t.Fatalf("read rejection: %v", err)
	}
	if string(msg.Type) != "connection.rejected" {
		t.Fatalf("expected connection.rejected, got %s", msg.Type)
	}

	shutdownAndWait(t, instance, stopped)
}

func TestWebSocketRejectsUntrustedOrigin(t *testing.T) {
	instance, _, listener := newTestServer(t)
	stopped := make(chan error, 1)
	go func() { stopped <- instance.Serve(listener) }()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	wsURL := "ws://" + listener.Addr().String() + "/ws"
	_, response, err := websocket.Dial(ctx, wsURL, &websocket.DialOptions{
		HTTPHeader: map[string][]string{"Origin": {"https://attacker.invalid"}},
	})
	if err == nil {
		t.Fatal("expected untrusted origin to be rejected")
	}
	if response == nil || response.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 response, got %+v", response)
	}

	shutdownAndWait(t, instance, stopped)
}

func TestWebSocketLobbyAndGameplay(t *testing.T) {
	instance, db, listener := newTestServer(t)
	repo := repository.New(db)
	createCatalogPuzzle(t, repo)

	stopped := make(chan error, 1)
	go func() { stopped <- instance.Serve(listener) }()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	baseURL := "http://" + listener.Addr().String()
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("cookie jar: %v", err)
	}
	httpClient := &http.Client{Jar: jar, Timeout: 2 * time.Second}
	_ = createRoomHTTP(t, httpClient, baseURL)

	// Allow the Room actor goroutine to start before subscribing.
	time.Sleep(200 * time.Millisecond)

	wsURL := "ws://" + listener.Addr().String() + "/ws"
	conn, _, err := websocket.Dial(ctx, wsURL, &websocket.DialOptions{
		HTTPClient: httpClient,
		HTTPHeader: map[string][]string{"Origin": {baseURL}},
	})
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "test done")

	// Initial connection snapshot.
	accepted := readServerMessage(t, ctx, conn, "connection.accepted")
	if accepted.Payload.Identity == nil || accepted.Payload.Identity.Role == nil || *accepted.Payload.Identity.Role != "Player" {
		t.Fatalf("expected accepted player identity, got %+v", accepted.Payload.Identity)
	}
	roomSnap := readServerMessage(t, ctx, conn, "room.snapshot")

	// Send ready.
	reqID := newRequestID(t)
	wsjson.Write(ctx, conn, realtime.ClientMessage{
		SchemaVersion:  1,
		RequestId:      realtime.Uuidv7(reqID),
		ClientSequence: 1,
		Type:           realtime.ClientMessageTypeRoomSetReady,
		Target:         roomTarget(roomSnap),
		Payload:        realtime.ClientMessagePayload{Ready: boolPtr(true)},
	})
	ack := readServerMessage(t, ctx, conn, "command.acknowledged")
	if ack.Payload.RequestId == nil || string(*ack.Payload.RequestId) != reqID {
		t.Fatalf("expected ack for request %s, got %+v", reqID, ack.Payload.RequestId)
	}
	roomSnap = readServerMessage(t, ctx, conn, "room.snapshot")

	// Start countdown.
	reqID = newRequestID(t)
	wsjson.Write(ctx, conn, realtime.ClientMessage{
		SchemaVersion:  1,
		RequestId:      realtime.Uuidv7(reqID),
		ClientSequence: 2,
		Type:           realtime.ClientMessageTypeRoomStartCountdown,
		Target:         roomTarget(roomSnap),
		Payload:        realtime.ClientMessagePayload{},
	})
	ack = readServerMessage(t, ctx, conn, "command.acknowledged")
	if ack.Payload.RequestId == nil || string(*ack.Payload.RequestId) != reqID {
		t.Fatalf("expected ack for start countdown, got %+v", ack.Payload.RequestId)
	}
	roomSnap = readServerMessage(t, ctx, conn, "room.snapshot")
	if roomSnap.Payload.Room == nil {
		t.Fatalf("expected room snapshot payload")
	}
	roomState, ok := roomSnap.Payload.Room["state"].(string)
	if !ok || roomState != "Countdown" {
		t.Fatalf("expected room state Countdown, got %v", roomSnap.Payload.Room["state"])
	}

	// Wait for countdown activation and sync to get the match board.
	time.Sleep(3500 * time.Millisecond)
	reqID = newRequestID(t)
	wsjson.Write(ctx, conn, realtime.ClientMessage{
		SchemaVersion:  1,
		RequestId:      realtime.Uuidv7(reqID),
		ClientSequence: 3,
		Type:           realtime.ClientMessageTypeConnectionInitialize,
		Payload:        realtime.ClientMessagePayload{},
	})
	_ = readServerMessage(t, ctx, conn, "connection.status")
	roomSnap = readServerMessage(t, ctx, conn, "room.snapshot")
	matchSnap := readServerMessage(t, ctx, conn, "match.snapshot")
	matchID := matchIDFromSnapshot(matchSnap)
	matchVersion := matchSnap.AggregateVersion

	// Place a value on the first non-clue cell.
	cell, value := firstEmptyCell(matchSnap)
	reqID = newRequestID(t)
	wsjson.Write(ctx, conn, realtime.ClientMessage{
		SchemaVersion:  1,
		RequestId:      realtime.Uuidv7(reqID),
		ClientSequence: 4,
		Type:           realtime.ClientMessageTypeMatchPlaceValue,
		Target: &realtime.ClientMessageTarget{
			Kind:            realtime.ClientMessageTargetKindMatch,
			Id:              realtime.Uuidv7(matchID),
			ExpectedVersion: realtime.SafeInteger(matchVersion),
		},
		Payload: realtime.ClientMessagePayload{
			Cell:  uint8Ptr(cell),
			Value: uint8Ptr(value),
		},
	})
	ack = readServerMessage(t, ctx, conn, "command.acknowledged")
	if ack.Payload.RequestId == nil || string(*ack.Payload.RequestId) != reqID {
		t.Fatalf("expected ack for place value, got %+v", ack.Payload.RequestId)
	}
	evt := readServerMessage(t, ctx, conn, "match.event")
	if evt.Payload.Event == nil || evt.Payload.Event["type"] != "ValuePlaced" {
		t.Fatalf("expected ValuePlaced event, got %+v", evt.Payload.Event)
	}

	shutdownAndWait(t, instance, stopped)
}

func TestWebSocketDuplicateRequestAcrossReconnect(t *testing.T) {
	instance, db, listener := newTestServer(t)
	stopped := make(chan error, 1)
	go func() { stopped <- instance.Serve(listener) }()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	baseURL := "http://" + listener.Addr().String()
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("cookie jar: %v", err)
	}
	client := &http.Client{Jar: jar, Timeout: 2 * time.Second}
	_ = createRoomHTTP(t, client, baseURL)
	baseParsed, _ := url.Parse(baseURL)
	cookies := jar.Cookies(baseParsed)
	if len(cookies) != 1 {
		t.Fatalf("expected one room cookie, got %d", len(cookies))
	}
	cookieHeader := cookies[0].String()

	dial := func() *websocket.Conn {
		conn, _, dialErr := websocket.Dial(ctx, "ws://"+listener.Addr().String()+"/ws", &websocket.DialOptions{
			HTTPClient: client,
			HTTPHeader: map[string][]string{"Origin": {baseURL}, "Cookie": {cookieHeader}},
		})
		if dialErr != nil {
			t.Fatalf("dial: %v", dialErr)
		}
		_ = readServerMessage(t, ctx, conn, "connection.accepted")
		return conn
	}

	conn := dial()
	roomSnapshot := readServerMessage(t, ctx, conn, "room.snapshot")
	requestID := newRequestID(t)
	command := realtime.ClientMessage{
		SchemaVersion:  1,
		RequestId:      realtime.Uuidv7(requestID),
		ClientSequence: 1,
		Type:           realtime.ClientMessageTypeRoomSetReady,
		Target:         roomTarget(roomSnapshot),
		Payload:        realtime.ClientMessagePayload{Ready: boolPtr(true)},
	}
	if err := wsjson.Write(ctx, conn, command); err != nil {
		t.Fatalf("write first command: %v", err)
	}
	firstAck := readServerMessage(t, ctx, conn, "command.acknowledged")
	_ = readServerMessage(t, ctx, conn, "room.snapshot")
	_ = conn.Close(websocket.StatusNormalClosure, "reconnect")
	time.Sleep(50 * time.Millisecond)
	if _, err := repository.New(db).GetRoomSessionByHash(ctx, session.Hash(cookies[0].Value)); err != nil {
		t.Fatalf("room session disappeared after disconnect: %v", err)
	}

	conn = dial()
	defer conn.Close(websocket.StatusNormalClosure, "test done")
	_ = readServerMessage(t, ctx, conn, "room.snapshot")
	if err := wsjson.Write(ctx, conn, command); err != nil {
		t.Fatalf("write duplicate command: %v", err)
	}
	duplicateAck := readServerMessage(t, ctx, conn, "command.acknowledged")
	if firstAck.Payload.RequestId == nil || duplicateAck.Payload.RequestId == nil ||
		*firstAck.Payload.RequestId != *duplicateAck.Payload.RequestId {
		t.Fatal("duplicate request did not return the original terminal outcome")
	}

	shutdownAndWait(t, instance, stopped)
}

func TestWebSocketHundredConnectionSmoke(t *testing.T) {
	instance, _, listener := newTestServer(t)
	stopped := make(chan error, 1)
	go func() { stopped <- instance.Serve(listener) }()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	baseURL := "http://" + listener.Addr().String()
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("cookie jar: %v", err)
	}
	client := &http.Client{Jar: jar, Timeout: 3 * time.Second}
	_ = createRoomHTTP(t, client, baseURL)
	baseParsed, _ := url.Parse(baseURL)
	cookies := jar.Cookies(baseParsed)
	if len(cookies) != 1 {
		t.Fatalf("expected one room cookie, got %d", len(cookies))
	}
	cookieHeader := cookies[0].String()

	connections := make([]*websocket.Conn, 0, 100)
	for i := 0; i < 100; i++ {
		conn, _, dialErr := websocket.Dial(ctx, "ws://"+listener.Addr().String()+"/ws", &websocket.DialOptions{
			HTTPClient: client,
			HTTPHeader: map[string][]string{"Origin": {baseURL}, "Cookie": {cookieHeader}},
		})
		if dialErr != nil {
			t.Fatalf("dial connection %d: %v", i+1, dialErr)
		}
		_ = readServerMessage(t, ctx, conn, "connection.accepted")
		_ = readServerMessage(t, ctx, conn, "room.snapshot")
		connections = append(connections, conn)
	}
	for _, conn := range connections {
		_ = conn.Close(websocket.StatusNormalClosure, "smoke complete")
	}

	shutdownAndWait(t, instance, stopped)
}

func createRoomHTTP(t *testing.T, client *http.Client, baseURL string) string {
	t.Helper()
	reqID, err := idgen.Generator{}.RequestID()
	if err != nil {
		t.Fatalf("request id: %v", err)
	}
	body := fmt.Sprintf(`{"displayName":"%s","mode":"%s","difficulty":"%s"}`, "Host", "Coop", "Easy")
	req, err := http.NewRequest(http.MethodPost, baseURL+"/api/v1/rooms", strings.NewReader(body))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", reqID.String())
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("create room: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("create room status %d: %s", resp.StatusCode, b)
	}
	var result struct {
		Room struct {
			Code string `json:"code"`
		} `json:"room"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode create room: %v", err)
	}
	return result.Room.Code
}

func createCatalogPuzzle(t *testing.T, repo *repository.Repository) {
	t.Helper()
	records, err := catalog.ReadFile("../../puzzle/catalog/catalog.jsonl")
	if err != nil {
		t.Fatalf("read catalog: %v", err)
	}
	created := 0
	for _, record := range records {
		if record.Difficulty != "Easy" || record.Lifecycle != "Active" {
			continue
		}
		if _, err := repo.GetPuzzle(context.Background(), record.ID, int64(record.Revision)); err != nil {
			t.Fatalf("get seeded puzzle %s: %v", record.ID, err)
		}
		created++
	}
	if created < 2 {
		t.Fatalf("expected at least two active Easy puzzles, got %d", created)
	}
}

func readServerMessage(t *testing.T, ctx context.Context, conn *websocket.Conn, wantType string) realtime.ServerMessage {
	t.Helper()
	for {
		var msg realtime.ServerMessage
		if err := wsjson.Read(ctx, conn, &msg); err != nil {
			t.Fatalf("read websocket: %v", err)
		}
		if string(msg.Type) == wantType {
			return msg
		}
		if string(msg.Type) == "command.rejected" {
			code := ""
			if msg.Payload.Code != nil {
				code = *msg.Payload.Code
			}
			t.Fatalf("unexpected rejection for %s: %s", wantType, code)
		}
	}
}

func roomTarget(snap realtime.ServerMessage) *realtime.ClientMessageTarget {
	room := snap.Payload.Room
	id, _ := room["id"].(string)
	return &realtime.ClientMessageTarget{
		Kind:            realtime.ClientMessageTargetKindRoom,
		Id:              realtime.Uuidv7(id),
		ExpectedVersion: realtime.SafeInteger(snap.AggregateVersion),
	}
}

func matchIDFromSnapshot(snap realtime.ServerMessage) string {
	match := snap.Payload.Match
	id, _ := match["id"].(string)
	return id
}

func firstEmptyCell(snap realtime.ServerMessage) (uint8, uint8) {
	cells, ok := snap.Payload.Match["cells"].([]any)
	if !ok {
		panic(fmt.Sprintf("bad cells: %T", snap.Payload.Match["cells"]))
	}
	for _, c := range cells {
		cell, ok := c.(map[string]any)
		if !ok {
			continue
		}
		if clue, _ := cell["isClue"].(bool); clue {
			continue
		}
		idx, _ := cell["index"].(float64)
		return uint8(idx), 1
	}
	return 0, 1
}

func newRequestID(t *testing.T) string {
	t.Helper()
	id, err := idgen.Generator{}.RequestID()
	if err != nil {
		t.Fatalf("request id: %v", err)
	}
	return id.String()
}

func shutdownAndWait(t *testing.T, instance *Server, stopped <-chan error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := instance.Shutdown(ctx); err != nil {
		t.Fatalf("shutdown: %v", err)
	}
	if err := <-stopped; err != nil {
		t.Fatalf("serve stopped: %v", err)
	}
}

func boolPtr(v bool) *bool    { return &v }
func uint8Ptr(v uint8) *uint8 { return &v }

func init() {
	// Silence logs during websocket tests.
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	slog.SetDefault(logger)
}

func TestMain(m *testing.M) {
	// Ensure absolute path resolution for catalog tests from the apps/server root.
	if _, err := filepath.Abs("."); err != nil {
		panic(err)
	}
	m.Run()
}
