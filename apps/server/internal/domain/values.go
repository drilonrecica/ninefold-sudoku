package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

const MaxSafeInteger uint64 = 1<<53 - 1

const roomCodeAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

type RoomCode string

func ParseRoomCode(value string) (RoomCode, error) {
	normalized := strings.ToUpper(strings.TrimSpace(value))
	if len(normalized) != 6 {
		return "", errors.New("room code must contain exactly six ASCII characters")
	}
	for _, char := range normalized {
		if !strings.ContainsRune(roomCodeAlphabet, char) {
			return "", errors.New("room code contains an unsupported character")
		}
	}
	return RoomCode(normalized), nil
}

func (code RoomCode) String() string { return string(code) }

type CellIndex uint8

func NewCellIndex(value int) (CellIndex, error) {
	if value < 0 || value > 80 {
		return 0, errors.New("cell index must be between 0 and 80")
	}
	return CellIndex(value), nil
}

func (cell CellIndex) Row() uint8    { return uint8(cell) / 9 }
func (cell CellIndex) Column() uint8 { return uint8(cell) % 9 }
func (cell CellIndex) Box() uint8    { return (cell.Row()/3)*3 + cell.Column()/3 }

type Digit uint8

func NewDigit(value int) (Digit, error) {
	if value < 1 || value > 9 {
		return 0, errors.New("digit must be between 1 and 9")
	}
	return Digit(value), nil
}

type CandidateSet uint16

const allCandidates CandidateSet = 0b11_1111_1110

func NewCandidateSet(digits ...Digit) (CandidateSet, error) {
	var set CandidateSet
	for _, digit := range digits {
		if digit < 1 || digit > 9 {
			return 0, errors.New("candidate digit must be between 1 and 9")
		}
		set = set.Add(digit)
	}
	return set, nil
}

func AllCandidates() CandidateSet { return allCandidates }
func (set CandidateSet) Valid() bool {
	return set&^allCandidates == 0
}
func (set CandidateSet) Contains(digit Digit) bool {
	return digit >= 1 && digit <= 9 && set&(1<<digit) != 0
}
func (set CandidateSet) Add(digit Digit) CandidateSet {
	if digit < 1 || digit > 9 {
		return set
	}
	return set | 1<<digit
}
func (set CandidateSet) Remove(digit Digit) CandidateSet {
	if digit < 1 || digit > 9 {
		return set
	}
	return set &^ (1 << digit)
}
func (set CandidateSet) Len() int {
	count := 0
	for digit := Digit(1); digit <= 9; digit++ {
		if set.Contains(digit) {
			count++
		}
	}
	return count
}
func (set CandidateSet) Digits() []Digit {
	digits := make([]Digit, 0, set.Len())
	for digit := Digit(1); digit <= 9; digit++ {
		if set.Contains(digit) {
			digits = append(digits, digit)
		}
	}
	return digits
}

type (
	RoomVersion    uint64
	MatchVersion   uint64
	EventNumber    uint64
	ClientSequence uint64
)

func NewRoomVersion(value uint64) (RoomVersion, error) {
	if value > MaxSafeInteger {
		return 0, errors.New("room version exceeds the JSON safe integer maximum")
	}
	return RoomVersion(value), nil
}

func NewMatchVersion(value uint64) (MatchVersion, error) {
	if value > MaxSafeInteger {
		return 0, errors.New("match version exceeds the JSON safe integer maximum")
	}
	return MatchVersion(value), nil
}

func NewEventNumber(value uint64) (EventNumber, error) {
	if value == 0 || value > MaxSafeInteger {
		return 0, errors.New("event number must be between 1 and the JSON safe integer maximum")
	}
	return EventNumber(value), nil
}

func NewClientSequence(value uint64) (ClientSequence, error) {
	if value == 0 || value > MaxSafeInteger {
		return 0, errors.New("client sequence must be between 1 and the JSON safe integer maximum")
	}
	return ClientSequence(value), nil
}

type Timestamp struct {
	milliseconds int64
}

func NewTimestamp(value time.Time) (Timestamp, error) {
	if value.IsZero() {
		return Timestamp{}, errors.New("timestamp must not be zero")
	}
	utc := value.UTC()
	if utc.UnixMilli() <= 0 {
		return Timestamp{}, errors.New("timestamp milliseconds must be positive")
	}
	return Timestamp{milliseconds: utc.UnixMilli()}, nil
}

func TimestampFromMilliseconds(value int64) (Timestamp, error) {
	if value <= 0 {
		return Timestamp{}, errors.New("timestamp milliseconds must be positive")
	}
	return Timestamp{milliseconds: value}, nil
}

func (timestamp Timestamp) Time() time.Time {
	return time.UnixMilli(timestamp.milliseconds).UTC()
}

func (timestamp Timestamp) Milliseconds() int64 {
	return timestamp.milliseconds
}

func validateEnum[T ~string](kind string, value T, allowed ...T) (T, error) {
	for _, candidate := range allowed {
		if value == candidate {
			return value, nil
		}
	}
	var zero T
	return zero, fmt.Errorf("unsupported %s %q", kind, value)
}
