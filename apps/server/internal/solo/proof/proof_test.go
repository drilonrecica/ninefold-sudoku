package proof

import (
	"strings"
	"testing"
)

func TestAssignmentProofRoundTripAndTamper(t *testing.T) {
	secret := []byte(strings.Repeat("s", 32))
	claims := Claims{
		Version: 1, AttemptID: "01900000-0000-7000-8000-000000000001",
		PuzzleID: "01900000-0000-7000-8000-000000000002", Revision: 3,
		TransformationSeed: 42, IssuedAtMs: 1767225600000, PlayStyle: "Guided",
	}
	token, err := Sign(secret, claims)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Verify(secret, token)
	if err != nil || got != claims {
		t.Fatalf("Verify() = %+v, %v", got, err)
	}
	for _, mutation := range []string{
		"A" + token[1:],
		token[:len(token)-1] + "A",
		token + ".extra",
		"not-a-proof",
	} {
		if _, err := Verify(secret, mutation); err == nil {
			t.Fatalf("accepted mutated proof %q", mutation)
		}
	}
}
