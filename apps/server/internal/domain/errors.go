package domain

import (
	"errors"
	"fmt"
	"strings"
)

type ErrorCode string

const (
	ErrRoomNotFound             ErrorCode = "ROOM_NOT_FOUND"
	ErrRoomExpired              ErrorCode = "ROOM_EXPIRED"
	ErrRoomCancelled            ErrorCode = "ROOM_CANCELLED"
	ErrRoomTerminated           ErrorCode = "ROOM_TERMINATED"
	ErrRoomLocked               ErrorCode = "ROOM_LOCKED"
	ErrRoomFull                 ErrorCode = "ROOM_FULL"
	ErrSpectatorCapacityReached ErrorCode = "SPECTATOR_CAPACITY_REACHED"
	ErrSessionInvalid           ErrorCode = "SESSION_INVALID"
	ErrSessionExpired           ErrorCode = "SESSION_EXPIRED"
	ErrActiveRoomSessionExists  ErrorCode = "ACTIVE_ROOM_SESSION_EXISTS"
	ErrNameInvalid              ErrorCode = "NAME_INVALID"
	ErrNameAlreadyUsed          ErrorCode = "NAME_ALREADY_USED"
	ErrParticipantBlocked       ErrorCode = "PARTICIPANT_BLOCKED"
	ErrParticipantNotFound      ErrorCode = "PARTICIPANT_NOT_FOUND"
	ErrNotRoomHost              ErrorCode = "NOT_ROOM_HOST"
	ErrHostTransferInvalid      ErrorCode = "HOST_TRANSFER_INVALID"
	ErrRoleChangeInvalid        ErrorCode = "ROLE_CHANGE_INVALID"
	ErrPlayerNotReady           ErrorCode = "PLAYER_NOT_READY"
	ErrPlayersNotReady          ErrorCode = "PLAYERS_NOT_READY"
	ErrInsufficientPlayers      ErrorCode = "INSUFFICIENT_PLAYERS"
	ErrInvalidPlayerCount       ErrorCode = "INVALID_PLAYER_COUNT"
	ErrCountdownAlreadyStarted  ErrorCode = "COUNTDOWN_ALREADY_STARTED"
	ErrCountdownNotActive       ErrorCode = "COUNTDOWN_NOT_ACTIVE"
	ErrMatchAlreadyStarted      ErrorCode = "MATCH_ALREADY_STARTED"
	ErrMatchAlreadyExists       ErrorCode = "MATCH_ALREADY_EXISTS"
	ErrMatchNotActive           ErrorCode = "MATCH_NOT_ACTIVE"
	ErrMatchAlreadyCompleted    ErrorCode = "MATCH_ALREADY_COMPLETED"
	ErrMatchPuzzleInvalid       ErrorCode = "MATCH_PUZZLE_INVALID"
	ErrMatchRulesInvalid        ErrorCode = "MATCH_RULES_INVALID"
	ErrMatchCommandInvalid      ErrorCode = "MATCH_COMMAND_INVALID"
	ErrRoomStateInvalid         ErrorCode = "ROOM_STATE_INVALID"
	ErrMatchStateInvalid        ErrorCode = "MATCH_STATE_INVALID"
	ErrSettingsLocked           ErrorCode = "SETTINGS_LOCKED"
	ErrCellIndexInvalid         ErrorCode = "CELL_INDEX_INVALID"
	ErrDigitInvalid             ErrorCode = "DIGIT_INVALID"
	ErrCellFixed                ErrorCode = "CELL_FIXED"
	ErrCellSoftLocked           ErrorCode = "CELL_SOFT_LOCKED"
	ErrCellNotEditable          ErrorCode = "CELL_NOT_EDITABLE"
	ErrInvalidValue             ErrorCode = "INVALID_VALUE"
	ErrValueRejectedByRules     ErrorCode = "VALUE_REJECTED_BY_RULES"
	ErrNoteInvalid              ErrorCode = "NOTE_INVALID"
	ErrHintsDisabled            ErrorCode = "HINTS_DISABLED"
	ErrHintLevelUnavailable     ErrorCode = "HINT_LEVEL_UNAVAILABLE"
	ErrActionNotAllowedForRole  ErrorCode = "ACTION_NOT_ALLOWED_FOR_ROLE"
	ErrActionNotAllowedInMode   ErrorCode = "ACTION_NOT_ALLOWED_IN_MODE"
	ErrReplayNotAvailable       ErrorCode = "REPLAY_NOT_AVAILABLE"
	ErrStaleVersion             ErrorCode = "STALE_VERSION"
	ErrDuplicateRequest         ErrorCode = "DUPLICATE_REQUEST"
	ErrClientSequenceStale      ErrorCode = "CLIENT_SEQUENCE_STALE"
	ErrRecoveryRequired         ErrorCode = "RECOVERY_REQUIRED"
	ErrRecoveryFailed           ErrorCode = "RECOVERY_FAILED"
	ErrReconnectWindowExpired   ErrorCode = "RECONNECT_WINDOW_EXPIRED"
	ErrTimerTokenStale          ErrorCode = "TIMER_TOKEN_STALE"
	ErrCommandNotRetryable      ErrorCode = "COMMAND_NOT_RETRYABLE"
	ErrCommandOutcomeUnknown    ErrorCode = "COMMAND_OUTCOME_UNKNOWN"
	ErrServerBusy               ErrorCode = "SERVER_BUSY"
	ErrRateLimited              ErrorCode = "RATE_LIMITED"
	ErrPersistenceFailed        ErrorCode = "PERSISTENCE_FAILED"
	ErrReplayExpired            ErrorCode = "REPLAY_EXPIRED"
	ErrReplayDeleted            ErrorCode = "REPLAY_DELETED"
	ErrReplayCapabilityInvalid  ErrorCode = "REPLAY_CAPABILITY_INVALID"
	ErrReplayEventGap           ErrorCode = "REPLAY_EVENT_GAP"
	ErrReplayHashInvalid        ErrorCode = "REPLAY_HASH_INVALID"
	ErrReplaySignatureInvalid   ErrorCode = "REPLAY_SIGNATURE_INVALID"
	ErrReplayFormatUnsupported  ErrorCode = "REPLAY_FORMAT_UNSUPPORTED"
)

