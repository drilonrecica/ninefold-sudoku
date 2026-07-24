// Code generated from contracts. DO NOT EDIT.

package realtime

type ClientMessage struct {
	// ClientSequence corresponds to the JSON schema field "clientSequence".
	ClientSequence SafePositiveInteger `json:"clientSequence"`

	// Payload corresponds to the JSON schema field "payload".
	Payload ClientMessagePayload `json:"payload"`

	// RequestId corresponds to the JSON schema field "requestId".
	RequestId Uuidv7 `json:"requestId"`

	// SchemaVersion corresponds to the JSON schema field "schemaVersion".
	SchemaVersion interface{} `json:"schemaVersion"`

	// Target corresponds to the JSON schema field "target".
	Target *ClientMessageTarget `json:"target,omitempty,omitzero"`

	// Type corresponds to the JSON schema field "type".
	Type ClientMessageType `json:"type"`
}

type ClientMessagePayload struct {
	// Cell corresponds to the JSON schema field "cell".
	Cell *uint8 `json:"cell,omitempty,omitzero"`

	// Digits corresponds to the JSON schema field "digits".
	Digits []uint8 `json:"digits,omitempty,omitzero"`

	// DisplayName corresponds to the JSON schema field "displayName".
	DisplayName *string `json:"displayName,omitempty,omitzero"`

	// Intent corresponds to the JSON schema field "intent".
	Intent *string `json:"intent,omitempty,omitzero"`

	// LastMatchEventNumber corresponds to the JSON schema field
	// "lastMatchEventNumber".
	LastMatchEventNumber *SafeInteger `json:"lastMatchEventNumber,omitempty,omitzero"`

	// LastMatchId corresponds to the JSON schema field "lastMatchId".
	LastMatchId *Uuidv7 `json:"lastMatchId,omitempty,omitzero"`

	// LastRoomVersion corresponds to the JSON schema field "lastRoomVersion".
	LastRoomVersion *SafeInteger `json:"lastRoomVersion,omitempty,omitzero"`

	// Level corresponds to the JSON schema field "level".
	Level *ClientMessagePayloadLevel `json:"level,omitempty,omitzero"`

	// ParticipantId corresponds to the JSON schema field "participantId".
	ParticipantId *Uuidv7 `json:"participantId,omitempty,omitzero"`

	// Reaction corresponds to the JSON schema field "reaction".
	Reaction *ClientMessagePayloadReaction `json:"reaction,omitempty,omitzero"`

	// Ready corresponds to the JSON schema field "ready".
	Ready *bool `json:"ready,omitempty,omitzero"`

	// RoomCode corresponds to the JSON schema field "roomCode".
	RoomCode *string `json:"roomCode,omitempty,omitzero"`

	// Settings corresponds to the JSON schema field "settings".
	Settings *ClientMessagePayloadSettings `json:"settings,omitempty,omitzero"`

	// TargetCell corresponds to the JSON schema field "targetCell".
	TargetCell *uint8 `json:"targetCell,omitempty,omitzero"`

	// Value corresponds to the JSON schema field "value".
	Value *uint8 `json:"value,omitempty,omitzero"`
}

type ClientMessagePayloadLevel string

const ClientMessagePayloadLevelNudge ClientMessagePayloadLevel = "Nudge"
const ClientMessagePayloadLevelReveal ClientMessagePayloadLevel = "Reveal"

type ClientMessagePayloadReaction string

const ClientMessagePayloadReactionAgree ClientMessagePayloadReaction = "agree"
const ClientMessagePayloadReactionNiceMove ClientMessagePayloadReaction = "nice_move"

type ClientMessagePayloadSettings struct {
	// AutoRemoveNotes corresponds to the JSON schema field "autoRemoveNotes".
	AutoRemoveNotes *bool `json:"autoRemoveNotes,omitempty,omitzero"`

	// Difficulty corresponds to the JSON schema field "difficulty".
	Difficulty *string `json:"difficulty,omitempty,omitzero"`

	// ErrorPreset corresponds to the JSON schema field "errorPreset".
	ErrorPreset *string `json:"errorPreset,omitempty,omitzero"`

	// HintsEnabled corresponds to the JSON schema field "hintsEnabled".
	HintsEnabled *bool `json:"hintsEnabled,omitempty,omitzero"`

	// Locked corresponds to the JSON schema field "locked".
	Locked *bool `json:"locked,omitempty,omitzero"`
}

type ClientMessageTarget struct {
	// ExpectedVersion corresponds to the JSON schema field "expectedVersion".
	ExpectedVersion SafeInteger `json:"expectedVersion"`

	// Id corresponds to the JSON schema field "id".
	Id Uuidv7 `json:"id"`

	// Kind corresponds to the JSON schema field "kind".
	Kind ClientMessageTargetKind `json:"kind"`
}

type ClientMessageTargetKind string

const ClientMessageTargetKindMatch ClientMessageTargetKind = "Match"
const ClientMessageTargetKindRoom ClientMessageTargetKind = "Room"

