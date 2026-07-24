package domain

import (
	"time"

	shared "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/domain"
)

// Participant represents a temporary room-scoped identity.
type Participant struct {
	ID            shared.ParticipantID
	Name          shared.DisplayName
	Role          shared.ParticipationRole
	IsHost        bool
	IsReady       bool
	JoinedAt      shared.Timestamp
	LeftAt        *shared.Timestamp
	RemovedAt     *shared.Timestamp
	RemovedReason string
}

// IsActive reports whether the participant is still present in the room.
func (p Participant) IsActive() bool {
	return p.LeftAt == nil && p.RemovedAt == nil
}

// MatchRules is the immutable ruleset snapshotted when a Match is prepared.
type MatchRules struct {
	Mode              shared.Mode
	Difficulty        shared.Difficulty
	ErrorPreset       shared.ErrorPreset
	HintsEnabled      bool
	SharedNotes       bool
	AutoRemoveNotes   bool
	SpectatorsAllowed bool
	RuleVersion       uint16
}

// CountdownState holds the transient preparation data between Countdown start
// and Match activation.
type CountdownState struct {
	MatchID    shared.MatchID
	Generation uint64
	DeadlineAt shared.Timestamp
	Rules      MatchRules
	Puzzle     shared.AssignedPuzzle
}

// Room is the aggregate root for the room lifecycle.
type Room struct {
	ID                shared.RoomID
	Code              shared.RoomCode
	Version           shared.RoomVersion
	State             shared.RoomState
	Participants      []Participant
	Rules             MatchRules
	HostParticipantID *shared.ParticipantID
	CurrentMatchID    *shared.MatchID
	CreatedAt         shared.Timestamp
	LastActivityAt    shared.Timestamp
	ExpiresAt         shared.Timestamp
	Countdown         *CountdownState
}

// Command is the interface implemented by every room command.
type Command interface {
	Metadata() shared.CommandMetadata
}

// Event is the interface implemented by every room event.
type Event interface {
	Metadata() shared.EventMetadata
}

// --- Commands ---

type CreateRoomCommand struct {
	Meta        shared.CommandMetadata
	DisplayName shared.DisplayName
	Mode        shared.Mode
	Difficulty  shared.Difficulty
}

func (c CreateRoomCommand) Metadata() shared.CommandMetadata { return c.Meta }

type RequestJoinCommand struct {
	Meta          shared.CommandMetadata
	Code          shared.RoomCode
	DisplayName   shared.DisplayName
	Role          shared.ParticipationRole
	ParticipantID shared.ParticipantID
}

func (c RequestJoinCommand) Metadata() shared.CommandMetadata { return c.Meta }

type LeaveRoomCommand struct {
	Meta   shared.CommandMetadata
	Intent string
}

func (c LeaveRoomCommand) Metadata() shared.CommandMetadata { return c.Meta }

type ChangeParticipationRoleCommand struct {
	Meta        shared.CommandMetadata
	Participant shared.ParticipantID
	Role        shared.ParticipationRole
}

func (c ChangeParticipationRoleCommand) Metadata() shared.CommandMetadata { return c.Meta }

type SetReadyCommand struct {
	Meta  shared.CommandMetadata
	Ready bool
}

func (c SetReadyCommand) Metadata() shared.CommandMetadata { return c.Meta }

type ChangeRoomSettingsCommand struct {
	Meta     shared.CommandMetadata
	Settings RoomSettingsPatch
}

func (c ChangeRoomSettingsCommand) Metadata() shared.CommandMetadata { return c.Meta }

// RoomSettingsPatch contains the host-controlled settings that may change in the Lobby.
type RoomSettingsPatch struct {
	Difficulty        *shared.Difficulty
	ErrorPreset       *shared.ErrorPreset
	HintsEnabled      *bool
	SharedNotes       *bool
	AutoRemoveNotes   *bool
	SpectatorsAllowed *bool
}

type LockRoomCommand struct{ Meta shared.CommandMetadata }

func (c LockRoomCommand) Metadata() shared.CommandMetadata { return c.Meta }

type UnlockRoomCommand struct{ Meta shared.CommandMetadata }

func (c UnlockRoomCommand) Metadata() shared.CommandMetadata { return c.Meta }

type RemoveParticipantCommand struct {
	Meta        shared.CommandMetadata
	Participant shared.ParticipantID
}

func (c RemoveParticipantCommand) Metadata() shared.CommandMetadata { return c.Meta }

type BlockParticipantCommand struct {
	Meta        shared.CommandMetadata
	Participant shared.ParticipantID
}

func (c BlockParticipantCommand) Metadata() shared.CommandMetadata { return c.Meta }

type TransferHostCommand struct {
	Meta        shared.CommandMetadata
	Participant shared.ParticipantID
}

func (c TransferHostCommand) Metadata() shared.CommandMetadata { return c.Meta }

type StartCountdownCommand struct {
	Meta    shared.CommandMetadata
	MatchID shared.MatchID
	Puzzle  shared.AssignedPuzzle
}

