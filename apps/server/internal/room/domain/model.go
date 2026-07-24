package domain

import shared "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/domain"

type Participant struct {
	ID      shared.ParticipantID
	Name    shared.DisplayName
	Role    shared.ParticipationRole
	IsHost  bool
	IsReady bool
}

type Rules struct {
	Mode            shared.Mode
	Difficulty      shared.Difficulty
	ErrorPreset     shared.ErrorPreset
	HintsEnabled    bool
	SharedNotes     bool
	AutoRemoveNotes bool
	Spectators      bool
}

type Room struct {
	ID             shared.RoomID
	Code           shared.RoomCode
	Version        shared.RoomVersion
	State          shared.RoomState
	Participants   []Participant
	Rules          Rules
	CurrentMatchID *shared.MatchID
}

type Command interface {
	Metadata() shared.CommandMetadata
}

type Event interface {
	Metadata() shared.EventMetadata
}
