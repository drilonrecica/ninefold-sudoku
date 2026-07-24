package domain

import (
	"crypto/rand"
	"math/big"

	shared "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/domain"
)

const codeAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

// GenerateCode creates a new cryptographically random six-character room code
// using the canonical unambiguous alphabet.
func GenerateCode() (shared.RoomCode, error) {
	max := big.NewInt(int64(len(codeAlphabet)))
	code := make([]byte, 6)
	for i := range code {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		code[i] = codeAlphabet[n.Int64()]
	}
	return shared.RoomCode(code), nil
}
