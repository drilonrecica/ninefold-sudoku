package domain

// AssignedPuzzle identifies the puzzle and transformation used for a Match.
// Clues and Solution are 81-byte grids; they are server-only and never sent to clients.
type AssignedPuzzle struct {
	PuzzleID           PuzzleID
	Revision           uint32
	TransformationSeed uint64
	Difficulty         Difficulty
	GeneratorVersion   string
	SolverVersion      string
	Clues              []byte
	Solution           []byte
}
