package domain

import shared "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/domain"

type Puzzle struct {
	ID                   shared.PuzzleID
	Revision             uint32
	State                shared.PuzzleState
	Difficulty           shared.Difficulty
	HardestTechnique     string
	QualityScore         uint16
	MultiplayerApproved  bool
	GeneratorVersion     string
	SolverVersion        string
	CanonicalFingerprint string
}

type CatalogSelector interface {
	Select(difficulty shared.Difficulty, excludedFingerprints map[string]struct{}) (Puzzle, error)
}