var allErrorCodes = []ErrorCode{
	ErrRoomNotFound, ErrRoomExpired, ErrRoomCancelled, ErrRoomTerminated, ErrRoomLocked, ErrRoomFull,
	ErrSpectatorCapacityReached, ErrSessionInvalid, ErrSessionExpired, ErrActiveRoomSessionExists,
	ErrNameInvalid, ErrNameAlreadyUsed, ErrParticipantBlocked, ErrParticipantNotFound, ErrNotRoomHost,
	ErrHostTransferInvalid, ErrRoleChangeInvalid, ErrPlayerNotReady, ErrPlayersNotReady,
	ErrInsufficientPlayers, ErrInvalidPlayerCount, ErrCountdownAlreadyStarted, ErrCountdownNotActive,
	ErrMatchAlreadyStarted, ErrMatchAlreadyExists, ErrMatchNotActive, ErrMatchAlreadyCompleted,
	ErrMatchPuzzleInvalid, ErrMatchRulesInvalid, ErrMatchCommandInvalid,
	ErrRoomStateInvalid, ErrMatchStateInvalid, ErrSettingsLocked, ErrCellIndexInvalid, ErrDigitInvalid,
	ErrCellFixed, ErrCellSoftLocked, ErrCellNotEditable, ErrInvalidValue, ErrValueRejectedByRules,
	ErrNoteInvalid, ErrHintsDisabled, ErrHintLevelUnavailable, ErrActionNotAllowedForRole,
	ErrActionNotAllowedInMode, ErrReplayNotAvailable, ErrStaleVersion, ErrDuplicateRequest,
	ErrClientSequenceStale, ErrRecoveryRequired, ErrRecoveryFailed, ErrReconnectWindowExpired,
	ErrTimerTokenStale, ErrCommandNotRetryable, ErrCommandOutcomeUnknown, ErrServerBusy, ErrRateLimited,
	ErrPersistenceFailed, ErrReplayExpired, ErrReplayDeleted, ErrReplayCapabilityInvalid,
	ErrReplayEventGap, ErrReplayHashInvalid, ErrReplaySignatureInvalid, ErrReplayFormatUnsupported,
}

var retryableCodes = map[ErrorCode]bool{
	ErrStaleVersion:          true,
	ErrDuplicateRequest:      true,
	ErrClientSequenceStale:   true,
	ErrRecoveryRequired:      true,
	ErrCommandOutcomeUnknown: true,
	ErrServerBusy:            true,
	ErrRateLimited:           true,
	ErrPersistenceFailed:     true,
}

type DomainError struct {
	Code      ErrorCode
	Retryable bool
	Details   map[string]any
}

func NewDomainError(code ErrorCode, details map[string]any) (DomainError, error) {
	if !isKnownErrorCode(code) {
		return DomainError{}, fmt.Errorf("unknown domain error code %q", code)
	}
	safe := make(map[string]any, len(details))
	for key, value := range details {
		lower := strings.ToLower(key)
		if key == "" || strings.ContainsAny(key, ".[]") ||
			strings.Contains(lower, "token") || strings.Contains(lower, "secret") ||
			strings.Contains(lower, "cookie") || strings.Contains(lower, "credential") ||
			strings.Contains(lower, "capability") || strings.Contains(lower, "solution") {
			return DomainError{}, errors.New("domain error detail key is unsafe")
		}
		switch typed := value.(type) {
		case string, bool, nil:
			safe[key] = typed
		case int:
			if typed < 0 || uint64(typed) > MaxSafeInteger {
				return DomainError{}, errors.New("domain error numeric detail is outside the JSON safe integer range")
			}
			safe[key] = typed
		case int64:
			if typed < 0 || uint64(typed) > MaxSafeInteger {
				return DomainError{}, errors.New("domain error numeric detail is outside the JSON safe integer range")
			}
			safe[key] = typed
		case uint64:
			if typed > MaxSafeInteger {
				return DomainError{}, errors.New("domain error numeric detail is outside the JSON safe integer range")
			}
			safe[key] = typed
		default:
			return DomainError{}, errors.New("domain error details must contain scalar JSON values only")
		}
	}
	return DomainError{Code: code, Retryable: retryableCodes[code], Details: safe}, nil
}

func (domainError DomainError) Error() string {
	return string(domainError.Code)
}

func AllErrorCodes() []ErrorCode {
	return append([]ErrorCode(nil), allErrorCodes...)
}

func isKnownErrorCode(code ErrorCode) bool {
	for _, candidate := range allErrorCodes {
		if code == candidate {
			return true
		}
	}
	return false
}
