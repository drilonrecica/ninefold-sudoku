package idgen

import (
	"strings"
	"testing"
)

func TestGeneratorProducesDistinctUUIDv7Identifiers(t *testing.T) {
	t.Parallel()
	generator := Generator{}
	first, err := generator.RequestID()
	if err != nil {
		t.Fatal(err)
	}
	second, err := generator.RequestID()
	if err != nil {
		t.Fatal(err)
	}
	if first == second || !strings.Contains(first.String(), "-7") {
		t.Fatalf("unexpected generated identifiers: %s %s", first, second)
	}

	generators := []func() error{
		func() error { _, err := generator.RoomID(); return err },
		func() error { _, err := generator.MatchID(); return err },
		func() error { _, err := generator.ParticipantID(); return err },
		func() error { _, err := generator.PuzzleID(); return err },
		func() error { _, err := generator.ReplayID(); return err },
		func() error { _, err := generator.ConnectionID(); return err },
	}
	for _, generate := range generators {
		if err := generate(); err != nil {
			t.Fatal(err)
		}
	}
}
