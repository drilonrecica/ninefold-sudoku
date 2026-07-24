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

type ReplayDocumentParticipantsElem struct {
	// Id corresponds to the JSON schema field "id".
	Id Uuidv7 `json:"id"`

	// Name corresponds to the JSON schema field "name".
	Name string `json:"name"`
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
