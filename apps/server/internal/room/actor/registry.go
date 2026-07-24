package actor

import (
	"context"
	"sync"
	"time"

	"log/slog"

	shared "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/domain"
	matchdomain "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/match/domain"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/gen"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/repository"
	roomdomain "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/room/domain"
)

const idleTimeout = 30 * time.Second

// Registry guarantees at most one active actor per Room.
type Registry struct {
	repo   *repository.Repository
	logger *slog.Logger
	mu     sync.Mutex
	actors map[shared.RoomID]*actorEntry
}

type actorEntry struct {
	actor     *Actor
	refs      int
	idleTimer *time.Timer
	mu        sync.Mutex
}

// NewRegistry creates an empty registry backed by the supplied repository.
func NewRegistry(repo *repository.Repository, logger *slog.Logger) *Registry {
	return &Registry{
		repo:   repo,
		logger: logger,
		actors: make(map[shared.RoomID]*actorEntry),
	}
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
	actor := NewActor(room, nil, reg.repo, reg.logger)
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
		domainEvents := make([]matchdomain.Event, 0, len(events))
		for _, e := range events {
			de, err := matchEventFromGen(e)
			if err != nil {
				return nil, err
			}
			domainEvents = append(domainEvents, de)
			lastEventNumber = uint64(e.EventNumber)
			lastEventHash = e.EventHash
		}
		match, err = matchdomain.ReconstructMatch(assigned, rules, ids, domainEvents)
		if err != nil {
			return nil, err
		}
		match.ID = shared.MatchID(gm.ID)
		match.RoomID = shared.RoomID(gm.RoomID)
		match.Version = shared.MatchVersion(gm.Version)
	}
	actor := NewActor(room, match, reg.repo, reg.logger)
	actor.lastEventNumber = lastEventNumber
	actor.lastEventHash = lastEventHash
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
