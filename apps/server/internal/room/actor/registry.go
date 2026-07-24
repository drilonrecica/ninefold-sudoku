package actor

import (
	"context"
	"fmt"
	"sync"
	"time"

	"log/slog"

	shared "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/domain"
	matchdomain "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/match/domain"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/gen"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/repository"
	replayproof "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/replay/proof"
	roomdomain "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/room/domain"
)

const idleTimeout = 30 * time.Second

// Registry guarantees at most one active actor per Room.
type Registry struct {
	repo   *repository.Repository
	logger *slog.Logger
	mu     sync.Mutex
	actors map[shared.RoomID]*actorEntry
	signer replayproof.Signer
}

// RecoverNonTerminal reconstructs every active match before readiness. Recovered
// actors remain registered because they own the authoritative recovery deadline.
func (reg *Registry) RecoverNonTerminal(ctx context.Context, now time.Time) error {
	matches, err := reg.repo.ListNonTerminalMatches(ctx)
	if err != nil {
		return err
	}
	for _, match := range matches {
		if match.State != string(shared.MatchActive) && match.State != string(shared.MatchRecoveryPending) {
			continue
		}
		roomID := shared.RoomID(match.RoomID)
		reg.mu.Lock()
		if _, exists := reg.actors[roomID]; exists {
			reg.mu.Unlock()
			continue
		}
		loaded, loadErr := reg.loadActor(ctx, roomID)
		if loadErr == nil {
			reg.actors[roomID] = &actorEntry{actor: loaded}
		}
		reg.mu.Unlock()
		if loadErr != nil {
			return fmt.Errorf("recover match %s: %w", match.ID, loadErr)
		}
		if match.State == string(shared.MatchActive) {
			if err := loaded.EnterRecovery(ctx, now); err != nil {
				return fmt.Errorf("enter recovery for match %s: %w", match.ID, err)
			}
		} else {
			loaded.ResumeRecoveryDeadline(now)
		}
	}
	return nil
}

type actorEntry struct {
	actor     *Actor
	refs      int
	idleTimer *time.Timer
	mu        sync.Mutex
}

// NewRegistry creates an empty registry backed by the supplied repository.
func NewRegistry(repo *repository.Repository, logger *slog.Logger, signers ...replayproof.Signer) *Registry {
	registry := &Registry{
		repo:   repo,
		logger: logger,
		actors: make(map[shared.RoomID]*actorEntry),
	}
	if len(signers) > 0 {
		registry.signer = signers[0]
	}
	return registry
}

// Acquire returns the active actor for a room, creating it from persistence if needed.
// The caller must call Release when finished.
func (reg *Registry) Acquire(ctx context.Context, roomID shared.RoomID) (*Actor, error) {
	reg.mu.Lock()
	defer reg.mu.Unlock()

	entry, ok := reg.actors[roomID]
	if ok && entry.actor != nil {
		entry.mu.Lock()
		entry.refs++
		if entry.idleTimer != nil {
			entry.idleTimer.Stop()
			entry.idleTimer = nil
		}
		entry.mu.Unlock()
		return entry.actor, nil
	}

	actor, err := reg.loadActor(ctx, roomID)
	if err != nil {
		return nil, err
	}
	entry = &actorEntry{actor: actor, refs: 1}
	reg.actors[roomID] = entry
	return actor, nil
}

// Release decrements the reference count and schedules deactivation when idle.
func (reg *Registry) Release(roomID shared.RoomID) {
	reg.mu.Lock()
	entry, ok := reg.actors[roomID]
	if !ok {
		reg.mu.Unlock()
		return
	}
	reg.mu.Unlock()

	entry.mu.Lock()
	defer entry.mu.Unlock()
	entry.refs--
	if entry.refs > 0 {
		return
	}
	if entry.actor.HasActiveTimers() {
		return
	}
	entry.idleTimer = time.AfterFunc(idleTimeout, func() {
		reg.evict(roomID)
	})
}

// Activate creates an actor for a brand-new room that has not yet been persisted.
func (reg *Registry) Activate(room *roomdomain.Room) *Actor {
	reg.mu.Lock()
	defer reg.mu.Unlock()
	actor := NewActor(room, nil, reg.repo, reg.logger, reg.signer)
	reg.actors[room.ID] = &actorEntry{actor: actor, refs: 1}
	return actor
}

// ShutdownAll stops every actor in the registry.
func (reg *Registry) ShutdownAll() {
	reg.mu.Lock()
	entries := make([]*actorEntry, 0, len(reg.actors))
	for _, e := range reg.actors {
		entries = append(entries, e)
	}
	reg.actors = make(map[shared.RoomID]*actorEntry)
	reg.mu.Unlock()

	for _, e := range entries {
		if e.actor != nil {
			e.actor.Stop()
		}
	}
}

// NotifyMaintenance broadcasts a reconnectable shutdown state before actors
// persist their final snapshots.
func (reg *Registry) NotifyMaintenance() {
	reg.mu.Lock()
	actors := make([]*Actor, 0, len(reg.actors))
	for _, entry := range reg.actors {
		if entry.actor != nil {
			actors = append(actors, entry.actor)
		}
	}
	reg.mu.Unlock()
	for _, actor := range actors {
		actor.NotifyMaintenance()
	}
}

