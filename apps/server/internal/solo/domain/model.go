package domain

import shared "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/domain"

type Assignment struct {
	PuzzleID           shared.PuzzleID
	Revision           uint32
	Difficulty         shared.Difficulty
	TransformationSeed uint64
	IssuedAt           shared.Timestamp
}
