package capability

import (
	"bytes"
	"encoding/base64"
	"testing"
)

func TestGenerateUsesOpaque256BitTokenAndStableHash(t *testing.T) {
	first, err := Generate()
	if err != nil {
		t.Fatal(err)
	}
	second, err := Generate()
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := base64.RawURLEncoding.DecodeString(first.Value)
	if err != nil || len(decoded) != tokenBytes {
		t.Fatalf("expected %d random bytes, got %d: %v", tokenBytes, len(decoded), err)
	}
	if first.Value == second.Value || bytes.Equal(first.Hash, second.Hash) {
		t.Fatal("generated capabilities must be unique")
	}
	if !bytes.Equal(first.Hash, Hash(first.Value)) {
		t.Fatal("stored hash does not match capability")
	}
}
