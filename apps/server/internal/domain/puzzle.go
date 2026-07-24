package domain

// AssignedPuzzle identifies the puzzle and transformation used for a Match.
type AssignedPuzzle struct {
	PuzzleID           PuzzleID
	Revision           uint32
	TransformationSeed uint64
	Difficulty         Difficulty
	GeneratorVersion   string
	SolverVersion      string
}
