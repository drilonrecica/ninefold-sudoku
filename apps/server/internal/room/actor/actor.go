package actor

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"

	shared "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/domain"
	matchdomain "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/match/domain"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/gen"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/repository"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/platform/idgen"
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
	Command        any
	NewSessionHash []byte
	ConnectionID   shared.ConnectionID
}

// Result is the outcome of a command processed by the actor.
type Result struct {
	Events           []roomdomain.Event
	Response         []byte
	OriginBroadcasts [][]byte
	Duplicate        bool
}

// ControlCommand is a non-mutating transport command that carries metadata only.
type ControlCommand struct {
	Meta shared.CommandMetadata
}

func (c ControlCommand) Metadata() shared.CommandMetadata { return c.Meta }

// Actor owns one room and serializes every authoritative mutation.
type Actor struct {
	roomID          shared.RoomID
	room            *roomdomain.Room
	match           *matchdomain.Match
	repo            *repository.Repository
	logger          *slog.Logger
	cmdCh           chan actorMsg
	stopCh          chan struct{}
	stopOnce        chan struct{}
	wg              sync.WaitGroup
	timers          map[uint64]*time.Timer
	timerMu         sync.Mutex
	lastEventNumber uint64
	lastEventHash   []byte
	subscribers     map[shared.ConnectionID]subscriber
	subMu           sync.RWMutex
	controllers     map[shared.ParticipantID]controllerInfo
	controllersMu   sync.RWMutex
}

type subscriber struct {
	participantID shared.ParticipantID
	sendCh        chan []byte
	disconnect    func()
}

type controllerInfo struct {
	connectionID shared.ConnectionID
	generation   uint64
	lastSequence shared.ClientSequence
}

type actorMsg struct {
	ctx     context.Context
	env     *Envelope
	control *controlReq
	resp    chan actorResult
}

type controlReq struct {
	kind          string
	connID        shared.ConnectionID
	participantID shared.ParticipantID
	sendCh        chan []byte
	disconnect    func()
}

type actorResult struct {
	result Result
	err    error
}

