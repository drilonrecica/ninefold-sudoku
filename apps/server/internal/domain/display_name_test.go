package domain

import "testing"

func TestDisplayNameNormalization(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		input      string
		display    string
		comparison string
		valid      bool
	}{
		{name: "unicode trim and NFC", input: " \tE\u0301va  ", display: "Éva", comparison: "éva", valid: true},
		{name: "comparison whitespace", input: "Alex\u00a0  Smith", display: "Alex\u00a0  Smith", comparison: "alex smith", valid: true},
		{name: "compatibility fold", input: "Ａlex", display: "Ａlex", comparison: "alex", valid: true},
		{name: "emoji ZWJ", input: "Jo 👨‍👩‍👧‍👦", display: "Jo 👨‍👩‍👧‍👦", comparison: "jo 👨‍👩‍👧‍👦", valid: true},
		{name: "one grapheme", input: "A", valid: false},
		{name: "too long", input: "123456789012345678901", valid: false},
		{name: "control", input: "Al\nex", valid: false},
		{name: "bidi control", input: "Al\u202eex", valid: false},
		{name: "only whitespace", input: "   ", valid: false},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			name, err := NewDisplayName(test.input)
			if !test.valid {
				if err == nil {
					t.Fatalf("invalid display name accepted: %q", test.input)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if name.String() != test.display || name.ComparisonKey() != test.comparison {
				t.Fatalf("got display=%q comparison=%q", name.String(), name.ComparisonKey())
			}
		})
	}
}
