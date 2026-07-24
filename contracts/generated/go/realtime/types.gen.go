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
	Target ClientMessageTarget `json:"target"`

	// Type corresponds to the JSON schema field "type".
	Type interface{} `json:"type"`
}

type ClientMessagePayload map[string]interface{}

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

type SafeInteger uint64

type SafePositiveInteger uint64

type ServerMessage struct {
	// AggregateVersion corresponds to the JSON schema field "aggregateVersion".
	AggregateVersion SafePositiveInteger `json:"aggregateVersion"`

	// EventNumber corresponds to the JSON schema field "eventNumber".
	EventNumber SafePositiveInteger `json:"eventNumber"`

	// Payload corresponds to the JSON schema field "payload".
	Payload ServerMessagePayload `json:"payload"`

	// SchemaVersion corresponds to the JSON schema field "schemaVersion".
	SchemaVersion interface{} `json:"schemaVersion"`

	// ServerTimestamp corresponds to the JSON schema field "serverTimestamp".
	ServerTimestamp uint64 `json:"serverTimestamp"`

	// Type corresponds to the JSON schema field "type".
	Type ServerMessageType `json:"type"`
}

type ServerMessagePayload map[string]interface{}

type ServerMessageType string

const ServerMessageTypeAck ServerMessageType = "ack"
const ServerMessageTypeRejection ServerMessageType = "rejection"
const ServerMessageTypeSnapshot ServerMessageType = "snapshot"

type Uuidv7 string
