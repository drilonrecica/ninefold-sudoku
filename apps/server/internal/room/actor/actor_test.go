package actor

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	shared "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/domain"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/gen"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/migrate"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/repository"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/sqlite"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/platform/idgen"
	roomdomain "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/room/domain"
	roomsession "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/room/session"
)

func newTestActorRepo(t *testing.T) (*repository.Repository, *sqlite.DB) {
	t.Helper()
	db, err := sqlite.New(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("new db: %v", err)
	}
	if err := migrate.Up(db.Writer()); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return repository.New(db), db
}

func createPuzzle(t *testing.T, repo *repository.Repository) {
	t.Helper()
	ctx := context.Background()
	id, _ := idgen.Generator{}.PuzzleID()
	p := gen.Puzzle{
		ID:                   id.String(),
		Revision:             1,
		State:                string(shared.PuzzleActive),
		Difficulty:           string(shared.DifficultyMedium),
		HardestTechnique:     "naked_single",
		QualityScore:         100,
		MultiplayerApproved:  1,
		GeneratorVersion:     "gen-1",
		SolverVersion:        "solver-1",
		CanonicalFingerprint: "fp",
		Clues:                []byte("000000000000000000000000000000000000000000000000000000000000000000000000000000000"),
		Solution:             []byte("123456789456789123789123456214365897365897214897214365531642978642978531978531642"),
		CreatedAtMs:          time.Now().UnixMilli(),
	}
	if err := repo.CreatePuzzle(ctx, p); err != nil {
		t.Fatalf("create puzzle: %v", err)
	}
}

func createRoomAndActor(t *testing.T, repo *repository.Repository) (*Actor, *roomdomain.Room, shared.RoomCode) {
	t.Helper()
	ctx := context.Background()
	g := idgen.Generator{}
	roomID, _ := g.RoomID()
	participantID, _ := g.ParticipantID()
	name, _ := shared.NewDisplayName("Mila")
	code, _ := roomdomain.GenerateCode()
	now := time.Now()
	host := roomdomain.Participant{ID: participantID, Name: name, Role: shared.RolePlayer, JoinedAt: nowTs(t, now)}
	rules := roomdomain.MatchRules{
		Mode: shared.ModeCoop, Difficulty: shared.DifficultyMedium, ErrorPreset: shared.ErrorPresetCasual,
		HintsEnabled: true, SharedNotes: true, AutoRemoveNotes: true, SpectatorsAllowed: true, RuleVersion: 1,
	}
	room, err := roomdomain.NewRoom(roomID, code, host, rules, now)
	if err != nil {
		t.Fatalf("new room: %v", err)
	}

	tx, txRepo, err := repo.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer repository.TxRollback(tx)
	token, _ := tokenAndHash()
	genSession := gen.RoomSession{
		TokenHash:     token.Hash,
		RoomID:        roomID.String(),
		ParticipantID: participantID.String(),
		CreatedAtMs:   now.UnixMilli(),
		ExpiresAtMs:   now.Add(7 * 24 * time.Hour).UnixMilli(),
	}
	if err := txRepo.CreateRoomTx(ctx, tx, roomToGen(room), []gen.RoomParticipant{participantToGen(host, roomID)}, genSession); err != nil {
		t.Fatalf("create room: %v", err)
	}
	if err := repository.TxCommit(tx); err != nil {
		t.Fatalf("commit: %v", err)
	}

	actor := NewActor(room, nil, repo, nil)
	t.Cleanup(actor.Stop)
	return actor, room, code
}

func TestActorJoinPersistsParticipant(t *testing.T) {
	repo, db := newTestActorRepo(t)
	defer db.Close()
	actor, room, _ := createRoomAndActor(t, repo)
	ctx := context.Background()

	g := idgen.Generator{}
	newPID, _ := g.ParticipantID()
	name, _ := shared.NewDisplayName("Noah")
	token, _ := tokenAndHash()

	reqID, _ := g.RequestID()
	cmd := roomdomain.RequestJoinCommand{
		Meta: shared.CommandMetadata{
			RequestID:                  reqID,
			AuthenticatedParticipantID: newPID,
			ClientSequence:             1,
			Target:                     shared.NewRoomTarget(room.ID),
			ExpectedVersion:            1,
		},
		Code:          room.Code,
		DisplayName:   name,
		Role:          shared.RolePlayer,
		ParticipantID: newPID,
	}
	_, err := actor.Submit(ctx, Envelope{
		RequestID:      reqID,
		CommandType:    "RequestJoin",
		ScopeHash:      token.Hash,
		Fingerprint:    "noah",
		Command:        cmd,
		NewSessionHash: token.Hash,
	})
	if err != nil {
		t.Fatalf("join: %v", err)
	}
	if len(actor.room.Participants) != 2 {
		t.Fatalf("expected 2 participants, got %d", len(actor.room.Participants))
	}
	if actor.room.Version != 2 {
		t.Fatalf("expected version 2, got %d", actor.room.Version)
	}
}

