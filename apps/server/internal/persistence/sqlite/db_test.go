package sqlite

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/migrate"
)

func newTestDB(t *testing.T) *DB {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")
	db, err := New(path)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := migrate.Up(db.Writer()); err != nil {
		t.Fatalf("migrate up: %v", err)
	}
	return db
}

func TestNewRejectsInMemory(t *testing.T) {
	_, err := New(":memory:")
	if err == nil {
		t.Fatal("expected error for :memory:")
	}
}

func TestNewCreatesParentDirectory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "test.db")
	db, err := New(path)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer db.Close()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("database file not created: %v", err)
	}
}

func TestVersionIsReported(t *testing.T) {
	db := newTestDB(t)
	if db.Version() == "" {
		t.Fatal("expected non-empty SQLite version")
	}
	parts := strings.Split(db.Version(), ".")
	if len(parts) != 3 {
		t.Fatalf("expected major.minor.patch version, got %q", db.Version())
	}
}

func TestHealthRequiresWAL(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "legacy.db")
	// Create an old-style database file without WAL.
	if _, err := os.Create(path); err != nil {
		t.Fatalf("create file: %v", err)
	}
	db, err := New(path)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer db.Close()
	if err := migrate.Up(db.Writer()); err != nil {
		t.Fatalf("migrate up: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := db.Health(ctx); err != nil {
		t.Fatalf("health: %v", err)
	}
	var mode string
	if err := db.Writer().QueryRow("PRAGMA journal_mode;").Scan(&mode); err != nil {
		t.Fatalf("read journal mode: %v", err)
	}
	if !strings.EqualFold(mode, "wal") {
		t.Fatalf("journal_mode=%s, want wal", mode)
	}
}

func TestReaderPoolSeparate(t *testing.T) {
	db := newTestDB(t)
	if db.Readers() == db.Writer() {
		t.Fatal("expected separate reader pool")
	}
	ctx := context.Background()
	if err := db.Readers().PingContext(ctx); err != nil {
		t.Fatalf("reader ping: %v", err)
	}
}

func TestMinimumVersionCheck(t *testing.T) {
	// The current bundled SQLite should be far newer than 3.40.0.
	cmp, err := compareSQLiteVersions("3.40.0", "3.40.0")
	if err != nil {
		t.Fatalf("compare: %v", err)
	}
	if cmp != 0 {
		t.Fatalf("expected equal comparison, got %d", cmp)
	}
	cmp, err = compareSQLiteVersions("3.39.9", "3.40.0")
	if err != nil {
		t.Fatalf("compare: %v", err)
	}
	if cmp != -1 {
		t.Fatalf("expected -1, got %d", cmp)
	}
}

func TestCloseIsIdempotent(t *testing.T) {
	db := newTestDB(t)
	if err := db.Close(); err != nil {
		t.Fatalf("first close: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("second close: %v", err)
	}
}
