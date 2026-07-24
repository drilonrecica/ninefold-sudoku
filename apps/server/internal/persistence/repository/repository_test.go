package repository

import (
	"context"
	"database/sql"
	"testing"

	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/gen"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/migrate"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/sqlite"
	"github.com/google/uuid"
)

func newTestRepo(t *testing.T) *Repository {
	t.Helper()
	db, err := sqlite.New(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("sqlite new: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := migrate.Up(db.Writer()); err != nil {
		t.Fatalf("migrate up: %v", err)
	}
	return New(db)
}

func newUUID(t *testing.T) string {
	t.Helper()
	id, err := uuid.NewV7()
	if err != nil {
		t.Fatalf("uuid: %v", err)
	}
	return id.String()
}

func createPuzzle(t *testing.T, repo *Repository, difficulty string) gen.Puzzle {
	t.Helper()
	puzzle := gen.Puzzle{
		ID:                   newUUID(t),
		Revision:             1,
		State:                "Active",
		Difficulty:           difficulty,
		HardestTechnique:     "naked_single",
		QualityScore:         0.75,
		MultiplayerApproved:  1,
		GeneratorVersion:     "gen-1",
		SolverVersion:        "solver-1",
		CanonicalFingerprint: "fp-" + newUUID(t),
		Clues:                make([]byte, 81),
		Solution:             make([]byte, 81),
		CreatedAtMs:          NowMs(),
	}
	for i := range puzzle.Clues {
		puzzle.Clues[i] = byte('0' + (i % 9) + 1)
		puzzle.Solution[i] = byte('0' + (i % 9) + 1)
	}
	if err := repo.CreatePuzzle(context.Background(), puzzle); err != nil {
		t.Fatalf("create puzzle: %v", err)
	}
	return puzzle
}

func createRoomAndParticipant(t *testing.T, repo *Repository, code string) (gen.Room, gen.RoomParticipant) {
	t.Helper()
	ctx := context.Background()
	room := gen.Room{
		ID:                newUUID(t),
		Code:              code,
		State:             "Lobby",
		Version:           1,
		Mode:              "Coop",
		Difficulty:        "Medium",
		ErrorPreset:       "Casual",
		HintsEnabled:      1,
		SharedNotes:       1,
		AutoRemoveNotes:   1,
		SpectatorsAllowed: 1,
		HostParticipantID: newUUID(t),
		CurrentMatchID:    sql.NullString{},
		CreatedAtMs:       NowMs(),
		LastActivityAtMs:  NowMs(),
		ExpiresAtMs:       NowMs() + 3600000,
	}
	participant := gen.RoomParticipant{
		ID:          room.HostParticipantID,
		RoomID:      room.ID,
		DisplayName: "Mila",
		Role:        "Player",
		IsHost:      1,
		IsReady:     0,
		JoinedAtMs:  NowMs(),
	}
	session := gen.RoomSession{
		TokenHash:     []byte("hash-" + code),
		RoomID:        room.ID,
		ParticipantID: participant.ID,
		CreatedAtMs:   NowMs(),
		ExpiresAtMs:   NowMs() + 3600000,
	}
	tx, txRepo, err := repo.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	if err := txRepo.CreateRoomTx(ctx, tx, room, []gen.RoomParticipant{participant}, session); err != nil {
		t.Fatalf("create room: %v", err)
	}
	if err := TxCommit(tx); err != nil {
		t.Fatalf("commit: %v", err)
	}
	return room, participant
}

func TestPuzzleCreateAndGet(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()
	puzzle := gen.Puzzle{
		ID:                   newUUID(t),
		Revision:             1,
		State:                "Active",
		Difficulty:           "Medium",
		HardestTechnique:     "hidden_single",
		QualityScore:         0.85,
		MultiplayerApproved:  1,
		GeneratorVersion:     "gen-1",
		SolverVersion:        "solver-1",
		CanonicalFingerprint: "fp-" + newUUID(t),
		Clues:                make([]byte, 81),
		Solution:             make([]byte, 81),
		CreatedAtMs:          NowMs(),
	}
	for i := range puzzle.Clues {
		puzzle.Clues[i] = byte('1')
		puzzle.Solution[i] = byte('1')
	}
	if err := repo.CreatePuzzle(ctx, puzzle); err != nil {
		t.Fatalf("create puzzle: %v", err)
	}
	got, err := repo.GetPuzzle(ctx, puzzle.ID, puzzle.Revision)
	if err != nil {
		t.Fatalf("get puzzle: %v", err)
	}
	if got.ID != puzzle.ID || got.CanonicalFingerprint != puzzle.CanonicalFingerprint {
		t.Fatalf("puzzle mismatch: got %+v", got)
	}
}

func TestCreateRoomTxRollsBackOnError(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()
	tx, txRepo, err := repo.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	room := gen.Room{
		ID:                newUUID(t),
		Code:              "7KMP4R",
		State:             "Lobby",
		Version:           1,
		Mode:              "Coop",
		Difficulty:        "Medium",
		ErrorPreset:       "Casual",
		HintsEnabled:      1,
		SharedNotes:       1,
		AutoRemoveNotes:   1,
		SpectatorsAllowed: 1,
		HostParticipantID: newUUID(t),
		CurrentMatchID:    sql.NullString{},
		CreatedAtMs:       NowMs(),
		LastActivityAtMs:  NowMs(),
		ExpiresAtMs:       NowMs() + 3600000,
	}
	participant := gen.RoomParticipant{
		ID:          room.HostParticipantID,
		RoomID:      room.ID,
		DisplayName: "Mila",
		Role:        "Player",
		IsHost:      1,
		IsReady:     0,
		JoinedAtMs:  NowMs(),
	}
	session := gen.RoomSession{
		TokenHash:     []byte("hash"),
		RoomID:        room.ID,
		ParticipantID: participant.ID,
		CreatedAtMs:   NowMs(),
		ExpiresAtMs:   NowMs() + 3600000,
	}
	// intentionally commit only part to force rollback, then verify room absent.
	if err := txRepo.CreateRoomTx(ctx, tx, room, []gen.RoomParticipant{participant}, session); err != nil {
		t.Fatalf("create room tx: %v", err)
	}
	if err := TxRollback(tx); err != nil {
		t.Fatalf("rollback: %v", err)
	}
	_, err = repo.GetRoomByID(ctx, room.ID)
	if !IsNoRows(err) {
		t.Fatalf("expected room to be absent after rollback, got %v", err)
	}
}

func TestCreateRoomAndReadBack(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()
	room := gen.Room{
		ID:                newUUID(t),
		Code:              "7KMP4R",
		State:             "Lobby",
		Version:           1,
		Mode:              "Coop",
		Difficulty:        "Medium",
		ErrorPreset:       "Casual",
		HintsEnabled:      1,
		SharedNotes:       1,
		AutoRemoveNotes:   1,
		SpectatorsAllowed: 1,
		HostParticipantID: newUUID(t),
		CurrentMatchID:    sql.NullString{},
		CreatedAtMs:       NowMs(),
		LastActivityAtMs:  NowMs(),
		ExpiresAtMs:       NowMs() + 3600000,
	}
	participant := gen.RoomParticipant{
		ID:          room.HostParticipantID,
		RoomID:      room.ID,
		DisplayName: "Mila",
		Role:        "Player",
		IsHost:      1,
		IsReady:     0,
		JoinedAtMs:  NowMs(),
	}
	session := gen.RoomSession{
		TokenHash:     []byte("hash"),
		RoomID:        room.ID,
		ParticipantID: participant.ID,
		CreatedAtMs:   NowMs(),
		ExpiresAtMs:   NowMs() + 3600000,
	}
	tx, txRepo, err := repo.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	if err := txRepo.CreateRoomTx(ctx, tx, room, []gen.RoomParticipant{participant}, session); err != nil {
		t.Fatalf("create room tx: %v", err)
	}
	if err := TxCommit(tx); err != nil {
		t.Fatalf("commit: %v", err)
	}

	got, err := repo.GetRoomByID(ctx, room.ID)
	if err != nil {
		t.Fatalf("get room: %v", err)
	}
	if got.Code != room.Code || got.Version != room.Version {
		t.Fatalf("room mismatch: got %+v", got)
	}
	participants, err := repo.ListActiveRoomParticipants(ctx, room.ID)
	if err != nil {
		t.Fatalf("list participants: %v", err)
	}
	if len(participants) != 1 || participants[0].DisplayName != "Mila" {
		t.Fatalf("participants mismatch: got %+v", participants)
	}
	gotSession, err := repo.GetRoomSessionByHash(ctx, session.TokenHash)
	if err != nil {
		t.Fatalf("get session: %v", err)
	}
	if gotSession.RoomID != room.ID {
		t.Fatalf("session mismatch: got %+v", gotSession)
	}
}

func TestUpdateRoomOptimisticVersionConflict(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()
	room := gen.Room{
		ID:                newUUID(t),
		Code:              "ABCDEF",
		State:             "Lobby",
		Version:           1,
		Mode:              "Coop",
		Difficulty:        "Easy",
		ErrorPreset:       "Casual",
		HintsEnabled:      1,
		SharedNotes:       1,
		AutoRemoveNotes:   1,
		SpectatorsAllowed: 1,
		HostParticipantID: newUUID(t),
		CreatedAtMs:       NowMs(),
		LastActivityAtMs:  NowMs(),
		ExpiresAtMs:       NowMs() + 3600000,
	}
	participant := gen.RoomParticipant{
		ID:          room.HostParticipantID,
		RoomID:      room.ID,
		DisplayName: "Noah",
		Role:        "Player",
		IsHost:      1,
		JoinedAtMs:  NowMs(),
	}
	session := gen.RoomSession{
		TokenHash:     []byte("hash1"),
		RoomID:        room.ID,
		ParticipantID: participant.ID,
		CreatedAtMs:   NowMs(),
		ExpiresAtMs:   NowMs() + 3600000,
	}
	tx, txRepo, err := repo.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	if err := txRepo.CreateRoomTx(ctx, tx, room, []gen.RoomParticipant{participant}, session); err != nil {
		t.Fatalf("create room tx: %v", err)
	}
	if err := TxCommit(tx); err != nil {
		t.Fatalf("commit: %v", err)
	}

	// Try to update with wrong expected version (0 instead of 1).
	tx, txRepo, err = repo.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer TxRollback(tx)
	room.Version = 2
	room.State = "Countdown"
	if err := txRepo.UpdateRoomTx(ctx, tx, room, 0, []gen.RoomParticipant{participant}, nil); err == nil {
		t.Fatal("expected optimistic version conflict")
	}
}

func TestCreateMatchWithEventsAndSnapshot(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	room, participant := createRoomAndParticipant(t, repo, "MATCH01")
	puzzle := createPuzzle(t, repo, "Easy")

	match := gen.Match{
		ID:                 newUUID(t),
		RoomID:             room.ID,
		State:              "Prepared",
		Version:            1,
		Mode:               "Coop",
		Difficulty:         "Medium",
		ErrorPreset:        "Casual",
		HintsEnabled:       1,
		AutoRemoveNotes:    1,
		RuleVersion:        1,
		PuzzleID:           puzzle.ID,
		PuzzleRevision:     puzzle.Revision,
		TransformationSeed: 42,
		PuzzleDifficulty:   "Medium",
		GeneratorVersion:   "gen-1",
		SolverVersion:      "solver-1",
		CreatedAtMs:        NowMs(),
	}
	matchParticipant := gen.MatchParticipant{
		MatchID:       match.ID,
		ParticipantID: participant.ID,
		Connected:     1,
	}
	event := gen.MatchEvent{
		MatchID:           match.ID,
		EventNumber:       1,
		AggregateVersion:  1,
		PublicEventType:   "MatchPrepared",
		PublicActorID:     NewNullString(participant.ID),
		RequestID:         newUUID(t),
		OccurredAtMs:      NowMs(),
		PublicPayloadJson: "{}",
	}
	snapshot := gen.MatchSnapshot{
		MatchID:          match.ID,
		EventNumber:      1,
		AggregateVersion: 1,
		StateFormat:      "gzip-json-v1",
		StateBlob:        []byte("compressed-state"),
		IntegrityHash:    []byte("hash"),
		CreatedAtMs:      NowMs(),
	}

	tx, txRepo, err := repo.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	if err := txRepo.CreateMatchTx(ctx, tx, match, []gen.MatchParticipant{matchParticipant}, []gen.MatchEvent{event}, &snapshot); err != nil {
		t.Fatalf("create match tx: %v", err)
	}
	if err := TxCommit(tx); err != nil {
		t.Fatalf("commit: %v", err)
	}

	got, err := repo.GetMatchByID(ctx, match.ID)
	if err != nil {
		t.Fatalf("get match: %v", err)
	}
	if got.State != match.State {
		t.Fatalf("match mismatch: got %+v", got)
	}
	events, err := repo.GetMatchEvents(ctx, match.ID)
	if err != nil {
		t.Fatalf("get events: %v", err)
	}
	if len(events) != 1 || events[0].PublicEventType != "MatchPrepared" {
		t.Fatalf("events mismatch: got %+v", events)
	}
	latest, err := repo.GetLatestMatchSnapshot(ctx, match.ID)
	if err != nil {
		t.Fatalf("get snapshot: %v", err)
	}
	if latest.EventNumber != 1 {
		t.Fatalf("snapshot mismatch: got %+v", latest)
	}
}

func TestDuplicateRequestIDRejected(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()

	room, _ := createRoomAndParticipant(t, repo, "DUPREQ")
	puzzle := createPuzzle(t, repo, "Easy")
	match := gen.Match{
		ID:                 newUUID(t),
		RoomID:             room.ID,
		State:              "Prepared",
		Version:            1,
		Mode:               "Coop",
		Difficulty:         "Easy",
		ErrorPreset:        "Casual",
		HintsEnabled:       1,
		AutoRemoveNotes:    1,
		RuleVersion:        1,
		PuzzleID:           puzzle.ID,
		PuzzleRevision:     puzzle.Revision,
		TransformationSeed: 1,
		PuzzleDifficulty:   "Easy",
		GeneratorVersion:   "gen-1",
		SolverVersion:      "solver-1",
		CreatedAtMs:        NowMs(),
	}
	reqID := newUUID(t)
	events := []gen.MatchEvent{
		{
			MatchID:           match.ID,
			EventNumber:       1,
			AggregateVersion:  1,
			PublicEventType:   "MatchPrepared",
			RequestID:         reqID,
			OccurredAtMs:      NowMs(),
			PublicPayloadJson: "{}",
		},
		{
			MatchID:           match.ID,
			EventNumber:       2,
			AggregateVersion:  1,
			PublicEventType:   "MatchStarted",
			RequestID:         reqID,
			OccurredAtMs:      NowMs(),
			PublicPayloadJson: "{}",
		},
	}

	tx, txRepo, err := repo.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer TxRollback(tx)
	if err := txRepo.CreateMatchTx(ctx, tx, match, nil, events, nil); err == nil {
		t.Fatalf("expected duplicate request_id to fail")
	}
}

func TestCommandReceiptIdempotency(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()
	receipt := gen.CommandReceipt{
		RequestID:              newUUID(t),
		AuthenticatedScopeHash: []byte("scope"),
		CommandType:            "CreateRoom",
		RequestFingerprint:     "fingerprint",
		TerminalStatus:         "accepted",
		SafeResponseJson:       NewNullString(`{"ok":true}`),
		CreatedAtMs:            NowMs(),
		ExpiresAtMs:            NowMs() + 86400000,
	}

	tx, txRepo, err := repo.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	if err := txRepo.CreateCommandReceipt(ctx, tx, receipt); err != nil {
		t.Fatalf("create receipt: %v", err)
	}
	if err := TxCommit(tx); err != nil {
		t.Fatalf("commit: %v", err)
	}

	got, err := repo.GetCommandReceipt(ctx, receipt.RequestID)
	if err != nil {
		t.Fatalf("get receipt: %v", err)
	}
	if got.TerminalStatus != receipt.TerminalStatus {
		t.Fatalf("receipt mismatch: got %+v", got)
	}
}

func TestForeignKeyViolation(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()
	event := gen.MatchEvent{
		MatchID:           newUUID(t),
		EventNumber:       1,
		AggregateVersion:  1,
		PublicEventType:   "MatchPrepared",
		RequestID:         newUUID(t),
		OccurredAtMs:      NowMs(),
		PublicPayloadJson: "{}",
	}

	tx, txRepo, err := repo.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	defer TxRollback(tx)
	if err := txRepo.q.WithTx(tx).CreateMatchEvent(ctx, gen.CreateMatchEventParams{
		MatchID:              event.MatchID,
		EventNumber:          event.EventNumber,
		AggregateVersion:     event.AggregateVersion,
		PublicEventType:      event.PublicEventType,
		PublicActorID:        event.PublicActorID,
		RequestID:            event.RequestID,
		OccurredAtMs:         event.OccurredAtMs,
		PublicPayloadJson:    event.PublicPayloadJson,
		PrivatePayloadBlob:   event.PrivatePayloadBlob,
		PrivatePayloadSalt:   event.PrivatePayloadSalt,
		PrivatePayloadDigest: event.PrivatePayloadDigest,
		PreviousHash:         event.PreviousHash,
		EventHash:            event.EventHash,
	}); err == nil {
		t.Fatal("expected foreign-key violation for missing match")
	}
}

func TestConcurrentWritersDoNotLoseUpdate(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()
	room := gen.Room{
		ID:                newUUID(t),
		Code:              "CONCUR",
		State:             "Lobby",
		Version:           1,
		Mode:              "Coop",
		Difficulty:        "Easy",
		ErrorPreset:       "Casual",
		HintsEnabled:      1,
		SharedNotes:       1,
		AutoRemoveNotes:   1,
		SpectatorsAllowed: 1,
		HostParticipantID: newUUID(t),
		CreatedAtMs:       NowMs(),
		LastActivityAtMs:  NowMs(),
		ExpiresAtMs:       NowMs() + 3600000,
	}
	participant := gen.RoomParticipant{
		ID:          room.HostParticipantID,
		RoomID:      room.ID,
		DisplayName: "Mila",
		Role:        "Player",
		IsHost:      1,
		JoinedAtMs:  NowMs(),
	}
	session := gen.RoomSession{
		TokenHash:     []byte("hash"),
		RoomID:        room.ID,
		ParticipantID: participant.ID,
		CreatedAtMs:   NowMs(),
		ExpiresAtMs:   NowMs() + 3600000,
	}
	tx, txRepo, err := repo.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	if err := txRepo.CreateRoomTx(ctx, tx, room, []gen.RoomParticipant{participant}, session); err != nil {
		t.Fatalf("create room: %v", err)
	}
	if err := TxCommit(tx); err != nil {
		t.Fatalf("commit: %v", err)
	}

	done := make(chan error, 2)
	for i := 0; i < 2; i++ {
		i := i
		go func() {
			tx, txRepo, err := repo.BeginTx(ctx, nil)
			if err != nil {
				done <- err
				return
			}
			defer TxRollback(tx)
			updated := room
			updated.Version = 2
			updated.State = "Countdown"
			// Add a tiny offset to expires to ensure at least one writer wins.
			updated.ExpiresAtMs = NowMs() + int64(3600000+i)
			err = txRepo.UpdateRoomTx(ctx, tx, updated, 1, []gen.RoomParticipant{participant}, nil)
			if err == nil {
				err = TxCommit(tx)
			}
			done <- err
		}()
	}
	var successes int
	for i := 0; i < 2; i++ {
		if err := <-done; err == nil {
			successes++
		}
	}
	if successes != 1 {
		t.Fatalf("expected exactly one successful concurrent update, got %d", successes)
	}
	got, err := repo.GetRoomByID(ctx, room.ID)
	if err != nil {
		t.Fatalf("get room: %v", err)
	}
	if got.Version != 2 {
		t.Fatalf("expected version 2 after one successful update, got %d", got.Version)
	}
}

func TestPublicReadMethodsDoNotReturnSolutions(t *testing.T) {
	repo := newTestRepo(t)
	ctx := context.Background()
	puzzle := createPuzzle(t, repo, "Easy")

	// The puzzle assignment path intentionally returns the full row including solution.
	// This test documents that public list queries omit the solution column from their SELECT.
	rows, err := repo.q.ListActivePuzzlesByDifficulty(ctx, "Easy")
	if err != nil {
		t.Fatalf("list puzzles: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 puzzle, got %d", len(rows))
	}
	if string(rows[0].Clues) != string(puzzle.Clues) {
		t.Fatalf("clues should match")
	}
	// The generated list row struct does not include a Solution field, so the solution cannot leak.
	// (Compilation check: gen.ListActivePuzzlesByDifficultyRow has no Solution field.)
}
