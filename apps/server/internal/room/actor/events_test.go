package actor

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestReplayHashFixture(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "..", "contracts", "fixtures", "replay-hash-chain.json"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	var fixture struct {
		Envelope  matchEventPublicEnvelope `json:"envelope"`
		EventHash string                   `json:"eventHash"`
	}
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatalf("decode fixture: %v", err)
	}
	got, err := hashEnvelope(fixture.Envelope)
	if err != nil {
		t.Fatalf("hash envelope: %v", err)
	}
	if hex.EncodeToString(got) != fixture.EventHash {
		t.Fatalf("hash mismatch: got %x want %s", got, fixture.EventHash)
	}
}
