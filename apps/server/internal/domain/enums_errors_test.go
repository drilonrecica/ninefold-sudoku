package domain

import "testing"

func TestCurrentEnums(t *testing.T) {
	t.Parallel()
	parsers := []struct {
		name    string
		valid   []string
		invalid string
		parse   func(string) error
	}{
		{"mode", []string{"Coop", "Solo"}, "Race", func(value string) error { _, err := ParseMode(value); return err }},
		{"difficulty", []string{"Easy", "Medium", "Hard", "Expert", "Random"}, "Impossible", func(value string) error { _, err := ParseDifficulty(value); return err }},
		{"role", []string{"Player", "Spectator"}, "Host", func(value string) error { _, err := ParseParticipationRole(value); return err }},
		{"room state", []string{"Lobby", "Countdown", "InMatch", "Results", "Expired", "Cancelled", "RecoveryPending", "TerminatedByAdmin"}, "Unknown", func(value string) error { _, err := ParseRoomState(value); return err }},
		{"match state", []string{"Prepared", "Countdown", "Active", "Completed", "RecoveryPending", "Cancelled", "Abandoned"}, "Finishing", func(value string) error { _, err := ParseMatchState(value); return err }},
		{"puzzle state", []string{"Draft", "Verified", "Active", "Retired"}, "Deleted", func(value string) error { _, err := ParsePuzzleState(value); return err }},
		{"error preset", []string{"Casual", "Challenge", "Blind", "Clean"}, "Duel", func(value string) error { _, err := ParseErrorPreset(value); return err }},
		{"hint", []string{"Nudge", "Reveal"}, "Explain", func(value string) error { _, err := ParseHintLevel(value); return err }},
	}
	for _, parser := range parsers {
		parser := parser
		t.Run(parser.name, func(t *testing.T) {
			t.Parallel()
			for _, value := range parser.valid {
				if err := parser.parse(value); err != nil {
					t.Fatalf("valid value %q rejected: %v", value, err)
				}
			}
			if err := parser.parse(parser.invalid); err == nil {
				t.Fatalf("deferred or invalid value %q accepted", parser.invalid)
			}
		})
	}
}

func TestEveryCurrentErrorCode(t *testing.T) {
	t.Parallel()
	if len(AllErrorCodes()) != 65 {
		t.Fatalf("error catalog changed without updating its boundary test: %d", len(AllErrorCodes()))
	}
	for _, code := range AllErrorCodes() {
		domainError, err := NewDomainError(code, map[string]any{"currentVersion": uint64(1)})
		if err != nil || domainError.Code != code {
			t.Fatalf("error code %q rejected: %v", code, err)
		}
	}
	for _, details := range []map[string]any{
		{"sessionToken": "secret"},
		{"nested": map[string]string{"unsafe": "value"}},
		{"unsafe": MaxSafeInteger + 1},
	} {
		if _, err := NewDomainError(ErrRoomFull, details); err == nil {
			t.Fatalf("unsafe details accepted: %#v", details)
		}
	}
	if _, err := NewDomainError("UNKNOWN", nil); err == nil {
		t.Fatal("unknown error code accepted")
	}
}
