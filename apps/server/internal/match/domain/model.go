package domain

import (
	"time"

	shared "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/domain"
)

// Rules is the immutable ruleset copied from the Room at Match creation.
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
	MistakesByPlayer    map[shared.ParticipantID]uint32
	HintCount           uint32
	ContributionCount   uint32
}

// Cell represents the domain view of a single Sudoku cell.
type Cell struct {
	Index       shared.CellIndex
	IsClue      bool
	Value       *shared.Digit
	Notes       shared.CandidateSet
	Attribution shared.ParticipantID
	// Correct is nil when the match rules hide correctness from clients.
	Correct *bool
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

	Cells         [81]Cell
	Values        map[shared.CellIndex]shared.Digit
	Attribution   map[shared.CellIndex]shared.ParticipantID
	Notes         [81]shared.CandidateSet
	Mistakes      map[shared.ParticipantID]uint32
	Contributions map[shared.ParticipantID]uint32
	HintsUsed     uint32
	PenaltiesMs   uint64
	Assisted      bool
	StartedAt     *shared.Timestamp
	CompletedAt   *shared.Timestamp
	Result        *Result
	CreatedAt     shared.Timestamp

	processedRequestIDs map[shared.RequestID]struct{}
}

// Clone returns an independent aggregate copy suitable for speculative command
// application before the persistence transaction commits.
func (m *Match) Clone() *Match {
	if m == nil {
		return nil
	}
	clone := *m
	clone.Participants = append([]shared.ParticipantID(nil), m.Participants...)
	clone.Values = cloneDigitMap(m.Values)
	clone.Attribution = cloneAttributionMap(m.Attribution)
	clone.Mistakes = cloneCountMap(m.Mistakes)
	clone.Contributions = cloneCountMap(m.Contributions)
	clone.processedRequestIDs = make(map[shared.RequestID]struct{}, len(m.processedRequestIDs))
	for requestID := range m.processedRequestIDs {
		clone.processedRequestIDs[requestID] = struct{}{}
	}
	if m.Result != nil {
		result := *m.Result
		result.MistakesByPlayer = cloneCountMap(m.Result.MistakesByPlayer)
		clone.Result = &result
	}
	return &clone
}

func cloneDigitMap(source map[shared.CellIndex]shared.Digit) map[shared.CellIndex]shared.Digit {
	clone := make(map[shared.CellIndex]shared.Digit, len(source))
	for key, value := range source {
		clone[key] = value
	}
	return clone
}

func cloneAttributionMap(source map[shared.CellIndex]shared.ParticipantID) map[shared.CellIndex]shared.ParticipantID {
	clone := make(map[shared.CellIndex]shared.ParticipantID, len(source))
	for key, value := range source {
		clone[key] = value
	}
	return clone
}

func cloneCountMap(source map[shared.ParticipantID]uint32) map[shared.ParticipantID]uint32 {
	clone := make(map[shared.ParticipantID]uint32, len(source))
	for key, value := range source {
		clone[key] = value
	}
	return clone
}

// Command is the interface implemented by every match command.
type Command interface {
	Metadata() shared.CommandMetadata
}

// Event is the interface implemented by every match event.
type Event interface {
	Metadata() shared.EventMetadata
}

// --- Commands ---

type PlaceValueCommand struct {
	Meta  shared.CommandMetadata
	Cell  shared.CellIndex
	Digit shared.Digit
}

func (c PlaceValueCommand) Metadata() shared.CommandMetadata { return c.Meta }

type EraseValueCommand struct {
	Meta shared.CommandMetadata
	Cell shared.CellIndex
}

func (c EraseValueCommand) Metadata() shared.CommandMetadata { return c.Meta }

type AddNoteCommand struct {
	Meta   shared.CommandMetadata
	Cell   shared.CellIndex
	Digits []shared.Digit
}

func (c AddNoteCommand) Metadata() shared.CommandMetadata { return c.Meta }

type RemoveNoteCommand struct {
	Meta   shared.CommandMetadata
	Cell   shared.CellIndex
	Digits []shared.Digit
}

func (c RemoveNoteCommand) Metadata() shared.CommandMetadata { return c.Meta }

type UseHintCommand struct {
	Meta   shared.CommandMetadata
	Level  shared.HintLevel
	Target *shared.CellIndex // nil for Nudge if no specific cell is requested
}

func (c UseHintCommand) Metadata() shared.CommandMetadata { return c.Meta }

type PingCommand struct {
	Meta   shared.CommandMetadata
	Cell   shared.CellIndex
	Intent string
}

func (c PingCommand) Metadata() shared.CommandMetadata { return c.Meta }

type ParticipantDisconnectedCommand struct {
	Meta          shared.CommandMetadata
	ParticipantID shared.ParticipantID
}

func (c ParticipantDisconnectedCommand) Metadata() shared.CommandMetadata { return c.Meta }

type ParticipantReconnectedCommand struct {
	Meta          shared.CommandMetadata
	ParticipantID shared.ParticipantID
}

func (c ParticipantReconnectedCommand) Metadata() shared.CommandMetadata { return c.Meta }

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