// NewActor creates an actor for an already loaded room.
func NewActor(room *roomdomain.Room, match *matchdomain.Match, repo *repository.Repository, logger *slog.Logger) *Actor {
	a := &Actor{
		roomID:      room.ID,
		room:        room,
		match:       match,
		repo:        repo,
		logger:      logger,
		cmdCh:       make(chan actorMsg, commandQueueCapacity),
		stopCh:      make(chan struct{}),
		stopOnce:    make(chan struct{}),
		timers:      make(map[uint64]*time.Timer),
		subscribers: make(map[shared.ConnectionID]subscriber),
		controllers: make(map[shared.ParticipantID]controllerInfo),
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

// Subscribe registers a WebSocket connection for broadcasts and returns the
// initial connection.accepted message plus snapshots. It runs on the actor goroutine.
func (a *Actor) Subscribe(ctx context.Context, connID shared.ConnectionID, participantID shared.ParticipantID, sendCh chan []byte, disconnect ...func()) ([]byte, error) {
	var disconnectSlowClient func()
	if len(disconnect) > 0 {
		disconnectSlowClient = disconnect[0]
	}
	resp := make(chan actorResult, 1)
	select {
	case a.cmdCh <- actorMsg{ctx: ctx, control: &controlReq{kind: "subscribe", connID: connID, participantID: participantID, sendCh: sendCh, disconnect: disconnectSlowClient}, resp: resp}:
	case <-a.stopCh:
		return nil, shared.DomainError{Code: shared.ErrServerBusy}
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	select {
	case r := <-resp:
		return r.result.Response, r.err
	case <-a.stopCh:
		return nil, shared.DomainError{Code: shared.ErrServerBusy}
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Sync sends the current controller state and authoritative snapshots directly to the
// supplied channel. It is safe for reconnect and explicit resynchronization.
func (a *Actor) Sync(ctx context.Context, connID shared.ConnectionID, participantID shared.ParticipantID, sendCh chan []byte) error {
	resp := make(chan actorResult, 1)
	select {
	case a.cmdCh <- actorMsg{ctx: ctx, control: &controlReq{kind: "sync", connID: connID, participantID: participantID, sendCh: sendCh}, resp: resp}:
	case <-a.stopCh:
		return shared.DomainError{Code: shared.ErrServerBusy}
	case <-ctx.Done():
		return ctx.Err()
	}
	select {
	case r := <-resp:
		return r.err
	case <-a.stopCh:
		return shared.DomainError{Code: shared.ErrServerBusy}
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Unsubscribe removes a WebSocket connection from broadcasts.
func (a *Actor) Unsubscribe(ctx context.Context, connID shared.ConnectionID) error {
	resp := make(chan actorResult, 1)
	select {
	case a.cmdCh <- actorMsg{ctx: ctx, control: &controlReq{kind: "unsubscribe", connID: connID}, resp: resp}:
	case <-a.stopCh:
		return shared.DomainError{Code: shared.ErrServerBusy}
	case <-ctx.Done():
		return ctx.Err()
	}
	select {
	case r := <-resp:
		return r.err
	case <-a.stopCh:
		return shared.DomainError{Code: shared.ErrServerBusy}
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Submit enqueues a command and waits for the result or context cancellation.
func (a *Actor) Submit(ctx context.Context, env Envelope) (Result, error) {
	resp := make(chan actorResult, 1)
	select {
	case a.cmdCh <- actorMsg{ctx: ctx, env: &env, resp: resp}:
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

// PublishEphemeral broadcasts non-authoritative state without entering the
// durable command queue.
func (a *Actor) PublishEphemeral(message []byte) {
	a.broadcast(message, "", false)
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
	ctx, resp := msg.ctx, msg.resp
	var result Result
	var err error
	if msg.control != nil {
		result, err = a.processControl(ctx, *msg.control)
	} else {
		result, err = a.process(ctx, *msg.env)
	}
	select {
	case resp <- actorResult{result: result, err: err}:
	default:
	}
}

func (a *Actor) process(ctx context.Context, env Envelope) (Result, error) {
	now := time.Now()

	// Non-mutating commands that do not require a transaction.
	switch env.CommandType {
	case "command.status":
		return a.queryCommandStatus(ctx, env)
	case "connection.request_control":
		return a.requestControl(env)
	}

	meta, ok := commandMetadata(env.Command)
	if !ok {
		return Result{}, shared.DomainError{Code: shared.ErrMatchCommandInvalid}
	}
	participantID := meta.AuthenticatedParticipantID

	// Gameplay mutations require the participant's current controller connection.
	if isGameplayCommandType(env.CommandType) {
		a.controllersMu.RLock()
		ctrl := a.controllers[participantID]
		a.controllersMu.RUnlock()
		if ctrl.connectionID != env.ConnectionID {
			return Result{}, shared.DomainError{Code: shared.ErrActionNotAllowedForRole}
		}
	}

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
	if isGameplayCommandType(env.CommandType) {
		a.controllersMu.RLock()
		ctrl := a.controllers[participantID]
		a.controllersMu.RUnlock()
		if meta.ClientSequence <= ctrl.lastSequence {
			return Result{}, shared.DomainError{Code: shared.ErrClientSequenceStale}
		}
	}

	// Apply command to a copy of the aggregate.
	working := a.cloneRoom()
	var roomEvents []roomdomain.Event
	var matchEvents []matchdomain.Event
	var workingMatch *matchdomain.Match
	expectedMatchVersion := int64(0)
	nextEventHash := a.lastEventHash
	nextEventNumber := a.lastEventNumber

	// StartCountdown requires a fresh MatchID and an eligible puzzle when the
	// transport command does not provide them.
	if sc, ok := env.Command.(roomdomain.StartCountdownCommand); ok {
		if sc.MatchID == "" {
			matchID, genErr := idgen.Generator{}.MatchID()
			if genErr != nil {
				return Result{}, shared.DomainError{Code: shared.ErrPersistenceFailed}
			}
			sc.MatchID = matchID
		}
		if len(sc.Puzzle.Clues) == 0 || len(sc.Puzzle.Solution) == 0 {
			record, selErr := txRepo.SelectPuzzleForAssignment(ctx, string(working.Rules.Difficulty), nil, true)
			if selErr != nil {
				return Result{}, shared.DomainError{Code: shared.ErrPersistenceFailed}
			}
			sc.Puzzle = shared.AssignedPuzzle{
				PuzzleID:           shared.PuzzleID(record.ID),
				Revision:           uint32(record.Revision),
				TransformationSeed: 0,
				Difficulty:         working.Rules.Difficulty,
				GeneratorVersion:   record.GeneratorVersion,
				SolverVersion:      record.SolverVersion,
				Clues:              record.Clues,
				Solution:           record.Solution,
			}
		}
		env.Command = sc
	}

	if strings.HasPrefix(env.CommandType, "match.") {
		cmd, ok := env.Command.(matchdomain.Command)
		if !ok {
			return Result{}, shared.DomainError{Code: shared.ErrMatchCommandInvalid}
		}
		workingMatch = a.cloneMatch()
		if workingMatch == nil {
			return Result{}, shared.DomainError{Code: shared.ErrMatchNotActive}
		}
		expectedMatchVersion = int64(workingMatch.Version)
		matchEvents, err = workingMatch.Apply(cmd, a.lastEventNumber+1, now)
		if err == nil && workingMatch.State == shared.MatchCompleted && a.match.State != shared.MatchCompleted {
			completionEvents, completionErr := working.CompleteMatch(workingMatch.ID, now)
			if completionErr != nil {
				return Result{}, completionErr
			}
			roomEvents = append(roomEvents, completionEvents...)
		}
	} else {
		cmd, ok := env.Command.(roomdomain.Command)
		if !ok {
			return Result{}, shared.DomainError{Code: shared.ErrMatchCommandInvalid}
		}
		// Timer-fired commands are authoritative by generation; bind the current room version.
		if am, ok := cmd.(roomdomain.ActivateMatchCommand); ok {
			m := am.Meta
			m.ExpectedVersion = uint64(working.Version)
			am.Meta = m
			cmd = am
		}
		roomEvents, err = working.Apply(cmd, now)
	}
	if err != nil {
		return Result{}, err
	}
	if roomEvents == nil {
		roomEvents = []roomdomain.Event{}
	}

	// Room commands that create or activate the match.
	switch typed := env.Command.(type) {
	case roomdomain.StartCountdownCommand:
		participantIDs := a.playerIDs(working)
		workingMatch, matchEvents, err = matchdomain.NewPrepared(typed.MatchID, working.ID, matchRulesFromRoom(working.Rules), typed.Puzzle, participantIDs, now)
		if err != nil {
			return Result{}, err
		}
	case roomdomain.ActivateMatchCommand:
		workingMatch = a.cloneMatch()
		expectedMatchVersion = int64(workingMatch.Version)
		matchEvents, err = workingMatch.Activate(a.lastEventNumber+1, now)
		if err != nil {
			return Result{}, err
		}
	}

	// Match-only commands do not rewrite an unchanged Room projection.
	if !strings.HasPrefix(env.CommandType, "match.") || len(roomEvents) > 0 {
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
	}

	if env.NewSessionHash != nil {
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

	// Persist match projection and events.
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
			matchEventsGen, lastHash, err := matchEventsToGen(env.RequestID, workingMatch, matchEvents, nextEventHash)
			if err != nil {
				return Result{}, shared.DomainError{Code: shared.ErrPersistenceFailed}
			}
			if err := txRepo.CreateMatchTx(ctx, tx, matchToGen(workingMatch), matchParticipants, matchEventsGen, nil); err != nil {
				return Result{}, shared.DomainError{Code: shared.ErrPersistenceFailed}
			}
			nextEventHash = lastHash
		case roomdomain.ActivateMatchCommand, matchdomain.Command:
			matchEventsGen, lastHash, err := matchEventsToGen(env.RequestID, workingMatch, matchEvents, nextEventHash)
			if err != nil {
				return Result{}, shared.DomainError{Code: shared.ErrPersistenceFailed}
			}
			matchParticipants := matchParticipantsToGen(workingMatch)
			var matchResult *gen.MatchResult
			if workingMatch.State == shared.MatchCompleted && workingMatch.Result != nil {
				matchResult = &gen.MatchResult{
					MatchID:      workingMatch.ID.String(),
					ResultReason: "completed",
					ElapsedMs:    int64(workingMatch.Result.ElapsedMilliseconds),
					Assisted:     boolToInt(workingMatch.Result.Assisted),
					CreatedAtMs:  now.UnixMilli(),
				}
			}
			if err := txRepo.UpdateMatchTx(ctx, tx, matchToGen(workingMatch), expectedMatchVersion, matchParticipants, matchEventsGen, nil, matchResult); err != nil {
				if errors.Is(err, repository.ErrConflict) {
					return Result{}, shared.DomainError{Code: shared.ErrStaleVersion}
				}
				return Result{}, shared.DomainError{Code: shared.ErrPersistenceFailed}
			}
			nextEventHash = lastHash
		}
		if len(matchEvents) > 0 {
			nextEventNumber = uint64(matchEvents[len(matchEvents)-1].Metadata().EventNumber)
		}
	}

	// Revoke the session for a leaving participant.
	if _, ok := env.Command.(roomdomain.LeaveRoomCommand); ok {
		_ = txRepo.RevokeRoomSession(ctx, tx, env.ScopeHash, now.UnixMilli())
	}

	response := a.buildResponse(working, workingMatch, participantID)
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
	a.lastEventHash = nextEventHash
	a.lastEventNumber = nextEventNumber
	if isGameplayCommandType(env.CommandType) {
		a.controllersMu.Lock()
		ctrl := a.controllers[participantID]
		if ctrl.connectionID == env.ConnectionID {
			ctrl.lastSequence = meta.ClientSequence
			a.controllers[participantID] = ctrl
		}
		a.controllersMu.Unlock()
	}
	a.afterCommit(env.Command, working, now)
	originBroadcasts := a.broadcastSnapshotAndEvents(env.ConnectionID, roomEvents, matchEvents)

	return Result{Events: roomEvents, Response: response, OriginBroadcasts: originBroadcasts}, nil
}

func (a *Actor) processControl(ctx context.Context, req controlReq) (Result, error) {
	switch req.kind {
	case "subscribe":
		isController := false
		generation := uint64(1)
		a.controllersMu.Lock()
		current, exists := a.controllers[req.participantID]
		if !exists || current.connectionID == "" {
			a.controllers[req.participantID] = controllerInfo{connectionID: req.connID, generation: 1}
			isController = true
		} else {
			generation = current.generation
		}
		a.controllersMu.Unlock()

		a.subMu.Lock()
		a.subscribers[req.connID] = subscriber{participantID: req.participantID, sendCh: req.sendCh, disconnect: req.disconnect}
		a.subMu.Unlock()

		var role shared.ParticipationRole
		var isHost bool
		for _, p := range a.room.Participants {
			if p.ID == req.participantID {
				role = p.Role
				isHost = p.IsHost
				break
			}
		}
		msg := a.serverMessage("connection.accepted", 0, 0, map[string]any{
			"protocolVersion": 1,
			"identity": map[string]any{
				"participantId": req.participantID.String(),
				"role":          string(role),
				"isHost":        isHost,
			},
			"controllerGeneration": generation,
			"isController":         isController,
		})
		a.sendToConnection(req.connID, msg)
		if !isController {
			a.sendToConnection(req.connID, a.serverMessage("connection.read_only", 0, 0, map[string]any{
				"controllerGeneration": generation,
			}))
		}

		// Send authoritative snapshot(s) to the new subscriber.
		a.sendToConnection(req.connID, a.serverMessage("room.snapshot", 0, uint64(a.room.Version), map[string]any{
			"room": a.buildRoomView(a.room, a.match, req.participantID),
		}))
		if a.match != nil {
			a.sendToConnection(req.connID, a.serverMessage("match.snapshot", a.lastEventNumber, uint64(a.match.Version), map[string]any{
				"match": buildMatchView(a.match),
			}))
		}
		return Result{Response: msg}, nil
	case "sync":
		isController := false
		generation := uint64(1)
		a.controllersMu.Lock()
		current, exists := a.controllers[req.participantID]
		if exists && current.connectionID == req.connID {
			isController = true
			generation = current.generation
		} else {
			generation = current.generation
		}
		a.controllersMu.Unlock()

		var role shared.ParticipationRole
		var isHost bool
		for _, p := range a.room.Participants {
			if p.ID == req.participantID {
				role = p.Role
				isHost = p.IsHost
				break
			}
		}
		_ = a.sendToConnection(req.connID, a.serverMessage("connection.status", 0, 0, map[string]any{
			"identity": map[string]any{
				"participantId": req.participantID.String(),
				"role":          string(role),
				"isHost":        isHost,
			},
			"controllerGeneration": generation,
			"isController":         isController,
		}))
		_ = a.sendToConnection(req.connID, a.serverMessage("room.snapshot", 0, uint64(a.room.Version), map[string]any{
			"room": a.buildRoomView(a.room, a.match, req.participantID),
		}))
		if a.match != nil {
			_ = a.sendToConnection(req.connID, a.serverMessage("match.snapshot", a.lastEventNumber, uint64(a.match.Version), map[string]any{
				"match": buildMatchView(a.match),
			}))
		}
		return Result{}, nil
	case "unsubscribe":
		a.subMu.Lock()
		delete(a.subscribers, req.connID)
		a.subMu.Unlock()

		a.controllersMu.Lock()
		for pid, ctrl := range a.controllers {
			if ctrl.connectionID == req.connID {
				delete(a.controllers, pid)
				break
			}
		}
		a.controllersMu.Unlock()
		return Result{}, nil
	default:
		return Result{}, shared.DomainError{Code: shared.ErrMatchCommandInvalid}
	}
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
	return a.match.Clone()
}

func matchParticipantsToGen(match *matchdomain.Match) []gen.MatchParticipant {
	participants := make([]gen.MatchParticipant, 0, len(match.Participants))
	for _, participantID := range match.Participants {
		participants = append(participants, gen.MatchParticipant{
			MatchID:       match.ID.String(),
			ParticipantID: participantID.String(),
			Connected:     1,
			Mistakes:      int64(match.Mistakes[participantID]),
			HintsUsed:     int64(match.HintsUsed),
		})
	}
	return participants
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

func (a *Actor) afterCommit(cmd any, room *roomdomain.Room, now time.Time) {
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
	roomView := a.buildRoomView(room, match, participantID)
	view := map[string]any{
		"room":         roomView,
		"participants": roomView["participants"],
		"self":         roomView["self"],
	}
	if countdown, ok := roomView["countdown"]; ok {
		view["countdown"] = countdown
	}
	if matchSummary, ok := roomView["match"]; ok {
		view["match"] = matchSummary
	}
	b, _ := json.Marshal(view)
	return b
}

func (a *Actor) buildRoomView(room *roomdomain.Room, match *matchdomain.Match, participantID shared.ParticipantID) map[string]any {
	view := map[string]any{
		"id":             room.ID.String(),
		"code":           room.Code.String(),
		"state":          string(room.State),
		"version":        uint64(room.Version),
		"hostId":         nullParticipantID(room.HostParticipantID),
		"currentMatchId": nullMatchID(room.CurrentMatchID),
		"participants":   participantsToView(room.Participants),
		"settings": map[string]any{
			"mode":              string(room.Rules.Mode),
			"difficulty":        string(room.Rules.Difficulty),
			"errorPreset":       string(room.Rules.ErrorPreset),
			"hintsEnabled":      room.Rules.HintsEnabled,
			"sharedNotes":       room.Rules.SharedNotes,
			"autoRemoveNotes":   room.Rules.AutoRemoveNotes,
			"spectatorsAllowed": room.Rules.SpectatorsAllowed,
		},
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
	return view
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

func participantIDFromCommand(cmd any) shared.ParticipantID {
	if c, ok := cmd.(roomdomain.Command); ok {
		return c.Metadata().AuthenticatedParticipantID
	}
	if c, ok := cmd.(matchdomain.Command); ok {
		return c.Metadata().AuthenticatedParticipantID
	}
	return ""
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

func commandMetadata(cmd any) (shared.CommandMetadata, bool) {
	if c, ok := cmd.(roomdomain.Command); ok {
		return c.Metadata(), true
	}
	if c, ok := cmd.(matchdomain.Command); ok {
		return c.Metadata(), true
	}
	if c, ok := cmd.(ControlCommand); ok {
		return c.Metadata(), true
	}
	return shared.CommandMetadata{}, false
}

func isGameplayCommandType(commandType string) bool {
	switch commandType {
	case "room.set_ready", "room.change_settings", "room.start_countdown",
		"room.cancel_countdown", "room.leave", "room.transfer_host",
		"match.place_value", "match.erase_value", "match.add_note",
		"match.remove_note", "match.use_hint", "match.ping", "match.reaction":
		return true
	}
	return false
}

func (a *Actor) queryCommandStatus(ctx context.Context, env Envelope) (Result, error) {
	receipt, err := a.repo.GetCommandReceipt(ctx, env.RequestID.String())
	if err != nil {
		if repository.IsNoRows(err) {
			return Result{Response: commandStatusResponse(env.RequestID, "COMMAND_OUTCOME_UNKNOWN", nil)}, nil
		}
		return Result{}, shared.DomainError{Code: shared.ErrPersistenceFailed}
	}
	if !bytesEqual(receipt.AuthenticatedScopeHash, env.ScopeHash) {
		return Result{Response: commandStatusResponse(env.RequestID, "COMMAND_OUTCOME_UNKNOWN", nil)}, nil
	}
	var details map[string]any
	if receipt.SafeResponseJson.Valid {
		_ = json.Unmarshal([]byte(receipt.SafeResponseJson.String), &details)
	}
	return Result{Response: commandStatusResponse(env.RequestID, receipt.TerminalStatus, details)}, nil
}

func commandStatusResponse(requestID shared.RequestID, status string, details map[string]any) []byte {
	view := map[string]any{
		"requestId": requestID.String(),
		"status":    status,
	}
	if details != nil {
		view["details"] = details
	}
	b, _ := json.Marshal(view)
	return b
}

func (a *Actor) requestControl(env Envelope) (Result, error) {
	meta, ok := commandMetadata(env.Command)
	if !ok {
		return Result{}, shared.DomainError{Code: shared.ErrMatchCommandInvalid}
	}
	participantID := meta.AuthenticatedParticipantID

	a.controllersMu.Lock()
	current := a.controllers[participantID]
	if current.connectionID == env.ConnectionID {
		a.controllersMu.Unlock()
		return Result{Response: a.controlResponse(participantID, current.generation, false)}, nil
	}
	generation := current.generation + 1
	a.controllers[participantID] = controllerInfo{
		connectionID: env.ConnectionID,
		generation:   generation,
	}
	a.controllersMu.Unlock()

	// Notify old controller it was revoked.
	if current.connectionID != "" {
		a.sendToConnection(current.connectionID, a.serverMessage("connection.controller_revoked", 0, 0, map[string]any{
			"participantId":        participantID.String(),
			"controllerGeneration": generation,
		}))
	}

	return Result{Response: a.controlResponse(participantID, generation, true)}, nil
}

func (a *Actor) controlResponse(participantID shared.ParticipantID, generation uint64, transferred bool) []byte {
	b, _ := json.Marshal(map[string]any{
		"participantId":        participantID.String(),
		"controllerGeneration": generation,
		"transferred":          transferred,
	})
	return b
}

func (a *Actor) broadcastSnapshotAndEvents(originConnID shared.ConnectionID, roomEvents []roomdomain.Event, matchEvents []matchdomain.Event) [][]byte {
	originBroadcasts := make([][]byte, 0, 1+len(matchEvents))
	if len(roomEvents) > 0 {
		msg := a.serverMessage("room.snapshot", 0, uint64(a.room.Version), map[string]any{
			"room": a.buildRoomView(a.room, a.match, ""),
		})
		a.broadcast(msg, originConnID, true)
		originBroadcasts = append(originBroadcasts, msg)
	}
	for _, e := range matchEvents {
		payload, _ := matchEventPublicPayload(e)
		meta := e.Metadata()
		msg := a.serverMessage("match.event", uint64(meta.EventNumber), uint64(meta.AggregateVersion), map[string]any{
			"event": map[string]any{
				"type":    eventTypeName(e),
				"payload": json.RawMessage(payload),
			},
		})
		a.broadcast(msg, originConnID, true)
		originBroadcasts = append(originBroadcasts, msg)
	}
	return originBroadcasts
}

func buildMatchView(m *matchdomain.Match) map[string]any {
	cells := make([]map[string]any, 81)
	for i := 0; i < 81; i++ {
		c := m.Cells[i]
		cellView := map[string]any{
			"index":  i,
			"isClue": c.IsClue,
		}
		if c.Value != nil {
			cellView["value"] = uint8(*c.Value)
		}
		if !c.IsClue {
			cellView["notes"] = c.Notes.Digits()
			cellView["attribution"] = c.Attribution.String()
			cellView["correct"] = c.Correct
		}
		cells[i] = cellView
	}
	values := make(map[string]uint8)
	for idx, d := range m.Values {
		values[strconv.Itoa(int(idx))] = uint8(d)
	}
	mistakes := make(map[string]uint32, len(m.Mistakes))
	for participantID, count := range m.Mistakes {
		mistakes[participantID.String()] = count
	}
	contributions := make(map[string]uint32, len(m.Contributions))
	for participantID, count := range m.Contributions {
		contributions[participantID.String()] = count
	}
	view := map[string]any{
		"id":            m.ID.String(),
		"state":         string(m.State),
		"version":       uint64(m.Version),
		"penaltiesMs":   m.PenaltiesMs,
		"hintsUsed":     m.HintsUsed,
		"assisted":      m.Assisted,
		"mistakes":      mistakes,
		"contributions": contributions,
		"rules": map[string]any{
			"mode":            string(m.Rules.Mode),
			"difficulty":      string(m.Rules.Difficulty),
			"errorPreset":     string(m.Rules.ErrorPreset),
			"hintsEnabled":    m.Rules.HintsEnabled,
			"autoRemoveNotes": m.Rules.AutoRemoveNotes,
			"ruleVersion":     m.Rules.RuleVersion,
		},
		"cells":  cells,
		"values": values,
	}
	if m.StartedAt != nil {
		view["startedAt"] = m.StartedAt.Milliseconds()
	}
	if m.CompletedAt != nil {
		view["completedAt"] = m.CompletedAt.Milliseconds()
	}
	return view
}

func (a *Actor) serverMessage(msgType string, eventNumber uint64, aggregateVersion uint64, payload map[string]any) []byte {
	view := map[string]any{
		"schemaVersion":    1,
		"eventNumber":      eventNumber,
		"aggregateVersion": aggregateVersion,
		"serverTimestamp":  time.Now().UnixMilli(),
		"type":             msgType,
		"payload":          payload,
	}
	b, _ := json.Marshal(view)
	return b
}

func (a *Actor) broadcast(msg []byte, exceptConnID shared.ConnectionID, durable bool) {
	a.subMu.RLock()
	subs := make(map[shared.ConnectionID]subscriber, len(a.subscribers))
	for id, s := range a.subscribers {
		subs[id] = s
	}
	a.subMu.RUnlock()
	for id, s := range subs {
		if id == exceptConnID {
			continue
		}
		select {
		case s.sendCh <- msg:
		default:
			if durable && s.disconnect != nil {
				s.disconnect()
			}
		}
	}
}

func (a *Actor) sendToConnection(connID shared.ConnectionID, msg []byte) bool {
	a.subMu.RLock()
	s, ok := a.subscribers[connID]
	a.subMu.RUnlock()
	if !ok {
		return false
	}
	select {
	case s.sendCh <- msg:
		return true
	default:
		if s.disconnect != nil {
			s.disconnect()
		}
		return false
	}
}

func generateRequestID() string {
	reqID, err := idgen.Generator{}.RequestID()
	if err != nil {
		panic(err)
	}
	return reqID.String()
}
