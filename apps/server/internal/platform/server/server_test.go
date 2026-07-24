package server

import (
	"context"
	"io"
	"log/slog"
	"net"
	"net/http"
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
	cfg := config.Config{Environment: config.Test, PublicURL: &url.URL{Scheme: "http", Host: "localhost"}}
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