func TestActorDuplicateRequestReturnsStoredResponse(t *testing.T) {
	repo, db := newTestActorRepo(t)
	defer db.Close()
	actor, room, _ := createRoomAndActor(t, repo)
	ctx := context.Background()
	g := idgen.Generator{}
	reqID, _ := g.RequestID()
	token, _ := tokenAndHash()

	cmd := roomdomain.LeaveRoomCommand{
		Meta: shared.CommandMetadata{
			RequestID:                  reqID,
			AuthenticatedParticipantID: room.Participants[0].ID,
			ClientSequence:             1,
			Target:                     shared.NewRoomTarget(room.ID),
			ExpectedVersion:            1,
		},
		Intent: "leave_lobby",
	}
	env := Envelope{
		RequestID:   reqID,
		CommandType: "LeaveRoom",
		ScopeHash:   token.Hash,
		Fingerprint: "leave",
		Command:     cmd,
	}
	_, err := actor.Submit(ctx, env)
	if err != nil {
		t.Fatalf("leave: %v", err)
	}
	result, err := actor.Submit(ctx, env)
	if err != nil {
		t.Fatalf("duplicate: %v", err)
	}
	if !result.Duplicate {
		t.Fatalf("expected duplicate result")
	}
}

func TestActorConcurrentSubmissionsDoNotLoseUpdate(t *testing.T) {
	repo, db := newTestActorRepo(t)
	defer db.Close()
	actor, room, _ := createRoomAndActor(t, repo)
	ctx := context.Background()
	g := idgen.Generator{}

	ready := func(reqID shared.RequestID, tokenHash []byte) Envelope {
		return Envelope{
			RequestID:   reqID,
			CommandType: "SetReady",
			ScopeHash:   tokenHash,
			Fingerprint: "ready",
			Command: roomdomain.SetReadyCommand{
				Meta: shared.CommandMetadata{
					RequestID:                  reqID,
					AuthenticatedParticipantID: room.Participants[0].ID,
					ClientSequence:             1,
					Target:                     shared.NewRoomTarget(room.ID),
					ExpectedVersion:            1,
				},
				Ready: true,
			},
		}
	}

	var successes int
	var failures int
	for i := 0; i < 10; i++ {
		reqID, _ := g.RequestID()
		token, _ := tokenAndHash()
		_, err := actor.Submit(ctx, ready(reqID, token.Hash))
		if err != nil {
			failures++
		} else {
			successes++
		}
	}
	if successes != 1 {
		t.Fatalf("expected exactly one success, got %d", successes)
	}
	if failures != 9 {
		t.Fatalf("expected 9 failures, got %d", failures)
	}
}

func TestActorDoesNotBroadcastOrAcknowledgeFailedCommit(t *testing.T) {
	repo, db := newTestActorRepo(t)
	actor, room, _ := createRoomAndActor(t, repo)
	ctx := context.Background()
	g := idgen.Generator{}
	connID, _ := g.ConnectionID()
	outbound := make(chan []byte, 8)
	if _, err := actor.Subscribe(ctx, connID, room.Participants[0].ID, outbound); err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	<-outbound // connection.accepted
	<-outbound // room.snapshot

	if err := db.Close(); err != nil {
		t.Fatalf("close database: %v", err)
	}
	reqID, _ := g.RequestID()
	_, err := actor.Submit(ctx, Envelope{
		RequestID:    reqID,
		CommandType:  "room.set_ready",
		ConnectionID: connID,
		Command: roomdomain.SetReadyCommand{
			Meta: shared.CommandMetadata{
				RequestID:                  reqID,
				AuthenticatedParticipantID: room.Participants[0].ID,
				ClientSequence:             1,
				Target:                     shared.NewRoomTarget(room.ID),
				ExpectedVersion:            uint64(room.Version),
			},
			Ready: true,
		},
	})
	if err == nil {
		t.Fatal("expected persistence failure")
	}
	select {
	case message := <-outbound:
		t.Fatalf("unexpected pre-commit broadcast: %s", message)
	case <-time.After(50 * time.Millisecond):
	}
	if actor.room.Participants[0].IsReady {
		t.Fatal("failed command mutated committed actor state")
	}
}

func tokenAndHash() (roomsession.Token, error) {
	return roomsession.Generate()
}

func nowTs(t *testing.T, now time.Time) shared.Timestamp {
	ts, err := shared.NewTimestamp(now)
	if err != nil {
		t.Fatalf("timestamp: %v", err)
	}
	return ts
}
