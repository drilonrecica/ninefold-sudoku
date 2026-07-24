package proof

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"

	"github.com/gowebpki/jcs"
)

const Version = 1

var GenesisHash = make([]byte, sha256.Size)

type Envelope struct {
	ProofVersion         int             `json:"proofVersion"`
	MatchID              string          `json:"matchId"`
	EventNumber          uint64          `json:"eventNumber"`
	AggregateVersion     uint64          `json:"aggregateVersion"`
	PublicEventType      string          `json:"publicEventType"`
	PublicActorID        string          `json:"publicActorId"`
	OccurredAtMs         int64           `json:"occurredAtMs"`
	PublicPayload        json.RawMessage `json:"publicPayload"`
	PrivatePayloadDigest string          `json:"privatePayloadDigest"`
	PreviousEventHash    []byte          `json:"previousEventHash"`
}

type Signer struct {
	KeyID      string
	PrivateKey ed25519.PrivateKey
}

func (s Signer) Valid() bool {
	return s.KeyID != "" && len(s.PrivateKey) == ed25519.PrivateKeySize
}

func (s Signer) Sign(finalHash []byte) ([]byte, error) {
	if !s.Valid() || len(finalHash) != sha256.Size {
		return nil, errors.New("invalid replay signer or final hash")
	}
	return ed25519.Sign(s.PrivateKey, finalHash), nil
}

func Canonicalize(value json.RawMessage) ([]byte, error) {
	return jcs.Transform(value)
}

func CanonicalEnvelope(envelope Envelope) ([]byte, error) {
	raw, err := json.Marshal(envelope)
	if err != nil {
		return nil, err
	}
	return Canonicalize(raw)
}

func HashEnvelope(envelope Envelope) ([]byte, error) {
	canonical, err := CanonicalEnvelope(envelope)
	if err != nil {
		return nil, err
	}
	hash := sha256.Sum256(canonical)
	return hash[:], nil
}

func PrivateCommitment(privatePayload json.RawMessage) (digest, salt []byte, err error) {
	salt = make([]byte, 16)
	if _, err = rand.Read(salt); err != nil {
		return nil, nil, err
	}
	return PrivateCommitmentWithSalt(privatePayload, salt)
}

func PrivateCommitmentWithSalt(privatePayload json.RawMessage, salt []byte) (digest, copiedSalt []byte, err error) {
	if len(salt) < 16 {
		return nil, nil, errors.New("private commitment salt must be at least 16 bytes")
	}
	canonical, err := Canonicalize(privatePayload)
	if err != nil {
		return nil, nil, err
	}
	hash := sha256.New()
	_, _ = hash.Write(salt)
	_, _ = hash.Write(canonical)
	return hash.Sum(nil), append([]byte(nil), salt...), nil
}

func Hex(value []byte) string {
	return hex.EncodeToString(value)
}