func (reg *Registry) evict(roomID shared.RoomID) {
	reg.mu.Lock()
	entry, ok := reg.actors[roomID]
	if !ok {
		reg.mu.Unlock()
		return
	}
	entry.mu.Lock()
	if entry.refs > 0 || entry.actor.HasActiveTimers() {
		entry.mu.Unlock()
		reg.mu.Unlock()
		return
	}
	delete(reg.actors, roomID)
	entry.mu.Unlock()
	reg.mu.Unlock()
	entry.actor.Stop()
}

func (reg *Registry) loadActor(ctx context.Context, roomID shared.RoomID) (*Actor, error) {
	gr, err := reg.repo.GetRoomByID(ctx, roomID.String())
	if err != nil {
		return nil, err
	}
	participants, err := reg.repo.ListActiveRoomParticipants(ctx, roomID.String())
	if err != nil {
		return nil, err
	}
	room, err := roomFromGen(gr, participants)
	if err != nil {
		return nil, err
	}
	var match *matchdomain.Match
	var lastEventNumber uint64
	var lastEventHash []byte
	var domainEvents []matchdomain.Event
	var validSnapshot *gen.MatchSnapshot
	var snapshotState *matchSnapshotState
	if room.CurrentMatchID != nil {
		gm, err := reg.repo.GetMatchByID(ctx, room.CurrentMatchID.String())
		if err != nil {
			return nil, err
		}
		matchParticipants, err := reg.repo.ListMatchParticipants(ctx, gm.ID)
		if err != nil {
			return nil, err
		}
		ids := make([]shared.ParticipantID, 0, len(matchParticipants))
		for _, mp := range matchParticipants {
			ids = append(ids, shared.ParticipantID(mp.ParticipantID))
		}
		puzzle, err := reg.repo.GetPuzzle(ctx, gm.PuzzleID, gm.PuzzleRevision)
		if err != nil {
			return nil, err
		}
		rules := matchRulesFromGenMatch(gm)
		assigned := shared.AssignedPuzzle{
			PuzzleID:           shared.PuzzleID(puzzle.ID),
			Revision:           uint32(puzzle.Revision),
			TransformationSeed: uint64(gm.TransformationSeed),
			Difficulty:         rules.Difficulty,
			GeneratorVersion:   puzzle.GeneratorVersion,
			SolverVersion:      puzzle.SolverVersion,
			Clues:              puzzle.Clues,
			Solution:           puzzle.Solution,
		}
		events, err := reg.repo.GetMatchEvents(ctx, gm.ID)
		if err != nil {
			return nil, err
		}
		if err := validatePersistedEventChain(events); err != nil {
			return nil, err
		}
		if snapshot, snapshotErr := reg.repo.GetLatestMatchSnapshot(ctx, gm.ID); snapshotErr == nil {
			decoded, decodeErr := decodeSnapshot(snapshot)
			boundary := int(snapshot.EventNumber)
			if decodeErr == nil && boundary > 0 && boundary <= len(events) &&
				events[boundary-1].EventNumber == snapshot.EventNumber &&
				events[boundary-1].AggregateVersion == snapshot.AggregateVersion {
				snapshotCopy := snapshot
				stateCopy := decoded
				validSnapshot = &snapshotCopy
				snapshotState = &stateCopy
			}
		}
		domainEvents = make([]matchdomain.Event, 0, len(events))
		for _, e := range events {
			de, err := matchEventFromGen(e)
			if err != nil {
				return nil, err
			}
			domainEvents = append(domainEvents, de)
			lastEventNumber = uint64(e.EventNumber)
			lastEventHash = e.EventHash
		}
		if snapshotState != nil {
			match, err = snapshotState.restore(assigned, rules, ids)
			if err == nil {
				for _, event := range domainEvents {
					if uint64(event.Metadata().EventNumber) <= uint64(validSnapshot.EventNumber) {
						continue
					}
					if err = match.ApplyEvent(event); err != nil {
						break
					}
					match.Version = shared.MatchVersion(event.Metadata().AggregateVersion)
				}
			}
			if err != nil {
				validSnapshot = nil
				snapshotState = nil
				match = nil
				err = nil
			}
		}
		if match == nil {
			match, err = matchdomain.ReconstructMatch(assigned, rules, ids, domainEvents)
			if err != nil {
				return nil, err
			}
		}
		match.ID = shared.MatchID(gm.ID)
		match.RoomID = shared.RoomID(gm.RoomID)
		match.Version = shared.MatchVersion(gm.Version)
	}
	actor := NewActor(room, match, reg.repo, reg.logger, reg.signer)
	actor.lastEventNumber = lastEventNumber
	actor.lastEventHash = lastEventHash
	if validSnapshot != nil {
		actor.lastSnapshotEvent = uint64(validSnapshot.EventNumber)
		actor.lastSnapshotAt = time.UnixMilli(validSnapshot.CreatedAtMs)
	}
	for _, event := range domainEvents {
		actor.appendBufferedDomainEvent(event)
	}
	return actor, nil
}

func matchRulesFromGenMatch(gm gen.Match) matchdomain.Rules {
	mode, _ := shared.ParseMode(gm.Mode)
	difficulty, _ := shared.ParseDifficulty(gm.Difficulty)
	errorPreset, _ := shared.ParseErrorPreset(gm.ErrorPreset)
	return matchdomain.Rules{
		Mode:            mode,
		Difficulty:      difficulty,
		ErrorPreset:     errorPreset,
		HintsEnabled:    gm.HintsEnabled != 0,
		AutoRemoveNotes: gm.AutoRemoveNotes != 0,
		RuleVersion:     uint16(gm.RuleVersion),
	}
}
