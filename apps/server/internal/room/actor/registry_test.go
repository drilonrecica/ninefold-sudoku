package actor

import (
	"context"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	shared "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/domain"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/migrate"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/repository"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/sqlite"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/platform/idgen"
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

func TestFileBackedRestartEntersRecoveryAndResumesOnReconnect(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "restart.db")
	db, err := sqlite.New(databasePath)
	if err != nil {
		t.Fatal(err)
	}
	if err := migrate.Up(db.Writer()); err != nil {
		t.Fatal(err)
	}
	repo := repository.New(db)
	defer db.Close()
	createPuzzle(t, repo)
	activeActor, room, _ := createRoomAndActor(t, repo)
	ctx := context.Background()
	connID := shared.ConnectionID("01900000-0000-7fff-8000-000000000099")
	outbound := make(chan []byte, 64)
	if _, err := activeActor.Subscribe(ctx, connID, room.Participants[0].ID, outbound); err != nil {
		t.Fatal(err)
	}

	g := idgen.Generator{}
	readyID, _ := g.RequestID()
	if _, err := activeActor.Submit(ctx, Envelope{
		RequestID:    readyID,
		CommandType:  "room.set_ready",
		ConnectionID: connID,
		Fingerprint:  "ready",
		Command: roomdomain.SetReadyCommand{
			Meta: shared.CommandMetadata{
				RequestID: readyID, AuthenticatedParticipantID: room.Participants[0].ID,
				ClientSequence: 1, Target: shared.NewRoomTarget(room.ID), ExpectedVersion: 1,
			},
			Ready: true,
		},
	}); err != nil {
		t.Fatal(err)
	}
	startID, _ := g.RequestID()
	if _, err := activeActor.Submit(ctx, Envelope{
		RequestID:    startID,
		CommandType:  "room.start_countdown",
		ConnectionID: connID,
		Fingerprint:  "start",
		Command: roomdomain.StartCountdownCommand{Meta: shared.CommandMetadata{
			RequestID: startID, AuthenticatedParticipantID: room.Participants[0].ID,
			ClientSequence: 2, Target: shared.NewRoomTarget(room.ID), ExpectedVersion: 2,
		}},
	}); err != nil {
		t.Fatal(err)
	}

	deadline := time.Now().Add(5 * time.Second)
	for {
		matches, err := repo.ListNonTerminalMatches(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if len(matches) == 1 && matches[0].State == string(shared.MatchActive) {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("match did not activate")
		}
		time.Sleep(10 * time.Millisecond)
	}
	activeActor.Stop()

	process := exec.Command(os.Args[0], "-test.run=^TestRecoveryHelperProcess$")
	process.Env = append(os.Environ(), "NINEFOLD_RECOVERY_HELPER_DB="+databasePath)
	if output, err := process.CombinedOutput(); err != nil {
		t.Fatalf("recovery helper process failed: %v\n%s", err, output)
	}

	restarted := NewRegistry(repo, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err := restarted.RecoverNonTerminal(ctx, time.Now()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(restarted.ShutdownAll)
	recoveredActor, err := restarted.Acquire(ctx, room.ID)
	if err != nil {
		t.Fatal(err)
	}
	if recoveredActor.match.State != shared.MatchRecoveryPending {
		t.Fatalf("expected RecoveryPending after restart, got %s", recoveredActor.match.State)
	}
	time.Sleep(2 * time.Millisecond)
	reconnectID := shared.ConnectionID("01900000-0000-7fff-8000-000000000100")
	if _, err := recoveredActor.Subscribe(ctx, reconnectID, room.Participants[0].ID, make(chan []byte, 16)); err != nil {
		t.Fatal(err)
	}
	if recoveredActor.match.State != shared.MatchActive {
		t.Fatalf("expected Active after eligible reconnect, got %s", recoveredActor.match.State)
	}
	if recoveredActor.match.PausedMilliseconds == 0 {
		t.Fatal("expected the server-caused recovery interval to be excluded")
	}
}

func TestRecoveryHelperProcess(t *testing.T) {
	databasePath := os.Getenv("NINEFOLD_RECOVERY_HELPER_DB")
	if databasePath == "" {
		return
	}
	db, err := sqlite.New(databasePath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := repository.New(db)
	registry := NewRegistry(repo, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err := registry.RecoverNonTerminal(context.Background(), time.Now()); err != nil {
		t.Fatal(err)
	}
	matches, err := repo.ListNonTerminalMatches(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 1 || matches[0].State != string(shared.MatchRecoveryPending) {
		t.Fatalf("expected one RecoveryPending match, got %#v", matches)
	}
}
