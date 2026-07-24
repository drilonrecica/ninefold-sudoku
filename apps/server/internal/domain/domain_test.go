package domain

import (
	"testing"
	"time"
)

const validUUIDv7 = "0190a7d3-8d9a-7f31-8d2a-1242f6f0d101"

func TestUUIDv7Identifiers(t *testing.T) {
	t.Parallel()
	parsers := []struct {
		name  string
		parse func(string) error
	}{
		{"RoomID", func(value string) error { _, err := ParseRoomID(value); return err }},
		{"MatchID", func(value string) error { _, err := ParseMatchID(value); return err }},
		{"ParticipantID", func(value string) error { _, err := ParseParticipantID(value); return err }},
		{"PuzzleID", func(value string) error { _, err := ParsePuzzleID(value); return err }},
		{"ReplayID", func(value string) error { _, err := ParseReplayID(value); return err }},
		{"RequestID", func(value string) error { _, err := ParseRequestID(value); return err }},
		{"ConnectionID", func(value string) error { _, err := ParseConnectionID(value); return err }},
	}
	for _, parser := range parsers {
		parser := parser
		t.Run(parser.name, func(t *testing.T) {
			t.Parallel()
			if err := parser.parse(validUUIDv7); err != nil {
				t.Fatalf("valid UUIDv7 rejected: %v", err)
			}
			for _, invalid := range []string{
				"",
				"0190a7d3-8d9a-4f31-8d2a-1242f6f0d101",
				"0190A7D3-8D9A-7F31-8D2A-1242F6F0D101",
				"0190a7d3-8d9a-7f31-1d2a-1242f6f0d101",
			} {
				if err := parser.parse(invalid); err == nil {
					t.Fatalf("invalid identifier accepted: %q", invalid)
				}
			}
		})
	}
}

func TestPrimitiveBoundaries(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name    string
		valid   []int
		invalid []int
		parse   func(int) error
	}{
		{
			name:    "cell",
			valid:   []int{0, 80},
			invalid: []int{-1, 81},
			parse:   func(value int) error { _, err := NewCellIndex(value); return err },
		},
		{
			name:    "digit",
			valid:   []int{1, 9},
			invalid: []int{0, 10},
			parse:   func(value int) error { _, err := NewDigit(value); return err },
		},
	} {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			for _, value := range test.valid {
				if err := test.parse(value); err != nil {
					t.Fatalf("valid %s %d rejected: %v", test.name, value, err)
				}
			}
			for _, value := range test.invalid {
				if err := test.parse(value); err == nil {
					t.Fatalf("invalid %s %d accepted", test.name, value)
				}
			}
		})
	}

	for _, value := range []string{"7KMP4R", "7kmp4r"} {
		code, err := ParseRoomCode(value)
		if err != nil || code.String() != "7KMP4R" {
			t.Fatalf("ParseRoomCode(%q) = %q, %v", value, code, err)
		}
	}
	for _, value := range []string{"", "7IMP4R", "7KMP4", "7KMP4R1"} {
		if _, err := ParseRoomCode(value); err == nil {
			t.Fatalf("invalid room code accepted: %q", value)
		}
	}

	if _, err := NewRoomVersion(MaxSafeInteger); err != nil {
		t.Fatal(err)
	}
	if _, err := NewRoomVersion(MaxSafeInteger + 1); err == nil {
		t.Fatal("unsafe room version accepted")
	}
	if _, err := NewEventNumber(0); err == nil {
		t.Fatal("zero event number accepted")
	}
	if _, err := NewClientSequence(MaxSafeInteger + 1); err == nil {
		t.Fatal("unsafe client sequence accepted")
	}
}

func TestCandidateSet(t *testing.T) {
	t.Parallel()
	one, _ := NewDigit(1)
	nine, _ := NewDigit(9)
	set, err := NewCandidateSet(one, nine, one)
	if err != nil {
		t.Fatal(err)
	}
	if set.Len() != 2 || !set.Contains(one) || !set.Contains(nine) {
		t.Fatalf("unexpected candidate set: %v", set.Digits())
	}
	set = set.Remove(one)
	if set.Contains(one) || set.Len() != 1 {
		t.Fatalf("candidate removal failed: %v", set.Digits())
	}
	if !AllCandidates().Valid() || AllCandidates().Len() != 9 {
		t.Fatal("all-candidate set is invalid")
	}
}

func TestTimestampUsesUTCMilliseconds(t *testing.T) {
	t.Parallel()
	value := time.Date(2026, 7, 24, 12, 34, 56, 789123000, time.FixedZone("offset", 2*60*60))
	timestamp, err := NewTimestamp(value)
	if err != nil {
		t.Fatal(err)
	}
	if timestamp.Time().Location() != time.UTC || timestamp.Time().Nanosecond()%1_000_000 != 0 {
		t.Fatalf("timestamp not normalized to UTC milliseconds: %s", timestamp.Time())
	}
}

func TestCommandAndEventMetadataBoundaries(t *testing.T) {
	t.Parallel()
	requestID, _ := ParseRequestID(validUUIDv7)
	participantID, _ := ParseParticipantID("0190a7d3-8d9a-7f31-8d2a-1242f6f0d102")
	matchID, _ := ParseMatchID("0190a7d3-8d9a-7f31-8d2a-1242f6f0d103")
	sequence, _ := NewClientSequence(1)
	metadata, err := NewCommandMetadata(requestID, participantID, sequence, NewMatchTarget(matchID), MaxSafeInteger)
	if err != nil || metadata.Target.Kind != AggregateMatch {
		t.Fatalf("valid command metadata rejected: %#v %v", metadata, err)
	}
	if _, err := NewCommandMetadata(requestID, participantID, sequence, NewMatchTarget(matchID), MaxSafeInteger+1); err == nil {
		t.Fatal("unsafe expected version accepted")
	}

	number, _ := NewEventNumber(1)
	occurredAt, _ := TimestampFromMilliseconds(1_720_000_000_000)
	event, err := NewEventMetadata(1, number, 1, NewMatchTarget(matchID), occurredAt)
	if err != nil || event.SchemaVersion != 1 {
		t.Fatalf("valid event metadata rejected: %#v %v", event, err)
	}
	if _, err := NewEventMetadata(0, number, 1, NewMatchTarget(matchID), occurredAt); err == nil {
		t.Fatal("zero schema version accepted")
	}
}
