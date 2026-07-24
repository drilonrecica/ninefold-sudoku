package actor

import (
	"context"
	"io"
	"log/slog"
	"path/filepath"
	"testing"
	"time"

	shared "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/domain"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/migrate"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/repository"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/sqlite"
	roomdomain "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/room/domain"
)

func TestRegistryActivateReleaseAcquire(t *testing.T) {
	db, _ := sqlite.New(filepath.Join(t.TempDir(), "test.db"))
	migrate.Up(db.Writer())
	repo := repository.New(db)
	reg := NewRegistry(repo, slog.New(slog.NewTextHandler(io.Discard, nil)))

	name, _ := shared.NewDisplayName("Host")
	host := roomdomain.Participant{ID: shared.ParticipantID("01900000-0000-7fff-8000-000000000001"), Name: name, Role: shared.RolePlayer}
	room, _ := roomdomain.NewRoom(shared.RoomID("01900000-0000-7fff-8000-000000000000"), "ABCDEF", host, roomdomain.MatchRules{Mode: shared.ModeCoop, Difficulty: shared.DifficultyEasy, ErrorPreset: shared.ErrorPresetCasual}, time.Now())

	a := reg.Activate(room)
	reg.Release(room.ID)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	got, err := reg.Acquire(ctx, room.ID)
	if err != nil {
		t.Fatalf("acquire: %v", err)
	}
	if got != a {
		t.Log("new actor loaded")
	}
	ch := make(chan []byte, 8)
	_, err = got.Subscribe(ctx, shared.ConnectionID("01900000-0000-7fff-8000-000000000002"), host.ID, ch)
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
}