type ClientMessageType string

const ClientMessageTypeCommandStatus ClientMessageType = "command.status"
const ClientMessageTypeConnectionHeartbeat ClientMessageType = "connection.heartbeat"
const ClientMessageTypeConnectionInitialize ClientMessageType = "connection.initialize"
const ClientMessageTypeConnectionRequestControl ClientMessageType = "connection.request_control"
const ClientMessageTypeMatchAddNote ClientMessageType = "match.add_note"
const ClientMessageTypeMatchEraseValue ClientMessageType = "match.erase_value"
const ClientMessageTypeMatchFocusCell ClientMessageType = "match.focus_cell"
const ClientMessageTypeMatchPing ClientMessageType = "match.ping"
const ClientMessageTypeMatchPlaceValue ClientMessageType = "match.place_value"
const ClientMessageTypeMatchReaction ClientMessageType = "match.reaction"
const ClientMessageTypeMatchReleaseFocus ClientMessageType = "match.release_focus"
const ClientMessageTypeMatchRemoveNote ClientMessageType = "match.remove_note"
const ClientMessageTypeMatchUseHint ClientMessageType = "match.use_hint"
const ClientMessageTypeRoomCancelCountdown ClientMessageType = "room.cancel_countdown"
const ClientMessageTypeRoomChangeSettings ClientMessageType = "room.change_settings"
const ClientMessageTypeRoomLeave ClientMessageType = "room.leave"
const ClientMessageTypeRoomPrepareRematch ClientMessageType = "room.prepare_rematch"
const ClientMessageTypeRoomSetReady ClientMessageType = "room.set_ready"
const ClientMessageTypeRoomStartCountdown ClientMessageType = "room.start_countdown"
const ClientMessageTypeRoomTransferHost ClientMessageType = "room.transfer_host"

type SafeInteger uint64

type SafeInteger_1 uint64

type SafePositiveInteger uint64

type ServerMessage struct {
	// AggregateVersion corresponds to the JSON schema field "aggregateVersion".
	AggregateVersion SafeInteger_1 `json:"aggregateVersion"`

	// EventNumber corresponds to the JSON schema field "eventNumber".
	EventNumber SafeInteger_1 `json:"eventNumber"`

	// Payload corresponds to the JSON schema field "payload".
	Payload ServerMessagePayload `json:"payload"`

	// SchemaVersion corresponds to the JSON schema field "schemaVersion".
	SchemaVersion interface{} `json:"schemaVersion"`

	// ServerTimestamp corresponds to the JSON schema field "serverTimestamp".
	ServerTimestamp uint64 `json:"serverTimestamp"`

	// Type corresponds to the JSON schema field "type".
	Type ServerMessageType `json:"type"`
}

type ServerMessagePayload struct {
	// Accepted corresponds to the JSON schema field "accepted".
	Accepted *bool `json:"accepted,omitempty,omitzero"`

	// Aggregate corresponds to the JSON schema field "aggregate".
	Aggregate *ServerMessagePayloadAggregate `json:"aggregate,omitempty,omitzero"`

	// Cell corresponds to the JSON schema field "cell".
	Cell *uint8 `json:"cell,omitempty,omitzero"`

	// Code corresponds to the JSON schema field "code".
	Code *string `json:"code,omitempty,omitzero"`

	// ConnectionState corresponds to the JSON schema field "connectionState".
	ConnectionState *ServerMessagePayloadConnectionState `json:"connectionState,omitempty,omitzero"`

	// ControllerGeneration corresponds to the JSON schema field
	// "controllerGeneration".
	ControllerGeneration *SafeInteger_1 `json:"controllerGeneration,omitempty,omitzero"`

	// CurrentMatchEventNumber corresponds to the JSON schema field
	// "currentMatchEventNumber".
	CurrentMatchEventNumber *SafeInteger_1 `json:"currentMatchEventNumber,omitempty,omitzero"`

	// CurrentVersion corresponds to the JSON schema field "currentVersion".
	CurrentVersion *SafeInteger_1 `json:"currentVersion,omitempty,omitzero"`

	// Details corresponds to the JSON schema field "details".
	Details ServerMessagePayloadDetails `json:"details,omitempty,omitzero"`

	// Event corresponds to the JSON schema field "event".
	Event ServerMessagePayloadEvent `json:"event,omitempty,omitzero"`

	// EventNumber corresponds to the JSON schema field "eventNumber".
	EventNumber *SafePositiveInteger `json:"eventNumber,omitempty,omitzero"`

	// Focused corresponds to the JSON schema field "focused".
	Focused *bool `json:"focused,omitempty,omitzero"`

	// Identity corresponds to the JSON schema field "identity".
	Identity *ServerMessagePayloadIdentity `json:"identity,omitempty,omitzero"`

	// Intent corresponds to the JSON schema field "intent".
	Intent *string `json:"intent,omitempty,omitzero"`

	// IsController corresponds to the JSON schema field "isController".
	IsController *bool `json:"isController,omitempty,omitzero"`

	// Match corresponds to the JSON schema field "match".
	Match ServerMessagePayloadMatch `json:"match,omitempty,omitzero"`

	// MatchVersion corresponds to the JSON schema field "matchVersion".
	MatchVersion *SafePositiveInteger `json:"matchVersion,omitempty,omitzero"`

	// Message corresponds to the JSON schema field "message".
	Message *string `json:"message,omitempty,omitzero"`

	// ParticipantId corresponds to the JSON schema field "participantId".
	ParticipantId *Uuidv7 `json:"participantId,omitempty,omitzero"`

	// ProtocolVersion corresponds to the JSON schema field "protocolVersion".
	ProtocolVersion *SafePositiveInteger `json:"protocolVersion,omitempty,omitzero"`

	// Reaction corresponds to the JSON schema field "reaction".
	Reaction *ServerMessagePayloadReaction `json:"reaction,omitempty,omitzero"`

	// RequestId corresponds to the JSON schema field "requestId".
	RequestId *Uuidv7 `json:"requestId,omitempty,omitzero"`

	// ResultingVersion corresponds to the JSON schema field "resultingVersion".
	ResultingVersion *SafeInteger_1 `json:"resultingVersion,omitempty,omitzero"`

	// Room corresponds to the JSON schema field "room".
	Room ServerMessagePayloadRoom `json:"room,omitempty,omitzero"`

	// RoomVersion corresponds to the JSON schema field "roomVersion".
	RoomVersion *SafePositiveInteger `json:"roomVersion,omitempty,omitzero"`

	// Status corresponds to the JSON schema field "status".
	Status *string `json:"status,omitempty,omitzero"`
}

