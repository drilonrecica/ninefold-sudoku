// Code generated from contracts. DO NOT EDIT.

package replay

type ReplayDocument struct {
	// Clues corresponds to the JSON schema field "clues".
	Clues string `json:"clues"`

	// Events corresponds to the JSON schema field "events".
	Events []ReplayDocumentEventsElem `json:"events"`

	// ExpiresAt corresponds to the JSON schema field "expiresAt".
	ExpiresAt SafePositiveInteger `json:"expiresAt"`

	// MatchId corresponds to the JSON schema field "matchId".
	MatchId Uuidv7 `json:"matchId"`

	// Participants corresponds to the JSON schema field "participants".
	Participants []ReplayDocumentParticipantsElem `json:"participants"`

	// Proof corresponds to the JSON schema field "proof".
	Proof ReplayDocumentProof `json:"proof"`

	// ReplayId corresponds to the JSON schema field "replayId".
	ReplayId Uuidv7 `json:"replayId"`

	// Rules corresponds to the JSON schema field "rules".
	Rules ReplayDocumentRules `json:"rules"`

	// SchemaVersion corresponds to the JSON schema field "schemaVersion".
	SchemaVersion interface{} `json:"schemaVersion"`
}

type ReplayDocumentEventsElem struct {
	// AggregateVersion corresponds to the JSON schema field "aggregateVersion".
	AggregateVersion SafePositiveInteger `json:"aggregateVersion"`

	// EventHash corresponds to the JSON schema field "eventHash".
	EventHash string `json:"eventHash"`

	// EventNumber corresponds to the JSON schema field "eventNumber".
	EventNumber SafePositiveInteger `json:"eventNumber"`

	// OccurredAtMs corresponds to the JSON schema field "occurredAtMs".
	OccurredAtMs SafePositiveInteger `json:"occurredAtMs"`

	// PreviousEventHash corresponds to the JSON schema field "previousEventHash".
	PreviousEventHash string `json:"previousEventHash"`

	// PrivatePayloadDigest corresponds to the JSON schema field
	// "privatePayloadDigest".
	PrivatePayloadDigest interface{} `json:"privatePayloadDigest"`

	// ProofVersion corresponds to the JSON schema field "proofVersion".
	ProofVersion interface{} `json:"proofVersion"`

	// PublicActorId corresponds to the JSON schema field "publicActorId".
	PublicActorId string `json:"publicActorId"`

	// PublicEventType corresponds to the JSON schema field "publicEventType".
	PublicEventType string `json:"publicEventType"`

	// PublicPayload corresponds to the JSON schema field "publicPayload".
	PublicPayload ReplayDocumentEventsElemPublicPayload `json:"publicPayload"`
}

type ReplayDocumentEventsElemPublicPayload map[string]interface{}

type ReplayDocumentParticipantsElem struct {
	// Id corresponds to the JSON schema field "id".
	Id Uuidv7 `json:"id"`

	// Name corresponds to the JSON schema field "name".
	Name string `json:"name"`
}

type ReplayDocumentProof struct {
	// FinalEventHash corresponds to the JSON schema field "finalEventHash".
	FinalEventHash string `json:"finalEventHash"`

	// FinalEventNumber corresponds to the JSON schema field "finalEventNumber".
	FinalEventNumber SafePositiveInteger `json:"finalEventNumber"`

	// KeyId corresponds to the JSON schema field "keyId".
	KeyId string `json:"keyId"`

	// MatchId corresponds to the JSON schema field "matchId".
	MatchId Uuidv7 `json:"matchId"`

	// ProofVersion corresponds to the JSON schema field "proofVersion".
	ProofVersion interface{} `json:"proofVersion"`

	// Signature corresponds to the JSON schema field "signature".
	Signature string `json:"signature"`

	// TerminalAtMs corresponds to the JSON schema field "terminalAtMs".
	TerminalAtMs SafePositiveInteger `json:"terminalAtMs"`
}

type ReplayDocumentRules struct {
	// AutoRemoveNotes corresponds to the JSON schema field "autoRemoveNotes".
	AutoRemoveNotes bool `json:"autoRemoveNotes"`

	// Difficulty corresponds to the JSON schema field "difficulty".
	Difficulty ReplayDocumentRulesDifficulty `json:"difficulty"`

	// ErrorPreset corresponds to the JSON schema field "errorPreset".
	ErrorPreset ReplayDocumentRulesErrorPreset `json:"errorPreset"`

	// HintsEnabled corresponds to the JSON schema field "hintsEnabled".
	HintsEnabled bool `json:"hintsEnabled"`

	// Mode corresponds to the JSON schema field "mode".
	Mode interface{} `json:"mode"`

	// RuleVersion corresponds to the JSON schema field "ruleVersion".
	RuleVersion SafePositiveInteger `json:"ruleVersion"`
}

type ReplayDocumentRulesDifficulty string

const ReplayDocumentRulesDifficultyEasy ReplayDocumentRulesDifficulty = "Easy"
const ReplayDocumentRulesDifficultyExpert ReplayDocumentRulesDifficulty = "Expert"
const ReplayDocumentRulesDifficultyHard ReplayDocumentRulesDifficulty = "Hard"
const ReplayDocumentRulesDifficultyMedium ReplayDocumentRulesDifficulty = "Medium"

type ReplayDocumentRulesErrorPreset string

const ReplayDocumentRulesErrorPresetBlind ReplayDocumentRulesErrorPreset = "Blind"
const ReplayDocumentRulesErrorPresetCasual ReplayDocumentRulesErrorPreset = "Casual"
const ReplayDocumentRulesErrorPresetChallenge ReplayDocumentRulesErrorPreset = "Challenge"
const ReplayDocumentRulesErrorPresetClean ReplayDocumentRulesErrorPreset = "Clean"

type ReplayProof struct {
	// FinalEventHash corresponds to the JSON schema field "finalEventHash".
	FinalEventHash string `json:"finalEventHash"`

	// FinalEventNumber corresponds to the JSON schema field "finalEventNumber".
	FinalEventNumber uint64 `json:"finalEventNumber"`

	// KeyId corresponds to the JSON schema field "keyId".
	KeyId string `json:"keyId"`

	// MatchId corresponds to the JSON schema field "matchId".
	MatchId string `json:"matchId"`

	// ProofVersion corresponds to the JSON schema field "proofVersion".
	ProofVersion interface{} `json:"proofVersion"`

	// Signature corresponds to the JSON schema field "signature".
	Signature string `json:"signature"`

	// TerminalAtMs corresponds to the JSON schema field "terminalAtMs".
	TerminalAtMs uint64 `json:"terminalAtMs"`
}

type SafePositiveInteger uint64

type Uuidv7 string
