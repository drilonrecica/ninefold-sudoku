package domain

type Mode string

const (
	ModeCoop Mode = "Coop"
	ModeSolo Mode = "Solo"
)

func ParseMode(value string) (Mode, error) {
	return validateEnum("mode", Mode(value), ModeCoop, ModeSolo)
}

type Difficulty string

const (
	DifficultyEasy   Difficulty = "Easy"
	DifficultyMedium Difficulty = "Medium"
	DifficultyHard   Difficulty = "Hard"
	DifficultyExpert Difficulty = "Expert"
	DifficultyRandom Difficulty = "Random"
)

func ParseDifficulty(value string) (Difficulty, error) {
	return validateEnum("difficulty", Difficulty(value), DifficultyEasy, DifficultyMedium, DifficultyHard, DifficultyExpert, DifficultyRandom)
}

type ParticipationRole string

const (
	RolePlayer    ParticipationRole = "Player"
	RoleSpectator ParticipationRole = "Spectator"
)

func ParseParticipationRole(value string) (ParticipationRole, error) {
	return validateEnum("participation role", ParticipationRole(value), RolePlayer, RoleSpectator)
}

type RoomState string

const (
	RoomLobby             RoomState = "Lobby"
	RoomCountdown         RoomState = "Countdown"
	RoomInMatch           RoomState = "InMatch"
	RoomResults           RoomState = "Results"
	RoomExpired           RoomState = "Expired"
	RoomCancelled         RoomState = "Cancelled"
	RoomRecoveryPending   RoomState = "RecoveryPending"
	RoomTerminatedByAdmin RoomState = "TerminatedByAdmin"
)

func ParseRoomState(value string) (RoomState, error) {
	return validateEnum("room state", RoomState(value), RoomLobby, RoomCountdown, RoomInMatch, RoomResults, RoomExpired, RoomCancelled, RoomRecoveryPending, RoomTerminatedByAdmin)
}

type MatchState string

const (
	MatchPrepared        MatchState = "Prepared"
	MatchCountdown       MatchState = "Countdown"
	MatchActive          MatchState = "Active"
	MatchCompleted       MatchState = "Completed"
	MatchRecoveryPending MatchState = "RecoveryPending"
	MatchCancelled       MatchState = "Cancelled"
	MatchAbandoned       MatchState = "Abandoned"
)

func ParseMatchState(value string) (MatchState, error) {
	return validateEnum("match state", MatchState(value), MatchPrepared, MatchCountdown, MatchActive, MatchCompleted, MatchRecoveryPending, MatchCancelled, MatchAbandoned)
}

type PuzzleState string

const (
	PuzzleDraft    PuzzleState = "Draft"
	PuzzleVerified PuzzleState = "Verified"
	PuzzleActive   PuzzleState = "Active"
	PuzzleRetired  PuzzleState = "Retired"
)

func ParsePuzzleState(value string) (PuzzleState, error) {
	return validateEnum("puzzle state", PuzzleState(value), PuzzleDraft, PuzzleVerified, PuzzleActive, PuzzleRetired)
}

type ErrorPreset string

const (
	ErrorPresetCasual    ErrorPreset = "Casual"
	ErrorPresetChallenge ErrorPreset = "Challenge"
	ErrorPresetBlind     ErrorPreset = "Blind"
	ErrorPresetClean     ErrorPreset = "Clean"
)

func ParseErrorPreset(value string) (ErrorPreset, error) {
	return validateEnum("error preset", ErrorPreset(value), ErrorPresetCasual, ErrorPresetChallenge, ErrorPresetBlind, ErrorPresetClean)
}

type HintLevel string

const (
	HintNudge  HintLevel = "Nudge"
	HintReveal HintLevel = "Reveal"
)

func ParseHintLevel(value string) (HintLevel, error) {
	return validateEnum("hint level", HintLevel(value), HintNudge, HintReveal)
}