type ServerMessagePayloadAggregate string

const ServerMessagePayloadAggregateMatch ServerMessagePayloadAggregate = "match"
const ServerMessagePayloadAggregateRoom ServerMessagePayloadAggregate = "room"

type ServerMessagePayloadConnectionState string

const ServerMessagePayloadConnectionStateConnected ServerMessagePayloadConnectionState = "connected"
const ServerMessagePayloadConnectionStateMaintenance ServerMessagePayloadConnectionState = "maintenance"
const ServerMessagePayloadConnectionStateReadOnly ServerMessagePayloadConnectionState = "read_only"
const ServerMessagePayloadConnectionStateReconnecting ServerMessagePayloadConnectionState = "reconnecting"
const ServerMessagePayloadConnectionStateRecoveryFailed ServerMessagePayloadConnectionState = "recovery_failed"
const ServerMessagePayloadConnectionStateSynchronizing ServerMessagePayloadConnectionState = "synchronizing"

type ServerMessagePayloadDetails map[string]interface{}

type ServerMessagePayloadEvent map[string]interface{}

type ServerMessagePayloadIdentity struct {
	// IsHost corresponds to the JSON schema field "isHost".
	IsHost *bool `json:"isHost,omitempty,omitzero"`

	// ParticipantId corresponds to the JSON schema field "participantId".
	ParticipantId *Uuidv7 `json:"participantId,omitempty,omitzero"`

	// Role corresponds to the JSON schema field "role".
	Role *string `json:"role,omitempty,omitzero"`
}

type ServerMessagePayloadMatch map[string]interface{}

type ServerMessagePayloadReaction string

const ServerMessagePayloadReactionAgree ServerMessagePayloadReaction = "agree"
const ServerMessagePayloadReactionNiceMove ServerMessagePayloadReaction = "nice_move"

type ServerMessagePayloadRoom map[string]interface{}

type ServerMessageType string

const ServerMessageTypeCommandAcknowledged ServerMessageType = "command.acknowledged"
const ServerMessageTypeCommandRejected ServerMessageType = "command.rejected"
const ServerMessageTypeCommandStatus ServerMessageType = "command.status"
const ServerMessageTypeConnectionAccepted ServerMessageType = "connection.accepted"
const ServerMessageTypeConnectionControllerRevoked ServerMessageType = "connection.controller_revoked"
const ServerMessageTypeConnectionReadOnly ServerMessageType = "connection.read_only"
const ServerMessageTypeConnectionRejected ServerMessageType = "connection.rejected"
const ServerMessageTypeConnectionStatus ServerMessageType = "connection.status"
const ServerMessageTypeEphemeralFocus ServerMessageType = "ephemeral.focus"
const ServerMessageTypeEphemeralReaction ServerMessageType = "ephemeral.reaction"
const ServerMessageTypeEphemeralSoftLock ServerMessageType = "ephemeral.soft_lock"
const ServerMessageTypeMatchCompleted ServerMessageType = "match.completed"
const ServerMessageTypeMatchEvent ServerMessageType = "match.event"
const ServerMessageTypeMatchSnapshot ServerMessageType = "match.snapshot"
const ServerMessageTypeRoomEvent ServerMessageType = "room.event"
const ServerMessageTypeRoomSnapshot ServerMessageType = "room.snapshot"

type Uuidv7 string
