package maintenance

import (
	"context"
	"io"
	"log/slog"
	"path/filepath"
	"testing"
	"time"

	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/migrate"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/repository"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/sqlite"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/platform/config"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/room/actor"
)

func TestRunOnceIsCancellableAndIdempotent(t *testing.T) {
	t.Parallel()
	db, err := sqlite.New(filepath.Join(t.TempDir(), "maintenance.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := migrate.Up(db.Writer()); err != nil {
		t.Fatal(err)
	}
	repo := repository.New(db)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	registry := actor.NewRegistry(repo, logger)
	defer registry.ShutdownAll()
	scheduler := New(repo, db, registry, config.Config{MatchTombstoneRetention: 30 * 24 * time.Hour}, logger)
	now := time.Unix(1_800_000_000, 0)
	if err := scheduler.RunOnce(context.Background(), now); err != nil {
		t.Fatal(err)
	}
	if err := scheduler.RunOnce(context.Background(), now); err != nil {
		t.Fatalf("rerun: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := scheduler.RunOnce(ctx, now); err == nil {
		t.Fatal("cancelled run should stop")
	}
}