type ValuePlacedEvent struct {
	Meta          shared.EventMetadata
	Cell          shared.CellIndex
	Digit         shared.Digit
	ParticipantID shared.ParticipantID
	Correct       *bool
	Conflict      bool
	ReplacesValue bool
	IsHint        bool
}

func (e ValuePlacedEvent) Metadata() shared.EventMetadata { return e.Meta }

type ValueRejectedEvent struct {
	Meta          shared.EventMetadata
	Cell          shared.CellIndex
	Digit         shared.Digit
	ParticipantID shared.ParticipantID
	Reason        string
	PenaltyMs     uint64
}

func (e ValueRejectedEvent) Metadata() shared.EventMetadata { return e.Meta }

type ValueErasedEvent struct {
	Meta          shared.EventMetadata
	Cell          shared.CellIndex
	ParticipantID shared.ParticipantID
}

func (e ValueErasedEvent) Metadata() shared.EventMetadata { return e.Meta }

type NotesAddedEvent struct {
	Meta          shared.EventMetadata
	Cell          shared.CellIndex
	Digits        []shared.Digit
	ParticipantID shared.ParticipantID
}

func (e NotesAddedEvent) Metadata() shared.EventMetadata { return e.Meta }

type NotesRemovedEvent struct {
	Meta          shared.EventMetadata
	Cell          shared.CellIndex
	Digits        []shared.Digit
	ParticipantID shared.ParticipantID
}

func (e NotesRemovedEvent) Metadata() shared.EventMetadata { return e.Meta }

type NotesAutoRemovedEvent struct {
	Meta     shared.EventMetadata
	Cell     shared.CellIndex
	Digits   []shared.Digit
	CausedBy shared.CellIndex
}

func (e NotesAutoRemovedEvent) Metadata() shared.EventMetadata { return e.Meta }

type HintUsedEvent struct {
	Meta          shared.EventMetadata
	Level         shared.HintLevel
	TargetCell    *shared.CellIndex
	Digit         *shared.Digit
	ParticipantID shared.ParticipantID
}

func (e HintUsedEvent) Metadata() shared.EventMetadata { return e.Meta }

type PingEvent struct {
	Meta          shared.EventMetadata
	Cell          shared.CellIndex
	Intent        string
	ParticipantID shared.ParticipantID
}

func (e PingEvent) Metadata() shared.EventMetadata { return e.Meta }

type ParticipantDisconnectedEvent struct {
	Meta          shared.EventMetadata
	ParticipantID shared.ParticipantID
}

func (e ParticipantDisconnectedEvent) Metadata() shared.EventMetadata { return e.Meta }

type ParticipantReconnectedEvent struct {
	Meta          shared.EventMetadata
	ParticipantID shared.ParticipantID
}

func (e ParticipantReconnectedEvent) Metadata() shared.EventMetadata { return e.Meta }

type MatchCompletedEvent struct {
	Meta   shared.EventMetadata
	Result Result
}

func (e MatchCompletedEvent) Metadata() shared.EventMetadata { return e.Meta }

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
		ID:                  id,
		RoomID:              roomID,
		Version:             version,
		State:               shared.MatchPrepared,
		Rules:               rules,
		Puzzle:              puzzle,
		Participants:        participants,
		Values:              make(map[shared.CellIndex]shared.Digit),
		Attribution:         make(map[shared.CellIndex]shared.ParticipantID),
		Mistakes:            make(map[shared.ParticipantID]uint32),
		Contributions:       make(map[shared.ParticipantID]uint32),
		processedRequestIDs: make(map[shared.RequestID]struct{}),
		CreatedAt:           ts,
	}
	m.initCells()

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

func (m *Match) initCells() {
	for i := 0; i < 81; i++ {
		idx := shared.CellIndex(i)
		isClue := m.Puzzle.Clues[i] != 0
		m.Cells[i] = Cell{
			Index:  idx,
			IsClue: isClue,
		}
		if isClue {
			d := shared.Digit(m.Puzzle.Clues[i])
			m.Cells[i].Value = &d
			m.Values[idx] = d
		}
	}
}

// Activate transitions a Prepared Match into Active.
func (m *Match) Activate(nextEventNumber uint64, now time.Time) ([]Event, error) {
	if m.State != shared.MatchPrepared && m.State != shared.MatchCountdown {
		return nil, shared.DomainError{Code: shared.ErrMatchStateInvalid}
	}
	m.State = shared.MatchActive
	m.Version++
	ts, err := shared.NewTimestamp(now)
	if err != nil {
		return nil, err
	}
	m.StartedAt = &ts
	meta, err := eventMeta(m, nextEventNumber, now)
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

func (m *Match) isParticipant(id shared.ParticipantID) bool {
	for _, p := range m.Participants {
		if p == id {
			return true
		}
	}
	return false
}

func (m *Match) bumpVersion() {
	m.Version++
}

func (m *Match) activeEventNumber(eventNumber uint64) uint64 {
	// Event numbers are not tracked inside the aggregate; callers use persisted event numbers.
	return eventNumber
}
