package domain

import (
	"time"

	shared "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/domain"
)

// Rules is the immutable ruleset copied from the Room at Countdown start.
type Rules struct {
	Mode            shared.Mode
	Difficulty      shared.Difficulty
	ErrorPreset     shared.ErrorPreset
	HintsEnabled    bool
	AutoRemoveNotes bool
	RuleVersion     uint16
}

// Result contains the terminal outcome of a completed Match.
type Result struct {
	CompletedAt         shared.Timestamp
	ElapsedMilliseconds uint64
	Assisted            bool
}

// Match is the aggregate root for an individual game.
type Match struct {
	ID           shared.MatchID
	RoomID       shared.RoomID
	Version      shared.MatchVersion
	State        shared.MatchState
	Rules        Rules
	Puzzle       shared.AssignedPuzzle
	Participants []shared.ParticipantID
	Result       *Result
	CreatedAt    shared.Timestamp
}

// Command is the interface implemented by every match command.
type Command interface {
	Metadata() shared.CommandMetadata
}

// Event is the interface implemented by every match event.
type Event interface {
	Metadata() shared.EventMetadata
}

// --- Events ---

type MatchPreparedEvent struct {
	Meta shared.EventMetadata
}

func (e MatchPreparedEvent) Metadata() shared.EventMetadata { return e.Meta }

type MatchCountdownStartedEvent struct {
	Meta       shared.EventMetadata
	DeadlineAt shared.Timestamp
}

func (e MatchCountdownStartedEvent) Metadata() shared.EventMetadata { return e.Meta }

type MatchStartedEvent struct {
	Meta shared.EventMetadata
}

func (e MatchStartedEvent) Metadata() shared.EventMetadata { return e.Meta }

// NewPrepared creates a Match in the Prepared state.
func NewPrepared(id shared.MatchID, roomID shared.RoomID, rules Rules, puzzle shared.AssignedPuzzle, participants []shared.ParticipantID, now time.Time) (*Match, []Event, error) {
	version, err := shared.NewMatchVersion(1)
	if err != nil {
		return nil, nil, err
	}
	ts, err := shared.NewTimestamp(now)
	if err != nil {
		return nil, nil, err
	}
	m := &Match{
		ID:           id,
		RoomID:       roomID,
		Version:      version,
		State:        shared.MatchPrepared,
		Rules:        rules,
		Puzzle:       puzzle,
		Participants: participants,
	}
	meta1, err := eventMeta(m, 1, now)
	if err != nil {
		return nil, nil, err
	}
	meta2, err := eventMeta(m, 2, now)
	if err != nil {
		return nil, nil, err
	}
	events := []Event{
		MatchPreparedEvent{Meta: meta1},
		MatchCountdownStartedEvent{Meta: meta2, DeadlineAt: ts},
	}
	return m, events, nil
}

// Activate transitions a Prepared Match into Active.
func (m *Match) Activate(now time.Time) ([]Event, error) {
	if m.State != shared.MatchPrepared && m.State != shared.MatchCountdown {
		return nil, shared.DomainError{Code: shared.ErrMatchStateInvalid}
	}
	m.State = shared.MatchActive
	m.Version++
	meta, err := eventMeta(m, 3, now)
	if err != nil {
		return nil, err
	}
	return []Event{MatchStartedEvent{Meta: meta}}, nil
}

func eventMeta(m *Match, number uint64, now time.Time) (shared.EventMetadata, error) {
	return shared.NewEventMetadata(1, shared.EventNumber(number), uint64(m.Version), shared.NewMatchTarget(m.ID), mustTimestamp(now))
}

func mustTimestamp(now time.Time) shared.Timestamp {
	ts, err := shared.NewTimestamp(now)
	if err != nil {
		panic(err)
	}
	return ts
}
