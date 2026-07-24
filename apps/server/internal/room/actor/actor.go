package actor

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	shared "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/domain"
	matchdomain "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/match/domain"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/gen"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/repository"
	roomdomain "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/room/domain"
)

const commandQueueCapacity = 256

// Envelope carries a command plus the idempotency and session context needed
// to persist and authorize it.
type Envelope struct {
	RequestID      shared.RequestID
	CommandType    string
	ScopeHash      []byte
	Fingerprint    string
	Command        roomdomain.Command
	NewSessionHash []byte
}

// Result is the outcome of a command processed by the actor.
type Result struct {
	Events    []roomdomain.Event
	Response  []byte
	Duplicate bool
}

// Actor owns one room and serializes every authoritative mutation.
type Actor struct {
	roomID   shared.RoomID
	room     *roomdomain.Room
	match    *matchdomain.Match
	repo     *repository.Repository
	logger   *slog.Logger
	cmdCh    chan actorMsg
	stopCh   chan struct{}
	stopOnce chan struct{}
	wg       sync.WaitGroup
	timers   map[uint64]*time.Timer
	timerMu  sync.Mutex
}

type actorMsg struct {
	ctx  context.Context
	env  Envelope
	resp chan actorResult
}

type actorResult struct {
	result Result
	err    error
}

// NewActor creates an actor for an already loaded room.
func NewActor(room *roomdomain.Room, match *matchdomain.Match, repo *repository.Repository, logger *slog.Logger) *Actor {
	a := &Actor{
		roomID:   room.ID,
		room:     room,
		match:    match,
		repo:     repo,
		logger:   logger,
		cmdCh:    make(chan actorMsg, commandQueueCapacity),
		stopCh:   make(chan struct{}),
		stopOnce: make(chan struct{}),
		timers:   make(map[uint64]*time.Timer),
	}
	a.wg.Add(1)
	go a.run()
	return a
}

// Stop gracefully terminates the actor run loop.
func (a *Actor) Stop() {
	close(a.stopCh)
	a.timerMu.Lock()
	for _, t := range a.timers {
		t.Stop()
	}
	a.timerMu.Unlock()
	a.wg.Wait()
}

// Submit enqueues a command and waits for the result or context cancellation.
func (a *Actor) Submit(ctx context.Context, env Envelope) (Result, error) {
	resp := make(chan actorResult, 1)
	select {
	case a.cmdCh <- actorMsg{ctx: ctx, env: env, resp: resp}:
	case <-a.stopCh:
		return Result{}, shared.DomainError{Code: shared.ErrServerBusy}
	case <-ctx.Done():
		return Result{}, ctx.Err()
	default:
		return Result{}, shared.DomainError{Code: shared.ErrServerBusy}
	}
	select {
	case r := <-resp:
		return r.result, r.err
	case <-a.stopCh:
		return Result{}, shared.DomainError{Code: shared.ErrServerBusy}
	case <-ctx.Done():
		return Result{}, ctx.Err()
	}
}

// HasActiveTimers reports whether the actor owns a pending authoritative timer.
func (a *Actor) HasActiveTimers() bool {
	a.timerMu.Lock()
	defer a.timerMu.Unlock()
	return len(a.timers) > 0
}

func (a *Actor) run() {
	defer a.wg.Done()
	for {
		select {
		case <-a.stopCh:
			return
		case msg := <-a.cmdCh:
			a.handle(msg)
		}
	}
}

func (a *Actor) handle(msg actorMsg) {
	ctx, env, resp := msg.ctx, msg.env, msg.resp
	result, err := a.process(ctx, env)
	select {
	case resp <- actorResult{result: result, err: err}:
	default:
	}
}