func (c StartCountdownCommand) Metadata() shared.CommandMetadata { return c.Meta }

type CancelCountdownCommand struct{ Meta shared.CommandMetadata }

func (c CancelCountdownCommand) Metadata() shared.CommandMetadata { return c.Meta }

type ActivateMatchCommand struct {
	Meta       shared.CommandMetadata
	Generation uint64
}

func (c ActivateMatchCommand) Metadata() shared.CommandMetadata { return c.Meta }

type ExpireRoomCommand struct{ Meta shared.CommandMetadata }

func (c ExpireRoomCommand) Metadata() shared.CommandMetadata { return c.Meta }

// --- Events ---

type RoomCreatedEvent struct {
	Meta        shared.EventMetadata
	Code        shared.RoomCode
	HostID      shared.ParticipantID
	Rules       MatchRules
	Participant Participant
}

func (e RoomCreatedEvent) Metadata() shared.EventMetadata { return e.Meta }

type ParticipantJoinedEvent struct {
	Meta        shared.EventMetadata
	Participant Participant
	Role        shared.ParticipationRole
}

func (e ParticipantJoinedEvent) Metadata() shared.EventMetadata { return e.Meta }

type ParticipantLeftEvent struct {
	Meta          shared.EventMetadata
	ParticipantID shared.ParticipantID
	Intent        string
}

func (e ParticipantLeftEvent) Metadata() shared.EventMetadata { return e.Meta }

type ParticipantRemovedEvent struct {
	Meta          shared.EventMetadata
	ParticipantID shared.ParticipantID
	Reason        string
}

func (e ParticipantRemovedEvent) Metadata() shared.EventMetadata { return e.Meta }

type ParticipantBlockedEvent struct {
	Meta          shared.EventMetadata
	ParticipantID shared.ParticipantID
}

func (e ParticipantBlockedEvent) Metadata() shared.EventMetadata { return e.Meta }

type ParticipantRoleChangedEvent struct {
	Meta          shared.EventMetadata
	ParticipantID shared.ParticipantID
	From          shared.ParticipationRole
	To            shared.ParticipationRole
}

func (e ParticipantRoleChangedEvent) Metadata() shared.EventMetadata { return e.Meta }

type ParticipantReadyStateChangedEvent struct {
	Meta          shared.EventMetadata
	ParticipantID shared.ParticipantID
	Ready         bool
}

func (e ParticipantReadyStateChangedEvent) Metadata() shared.EventMetadata { return e.Meta }

type RoomSettingsChangedEvent struct {
	Meta     shared.EventMetadata
	Settings MatchRules
}

func (e RoomSettingsChangedEvent) Metadata() shared.EventMetadata { return e.Meta }

type RoomReadyStatesResetEvent struct{ Meta shared.EventMetadata }

func (e RoomReadyStatesResetEvent) Metadata() shared.EventMetadata { return e.Meta }

type RoomLockedEvent struct{ Meta shared.EventMetadata }

func (e RoomLockedEvent) Metadata() shared.EventMetadata { return e.Meta }

type RoomUnlockedEvent struct{ Meta shared.EventMetadata }

func (e RoomUnlockedEvent) Metadata() shared.EventMetadata { return e.Meta }

type HostTransferredEvent struct {
	Meta shared.EventMetadata
	From *shared.ParticipantID
	To   shared.ParticipantID
}

func (e HostTransferredEvent) Metadata() shared.EventMetadata { return e.Meta }

type CountdownStartedEvent struct {
	Meta       shared.EventMetadata
	MatchID    shared.MatchID
	Generation uint64
	DeadlineAt shared.Timestamp
	Rules      MatchRules
}

func (e CountdownStartedEvent) Metadata() shared.EventMetadata { return e.Meta }

type CountdownCancelledEvent struct{ Meta shared.EventMetadata }

func (e CountdownCancelledEvent) Metadata() shared.EventMetadata { return e.Meta }

type RoomEnteredMatchEvent struct {
	Meta    shared.EventMetadata
	MatchID shared.MatchID
}

func (e RoomEnteredMatchEvent) Metadata() shared.EventMetadata { return e.Meta }

type RoomEnteredResultsEvent struct {
	Meta    shared.EventMetadata
	MatchID shared.MatchID
}

func (e RoomEnteredResultsEvent) Metadata() shared.EventMetadata { return e.Meta }

type RoomExpiredEvent struct{ Meta shared.EventMetadata }

func (e RoomExpiredEvent) Metadata() shared.EventMetadata { return e.Meta }

// roomState checks whether the room is in one of the supplied states.
func (r *Room) isState(states ...shared.RoomState) bool {
	for _, s := range states {
		if r.State == s {
			return true
		}
	}
	return false
}

// nowTimestamp builds a domain Timestamp from the supplied clock.
func nowTimestamp(now time.Time) (shared.Timestamp, error) {
	return shared.NewTimestamp(now)
}
