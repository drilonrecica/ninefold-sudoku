package proof

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestProofVersionOneFixture(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "..", "contracts", "fixtures", "replay-proof-v1.json"))
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Events []struct {
			Canonical string   `json:"canonical"`
			Envelope  Envelope `json:"envelope"`
			EventHash string   `json:"eventHash"`
		} `json:"events"`
		HiddenCommitment struct {
			CanonicalPrivatePayload string `json:"canonicalPrivatePayload"`
			Salt                    string `json:"salt"`
			Digest                  string `json:"digest"`
		} `json:"hiddenCommitment"`
		Signing struct {
			FinalHash string `json:"finalHash"`
			PublicKey string `json:"publicKey"`
			Signature string `json:"signature"`
		} `json:"signing"`
	}
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	for _, event := range fixture.Events {
		canonical, err := CanonicalEnvelope(event.Envelope)
		if err != nil || string(canonical) != event.Canonical {
			t.Fatalf("canonical mismatch: %v\n%s", err, canonical)
		}
		hash, err := HashEnvelope(event.Envelope)
		if err != nil || Hex(hash) != event.EventHash {
			t.Fatalf("hash mismatch: %v %x", err, hash)
		}
	}
	salt, _ := hex.DecodeString(fixture.HiddenCommitment.Salt)
	digest, _, err := PrivateCommitmentWithSalt(json.RawMessage(fixture.HiddenCommitment.CanonicalPrivatePayload), salt)
	if err != nil || Hex(digest) != fixture.HiddenCommitment.Digest {
		t.Fatalf("commitment mismatch: %v %x", err, digest)
	}
	publicKey, _ := base64.StdEncoding.DecodeString(fixture.Signing.PublicKey)
	signature, _ := base64.StdEncoding.DecodeString(fixture.Signing.Signature)
	finalHash, _ := hex.DecodeString(fixture.Signing.FinalHash)
	if !ed25519.Verify(publicKey, finalHash, signature) {
		t.Fatal("fixture signature did not verify")
	}
	signature[0] ^= 1
	if ed25519.Verify(publicKey, finalHash, signature) {
		t.Fatal("mutated signature verified")
	}
}
