package proof

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"

	shared "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/domain"
)

const Version = 1

type Claims struct {
	Version            int    `json:"v"`
	AttemptID          string `json:"a"`
	PuzzleID           string `json:"p"`
	Revision           int64  `json:"r"`
	TransformationSeed uint64 `json:"t"`
	IssuedAtMs         int64  `json:"i"`
	PlayStyle          string `json:"s"`
}

func Sign(secret []byte, claims Claims) (string, error) {
	if len(secret) < 32 || claims.Version != Version || claims.AttemptID == "" || claims.PuzzleID == "" {
		return "", errors.New("invalid assignment proof input")
	}
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	encoded := base64.RawURLEncoding.EncodeToString(payload)
	return encoded + "." + base64.RawURLEncoding.EncodeToString(mac(secret, encoded)), nil
}

func Verify(secret []byte, token string) (Claims, error) {
	var claims Claims
	encoded, signature, ok := strings.Cut(token, ".")
	if !ok || encoded == "" || signature == "" || strings.Contains(signature, ".") {
		return claims, errors.New("malformed assignment proof")
	}
	provided, err := base64.RawURLEncoding.DecodeString(signature)
	if err != nil || !hmac.Equal(provided, mac(secret, encoded)) {
		return claims, errors.New("invalid assignment proof signature")
	}
	payload, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil || len(payload) > 1024 {
		return claims, errors.New("invalid assignment proof payload")
	}
	decoder := json.NewDecoder(strings.NewReader(string(payload)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&claims); err != nil || claims.Version != Version {
		return Claims{}, errors.New("unsupported assignment proof")
	}
	if _, err := shared.ParseRequestID(claims.AttemptID); err != nil {
		return Claims{}, errors.New("invalid attempt identifier")
	}
	if _, err := shared.ParsePuzzleID(claims.PuzzleID); err != nil || claims.Revision <= 0 ||
		(claims.PlayStyle != "Guided" && claims.PlayStyle != "Classic") {
		return Claims{}, errors.New("invalid assignment claims")
	}
	return claims, nil
}

func mac(secret []byte, encoded string) []byte {
	hash := hmac.New(sha256.New, secret)
	_, _ = hash.Write([]byte("ninefold-solo-assignment-v1\x00"))
	_, _ = hash.Write([]byte(encoded))
	return hash.Sum(nil)
}
