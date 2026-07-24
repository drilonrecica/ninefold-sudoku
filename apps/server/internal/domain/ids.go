package domain

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type (
	RoomID        string
	MatchID       string
	ParticipantID string
	PuzzleID      string
	ReplayID      string
	RequestID     string
	ConnectionID  string
)

func parseUUIDv7(kind, value string) (string, error) {
	if value != strings.ToLower(value) {
		return "", fmt.Errorf("%s must use canonical lowercase UUID text", kind)
	}
	parsed, err := uuid.Parse(value)
	if err != nil || parsed.String() != value {
		return "", fmt.Errorf("%s must be a canonical UUID", kind)
	}
	if parsed.Version() != 7 {
		return "", fmt.Errorf("%s must be UUIDv7", kind)
	}
	if parsed.Variant() != uuid.RFC4122 {
		return "", fmt.Errorf("%s must use the RFC 4122 variant", kind)
	}
	return value, nil
}

func ParseRoomID(value string) (RoomID, error) {
	parsed, err := parseUUIDv7("RoomID", value)
	return RoomID(parsed), err
}

func ParseMatchID(value string) (MatchID, error) {
	parsed, err := parseUUIDv7("MatchID", value)
	return MatchID(parsed), err
}

func ParseParticipantID(value string) (ParticipantID, error) {
	parsed, err := parseUUIDv7("ParticipantID", value)
	return ParticipantID(parsed), err
}

func ParsePuzzleID(value string) (PuzzleID, error) {
	parsed, err := parseUUIDv7("PuzzleID", value)
	return PuzzleID(parsed), err
}

func ParseReplayID(value string) (ReplayID, error) {
	parsed, err := parseUUIDv7("ReplayID", value)
	return ReplayID(parsed), err
}

func ParseRequestID(value string) (RequestID, error) {
	parsed, err := parseUUIDv7("RequestID", value)
	return RequestID(parsed), err
}

func ParseConnectionID(value string) (ConnectionID, error) {
	parsed, err := parseUUIDv7("ConnectionID", value)
	return ConnectionID(parsed), err
}

func (id RoomID) String() string        { return string(id) }
func (id MatchID) String() string       { return string(id) }
func (id ParticipantID) String() string { return string(id) }
func (id PuzzleID) String() string      { return string(id) }
func (id ReplayID) String() string      { return string(id) }
func (id RequestID) String() string     { return string(id) }
func (id ConnectionID) String() string  { return string(id) }