func (a *Actor) process(ctx context.Context, env Envelope) (Result, error) {
	now := time.Now()

	tx, txRepo, err := a.repo.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return Result{}, shared.DomainError{Code: shared.ErrPersistenceFailed}
	}
	defer repository.TxRollback(tx)

	// Idempotency check.
	receipt, err := txRepo.GetCommandReceipt(ctx, env.RequestID.String())
	if err != nil && !repository.IsNoRows(err) {
		return Result{}, shared.DomainError{Code: shared.ErrPersistenceFailed}
	}
	if err == nil {
		scopeHash := env.ScopeHash
		if scopeHash == nil {
			scopeHash = []byte{}
		}
		if !bytesEqual(receipt.AuthenticatedScopeHash, scopeHash) ||
			receipt.CommandType != env.CommandType ||
			receipt.RequestFingerprint != env.Fingerprint {
			return Result{}, shared.DomainError{Code: shared.ErrDuplicateRequest}
		}
		if !receipt.SafeResponseJson.Valid {
			return Result{Duplicate: true}, nil
		}
		return Result{Response: []byte(receipt.SafeResponseJson.String), Duplicate: true}, nil
	}

	// Apply command to a copy of the aggregate.
	working := a.cloneRoom()
	cmd := env.Command
	// Timer-fired commands are authoritative by generation; bind the current version.
	if am, ok := cmd.(roomdomain.ActivateMatchCommand); ok {
		meta := am.Meta
		meta.ExpectedVersion = uint64(working.Version)
		am.Meta = meta
		cmd = am
	}
	events, err := working.Apply(cmd, now)
	if err != nil {
		return Result{}, err
	}
	if events == nil {
		events = []roomdomain.Event{}
	}

	var matchEvents []matchdomain.Event
	var workingMatch *matchdomain.Match
	expectedMatchVersion := int64(0)

	switch typed := cmd.(type) {
	case roomdomain.StartCountdownCommand:
		// Create the prepared match and its initial events.
		participantIDs := a.playerIDs(working)
		workingMatch, matchEvents, err = matchdomain.NewPrepared(typed.MatchID, working.ID, matchRulesFromRoom(working.Rules), typed.Puzzle, participantIDs, now)
		if err != nil {
			return Result{}, err
		}
	case roomdomain.ActivateMatchCommand:
		workingMatch = a.cloneMatch()
		expectedMatchVersion = int64(workingMatch.Version)
		matchEvents, err = workingMatch.Activate(now)
		if err != nil {
			return Result{}, err
		}
	}

	// Persist projections.
	expectedVersion := int64(a.room.Version)
	participants := make([]gen.RoomParticipant, 0, len(working.Participants))
	for _, p := range working.Participants {
		participants = append(participants, participantToGen(p, working.ID))
	}
	if err := txRepo.UpdateRoomTx(ctx, tx, roomToGen(working), expectedVersion, participants, nil); err != nil {
		if errors.Is(err, repository.ErrConflict) {
			return Result{}, shared.DomainError{Code: shared.ErrStaleVersion}
		}
		return Result{}, shared.DomainError{Code: shared.ErrPersistenceFailed}
	}

	if env.NewSessionHash != nil {
		participantID := env.Command.Metadata().AuthenticatedParticipantID
		expiresAt := now.Add(7 * 24 * time.Hour).UnixMilli()
		if err := txRepo.CreateRoomSessionTx(ctx, tx, gen.RoomSession{
			TokenHash:     env.NewSessionHash,
			RoomID:        working.ID.String(),
			ParticipantID: participantID.String(),
			CreatedAtMs:   now.UnixMilli(),
			ExpiresAtMs:   expiresAt,
		}); err != nil {
			return Result{}, shared.DomainError{Code: shared.ErrPersistenceFailed}
		}
	}

	if workingMatch != nil {
		switch env.Command.(type) {
		case roomdomain.StartCountdownCommand:
			matchParticipants := make([]gen.MatchParticipant, 0, len(workingMatch.Participants))
			for _, pid := range workingMatch.Participants {
				matchParticipants = append(matchParticipants, gen.MatchParticipant{
					MatchID:       workingMatch.ID.String(),
					ParticipantID: pid.String(),
					Connected:     1,
				})
			}
			matchEventsGen := matchEventsToGen(env.RequestID, workingMatch, matchEvents)
			if err := txRepo.CreateMatchTx(ctx, tx, matchToGen(workingMatch), matchParticipants, matchEventsGen, nil); err != nil {
				return Result{}, shared.DomainError{Code: shared.ErrPersistenceFailed}
			}
		case roomdomain.ActivateMatchCommand:
			matchEventsGen := matchEventsToGen(env.RequestID, workingMatch, matchEvents)
			if err := txRepo.UpdateMatchTx(ctx, tx, matchToGen(workingMatch), expectedMatchVersion, nil, matchEventsGen, nil, nil); err != nil {
				if errors.Is(err, repository.ErrConflict) {
					return Result{}, shared.DomainError{Code: shared.ErrStaleVersion}
				}
				return Result{}, shared.DomainError{Code: shared.ErrPersistenceFailed}
			}
		}
	}

	// Revoke the session for a leaving participant.
	if _, ok := env.Command.(roomdomain.LeaveRoomCommand); ok {
		participantID := env.Command.Metadata().AuthenticatedParticipantID
		// Best-effort revocation; if the session row is missing the leave still commits.
		_ = txRepo.RevokeRoomSession(ctx, tx, env.ScopeHash, now.UnixMilli())
		_ = participantID
	}

	response := a.buildResponse(working, workingMatch, participantIDFromCommand(env.Command))
	if err := a.saveReceipt(ctx, txRepo, tx, env, "ok", response, now); err != nil {
		return Result{}, shared.DomainError{Code: shared.ErrPersistenceFailed}
	}

	if err := repository.TxCommit(tx); err != nil {
		return Result{}, shared.DomainError{Code: shared.ErrPersistenceFailed}
	}

	// Commit succeeded: replace authoritative in-memory state.
	a.room = working
	if workingMatch != nil {
		a.match = workingMatch
	}
	a.afterCommit(env.Command, working, now)

	return Result{Events: events, Response: response}, nil
}

