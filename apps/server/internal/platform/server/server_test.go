package server

import (
	"context"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/migrate"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/repository"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/sqlite"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/platform/config"
)

func newTestServer(t *testing.T) (*Server, *sqlite.DB, net.Listener) {
	t.Helper()
	dir := t.TempDir()
	db, err := sqlite.New(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("new db: %v", err)
	}
	if err := migrate.Up(db.Writer()); err != nil {
		t.Fatalf("migrate up: %v", err)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	origin := "http://" + listener.Addr().String()
	publicURL, err := url.Parse(origin)
	if err != nil {
		t.Fatalf("parse public URL: %v", err)
	}
	cfg := config.Config{
		Environment: config.Test, PublicURL: publicURL, AllowedOrigins: []string{origin},
		AdminProxyHeader: "X-Ninefold-Admin", AdminTrustedProxies: []netip.Prefix{netip.MustParsePrefix("127.0.0.0/8")},
		MatchTombstoneRetention: 30 * 24 * time.Hour,
	}
	repo := repository.New(db)
	instance := New(listener.Addr().String(), "test-version", cfg, db, repo, slog.New(slog.NewTextHandler(io.Discard, nil)))
	return instance, db, listener
}

func TestLivenessAndGracefulShutdown(t *testing.T) {
	instance, _, listener := newTestServer(t)
	stopped := make(chan error, 1)
	go func() {
		stopped <- instance.Serve(listener)
	}()

	client := &http.Client{Timeout: 2 * time.Second}
	response, err := client.Get("http://" + listener.Addr().String() + "/health/live")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusOK || !strings.Contains(string(body), `"version":"test-version"`) {
		t.Fatalf("unexpected liveness response: status=%d body=%s", response.StatusCode, body)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := instance.Shutdown(ctx); err != nil {
		t.Fatal(err)
	}
	if err := <-stopped; err != nil {
		t.Fatal(err)
	}
}

func TestReadinessWhenMigrated(t *testing.T) {
	instance, _, listener := newTestServer(t)
	stopped := make(chan error, 1)
	go func() {
		stopped <- instance.Serve(listener)
	}()

	client := &http.Client{Timeout: 2 * time.Second}
	response, err := client.Get("http://" + listener.Addr().String() + "/health/ready")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusOK || !strings.Contains(string(body), `"status":"ready"`) {
		t.Fatalf("unexpected readiness response: status=%d body=%s", response.StatusCode, body)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := instance.Shutdown(ctx); err != nil {
		t.Fatal(err)
	}
	if err := <-stopped; err != nil {
		t.Fatal(err)
	}
}

func TestPrivateOperationsAndSecurityHeaders(t *testing.T) {
	instance, _, listener := newTestServer(t)
	stopped := make(chan error, 1)
	go func() { stopped <- instance.Serve(listener) }()
	baseURL := "http://" + listener.Addr().String()
	client := &http.Client{Timeout: 2 * time.Second}

	request, _ := http.NewRequest(http.MethodGet, baseURL+"/health/status", nil)
	response, err := client.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	_ = response.Body.Close()
	if response.StatusCode != http.StatusNotFound {
		t.Fatalf("untrusted status=%d", response.StatusCode)
	}

	request, _ = http.NewRequest(http.MethodGet, baseURL+"/health/status", nil)
	request.Header.Set("X-Ninefold-Admin", "test-operator")
	response, err = client.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(response.Body)
	_ = response.Body.Close()
	if response.StatusCode != http.StatusOK || strings.Contains(string(body), "cookie") {
		t.Fatalf("protected status=%d body=%s", response.StatusCode, body)
	}
	for _, header := range []string{"Content-Security-Policy", "X-Content-Type-Options", "Permissions-Policy"} {
		if response.Header.Get(header) == "" {
			t.Errorf("missing %s", header)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := instance.Shutdown(ctx); err != nil {
		t.Fatal(err)
	}
	if err := <-stopped; err != nil {
		t.Fatal(err)
	}
}

func TestNotReadyRejectsNewRoomsAndJoins(t *testing.T) {
	instance, _, listener := newTestServer(t)
	stopped := make(chan error, 1)
	go func() { stopped <- instance.Serve(listener) }()
	instance.ready.Store(false)

	client := &http.Client{Timeout: 2 * time.Second}
	for _, path := range []string{"/api/v1/rooms", "/api/v1/rooms/ABC123/join"} {
		response, err := client.Post("http://"+listener.Addr().String()+path, "application/json", strings.NewReader("{}"))
		if err != nil {
			t.Fatal(err)
		}
		_ = response.Body.Close()
		if response.StatusCode != http.StatusServiceUnavailable {
			t.Fatalf("%s status=%d, want %d", path, response.StatusCode, http.StatusServiceUnavailable)
		}
		if response.Header.Get("Retry-After") == "" {
			t.Fatalf("%s missing Retry-After", path)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := instance.Shutdown(ctx); err != nil {
		t.Fatal(err)
	}
	if err := <-stopped; err != nil {
		t.Fatal(err)
	}
}
