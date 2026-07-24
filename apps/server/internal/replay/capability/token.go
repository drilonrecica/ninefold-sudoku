package capability

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

const tokenBytes = 32

// Token is an opaque replay read capability. Only Hash is persisted.
type Token struct {
	Value string
	Hash  []byte
}

func Generate() (Token, error) {
	raw := make([]byte, tokenBytes)
	if _, err := rand.Read(raw); err != nil {
		return Token{}, err
	}
	value := base64.RawURLEncoding.EncodeToString(raw)
	return Token{Value: value, Hash: Hash(value)}, nil
}

func Hash(value string) []byte {
	hash := sha256.Sum256([]byte(value))
	return hash[:]
}