func (a *Actor) saveReceipt(ctx context.Context, txRepo *repository.Repository, tx *sql.Tx, env Envelope, status string, response []byte, now time.Time) error {
	scopeHash := env.ScopeHash
	if scopeHash == nil {
		scopeHash = []byte{}
	}
	expiresAt := now.Add(24 * time.Hour).UnixMilli()
	receipt := gen.CommandReceipt{
		RequestID:              env.RequestID.String(),
		AuthenticatedScopeHash: scopeHash,
		CommandType:            env.CommandType,
		RequestFingerprint:     env.Fingerprint,
		TerminalStatus:         status,
		SafeResponseJson:       sql.NullString{String: string(response), Valid: true},
		CreatedAtMs:            now.UnixMilli(),
		ExpiresAtMs:            expiresAt,
	}
	return txRepo.CreateCommandReceipt(ctx, tx, receipt)
}

func (a *Actor) cloneRoom() *roomdomain.Room {
	participants := make([]roomdomain.Participant, len(a.room.Participants))
	copy(participants, a.room.Participants)
	return &roomdomain.Room{
		ID:                a.room.ID,
		Code:              a.room.Code,
		Version:           a.room.Version,
		State:             a.room.State,
		Participants:      participants,
		Rules:             a.room.Rules,
		HostParticipantID: a.room.HostParticipantID,
		CurrentMatchID:    a.room.CurrentMatchID,
		CreatedAt:         a.room.CreatedAt,
		LastActivityAt:    a.room.LastActivityAt,
		ExpiresAt:         a.room.ExpiresAt,
		Countdown:         a.room.Countdown,
	}
}

func (a *Actor) cloneMatch() *matchdomain.Match {
	if a.match == nil {
		return nil
	}
	participants := make([]shared.ParticipantID, len(a.match.Participants))
	copy(participants, a.match.Participants)
	m := *a.match
	m.Participants = participants
	return &m
}

func (a *Actor) playerIDs(room *roomdomain.Room) []shared.ParticipantID {
	var ids []shared.ParticipantID
	for _, p := range room.Participants {
		if p.IsActive() && p.Role == shared.RolePlayer {
			ids = append(ids, p.ID)
		}
	}
	return ids
}

func (a *Actor) afterCommit(cmd roomdomain.Command, room *roomdomain.Room, now time.Time) {
	switch c := cmd.(type) {
	case roomdomain.StartCountdownCommand:
		if room.Countdown != nil {
			a.startTimer(room.Countdown.Generation, room.Countdown.DeadlineAt.Time())
		}
	case roomdomain.CancelCountdownCommand:
		a.clearTimer(1)
	case roomdomain.ActivateMatchCommand:
		a.clearTimer(c.Generation)
	}
}

func (a *Actor) startTimer(generation uint64, deadline time.Time) {
	a.timerMu.Lock()
	defer a.timerMu.Unlock()
	if t, ok := a.timers[generation]; ok {
		t.Stop()
	}
	delay := time.Until(deadline)
	if delay < 0 {
		delay = 0
	}
	a.timers[generation] = time.AfterFunc(delay, func() {
		_, _ = a.Submit(context.Background(), Envelope{
			RequestID:   shared.RequestID(generateRequestID()),
			CommandType: "ActivateMatch",
			Command: roomdomain.ActivateMatchCommand{
				Meta: shared.CommandMetadata{
					ExpectedVersion: uint64(a.room.Version),
				},
				Generation: generation,
			},
		})
	})
}

func (a *Actor) clearTimer(generation uint64) {
	a.timerMu.Lock()
	defer a.timerMu.Unlock()
	if t, ok := a.timers[generation]; ok {
		t.Stop()
		delete(a.timers, generation)
	}
}

