package domain

import shared "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/domain"

type Rules struct {
	Mode            shared.Mode
	Difficulty      shared.Difficulty
	ErrorPreset     shared.ErrorPreset
	HintsEnabled    bool
	AutoRemoveNotes bool
	RuleVersion     uint16
}

type AssignedPuzzle struct {
	PuzzleID           shared.PuzzleID
	Revision           uint32
	TransformationSeed uint64
	Difficulty         shared.Difficulty
	GeneratorVersion   string
	SolverVersion      string
}

type Result struct {
	CompletedAt         shared.Timestamp
	ElapsedMilliseconds uint64
	Assisted            bool
}

type Match struct {
	ID           shared.MatchID
	RoomID       *shared.RoomID
	Version      shared.MatchVersion
	State        shared.MatchState
	Rules        Rules
	Puzzle       AssignedPuzzle
	Participants []shared.ParticipantID
	Result       *Result
}

type Command interface {
	Metadata() shared.CommandMetadata
}

type Event interface {
	Metadata() shared.EventMetadata
}
