// Code generated from contracts. DO NOT EDIT.

package replay

type ReplayDocument struct {
	// Events corresponds to the JSON schema field "events".
	Events []ReplayDocumentEventsElem `json:"events"`

	// MatchId corresponds to the JSON schema field "matchId".
	MatchId Uuidv7 `json:"matchId"`

	// ReplayId corresponds to the JSON schema field "replayId".
	ReplayId Uuidv7 `json:"replayId"`

	// SchemaVersion corresponds to the JSON schema field "schemaVersion".
	SchemaVersion interface{} `json:"schemaVersion"`
}

type ReplayDocumentEventsElem struct {
	// AggregateVersion corresponds to the JSON schema field "aggregateVersion".
	AggregateVersion SafePositiveInteger `json:"aggregateVersion"`

	// EventNumber corresponds to the JSON schema field "eventNumber".
	EventNumber SafePositiveInteger `json:"eventNumber"`

	// Payload corresponds to the JSON schema field "payload".
	Payload ReplayDocumentEventsElemPayload `json:"payload"`

	// ServerTimestamp corresponds to the JSON schema field "serverTimestamp".
	ServerTimestamp SafePositiveInteger `json:"serverTimestamp"`

	// Type corresponds to the JSON schema field "type".
	Type string `json:"type"`
}

type ReplayDocumentEventsElemPayload map[string]interface{}

type ReplayProof struct {
	// FinalHash corresponds to the JSON schema field "finalHash".
	FinalHash string `json:"finalHash"`

	// KeyId corresponds to the JSON schema field "keyId".
	KeyId string `json:"keyId"`

	// SchemaVersion corresponds to the JSON schema field "schemaVersion".
	SchemaVersion interface{} `json:"schemaVersion"`

	// Signature corresponds to the JSON schema field "signature".
	Signature string `json:"signature"`
}

type SafePositiveInteger uint64

type Uuidv7 string