func (a *Actor) buildResponse(room *roomdomain.Room, match *matchdomain.Match, participantID shared.ParticipantID) []byte {
	view := map[string]any{
		"room": map[string]any{
			"id":                room.ID.String(),
			"code":              room.Code.String(),
			"state":             string(room.State),
			"version":           uint64(room.Version),
			"mode":              string(room.Rules.Mode),
			"difficulty":        string(room.Rules.Difficulty),
			"errorPreset":       string(room.Rules.ErrorPreset),
			"hintsEnabled":      room.Rules.HintsEnabled,
			"sharedNotes":       room.Rules.SharedNotes,
			"autoRemoveNotes":   room.Rules.AutoRemoveNotes,
			"spectatorsAllowed": room.Rules.SpectatorsAllowed,
			"hostId":            nullParticipantID(room.HostParticipantID),
			"currentMatchId":    nullMatchID(room.CurrentMatchID),
		},
		"participants": participantsToView(room.Participants),
		"self": map[string]any{
			"participantId": participantID.String(),
		},
	}
	if room.Countdown != nil {
		view["countdown"] = map[string]any{
			"matchId":    room.Countdown.MatchID.String(),
			"generation": room.Countdown.Generation,
			"deadlineAt": room.Countdown.DeadlineAt.Milliseconds(),
		}
	}
	if match != nil && room.State == shared.RoomInMatch {
		view["match"] = map[string]any{
			"id":      match.ID.String(),
			"state":   string(match.State),
			"version": uint64(match.Version),
		}
	}
	b, _ := json.Marshal(view)
	return b
}

func participantsToView(parts []roomdomain.Participant) []map[string]any {
	out := make([]map[string]any, 0, len(parts))
	for _, p := range parts {
		if !p.IsActive() {
			continue
		}
		out = append(out, map[string]any{
			"id":       p.ID.String(),
			"name":     p.Name.String(),
			"role":     string(p.Role),
			"isHost":   p.IsHost,
			"isReady":  p.IsReady,
			"joinedAt": p.JoinedAt.Milliseconds(),
		})
	}
	return out
}

func participantIDFromCommand(cmd roomdomain.Command) shared.ParticipantID {
	return cmd.Metadata().AuthenticatedParticipantID
}

func matchRulesFromRoom(rules roomdomain.MatchRules) matchdomain.Rules {
	return matchdomain.Rules{
		Mode:            rules.Mode,
		Difficulty:      rules.Difficulty,
		ErrorPreset:     rules.ErrorPreset,
		HintsEnabled:    rules.HintsEnabled,
		AutoRemoveNotes: rules.AutoRemoveNotes,
		RuleVersion:     rules.RuleVersion,
	}
}

func matchEventsToGen(requestID shared.RequestID, m *matchdomain.Match, events []matchdomain.Event) []gen.MatchEvent {
	out := make([]gen.MatchEvent, 0, len(events))
	for _, e := range events {
		meta := e.Metadata()
		out = append(out, gen.MatchEvent{
			MatchID:           m.ID.String(),
			EventNumber:       int64(meta.EventNumber),
			AggregateVersion:  int64(meta.AggregateVersion),
			PublicEventType:   eventTypeName(e),
			RequestID:         requestID.String(),
			OccurredAtMs:      meta.OccurredAt.Milliseconds(),
			PublicPayloadJson: "{}",
			PreviousHash:      nil,
			EventHash:         nil,
		})
	}
	return out
}

func eventTypeName(e matchdomain.Event) string {
	switch e.(type) {
	case matchdomain.MatchPreparedEvent:
		return "MatchPrepared"
	case matchdomain.MatchCountdownStartedEvent:
		return "MatchCountdownStarted"
	case matchdomain.MatchStartedEvent:
		return "MatchStarted"
	default:
		return "Unknown"
	}
}

func nullParticipantID(id *shared.ParticipantID) any {
	if id == nil {
		return nil
	}
	return id.String()
}

func nullMatchID(id *shared.MatchID) any {
	if id == nil {
		return nil
	}
	return id.String()
}

func bytesEqual(a, b []byte) bool {
	return sha256.Sum256(a) == sha256.Sum256(b)
}

func generateRequestID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf("019%s", hexEncode(b[:8]))
}

func hexEncode(b []byte) string {
	const hex = "0123456789abcdef"
	out := make([]byte, len(b)*2)
	for i, v := range b {
		out[i*2] = hex[v>>4]
		out[i*2+1] = hex[v&0x0f]
	}
	return string(out)
}
