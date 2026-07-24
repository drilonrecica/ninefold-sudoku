package engine

import (
	"context"
	"testing"
	"time"

	"pgregory.net/rapid"
)

func TestTransformationProperties(t *testing.T) {
	clues, _ := ParseClues("061000803000200000432000600058900007000507200900100050003001000840005000000000064")
	_, solution := CountSolutions(clues)
	original, err := SolveLogical(clues)
	if err != nil {
		t.Fatal(err)
	}
	rapid.Check(t, func(t *rapid.T) {
		seed := rapid.Uint64().Draw(t, "seed")
		transformedClues, transformedSolution, err := Transform(clues, solution, seed)
		if err != nil {
			t.Fatal(err)
		}
		if err := NewPuzzle(transformedClues, transformedSolution); err != nil {
			t.Fatal(err)
		}
		count, unique := CountSolutions(transformedClues)
		if count != 1 || unique.String() != transformedSolution.String() {
			t.Fatalf("seed %d changed uniqueness", seed)
		}
		logical, err := SolveLogical(transformedClues)
		if err != nil || logical.Grade != original.Grade ||
			logical.Solution.String() != transformedSolution.String() {
			t.Fatalf("seed %d changed deterministic grade or solution: %#v %v", seed, logical, err)
		}
	})
}

func TestGeneratorReplayability(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	first, err := Generate(ctx, 20260301, Easy, 3)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Generate(ctx, 20260301, Easy, 3)
	if err != nil {
		t.Fatal(err)
	}
	if first.Clues.String() != second.Clues.String() || first.Solution.String() != second.Solution.String() ||
		first.Grade != second.Grade || first.Hardest != second.Hardest {
		t.Fatal("fixed seed did not reproduce the generated puzzle")
	}
}

func TestGeneratorHonorsCancellationAndBounds(t *testing.T) {
	t.Parallel()
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := Generate(cancelled, 1, Expert, 1); err == nil {
		t.Fatal("cancelled generation succeeded")
	}
	if _, err := Generate(context.Background(), 1, Expert, 0); err == nil {
		t.Fatal("unbounded generation accepted")
	}
}

func FuzzParseClues(f *testing.F) {
	f.Add("530070000600195000098000060800060003400803001700020006060000280000419005000080079")
	f.Add("")
	f.Fuzz(func(t *testing.T, encoded string) {
		_, _ = ParseClues(encoded)
	})
}

func FuzzLogicalSolver(f *testing.F) {
	f.Add("530070000600195000098000060800060003400803001700020006060000280000419005000080079")
	f.Add("000000000000000000000000000000000000000000000000000000000000000000000000000000000")
	f.Fuzz(func(t *testing.T, encoded string) {
		clues, err := ParseClues(encoded)
		if err != nil {
			return
		}
		_, _ = SolveLogical(clues)
	})
}

func FuzzTransform(f *testing.F) {
	clues, _ := ParseClues("530070000600195000098000060800060003400803001700020006060000280000419005000080079")
	solution, _ := ParseSolution(standardSolution)
	f.Add(uint64(0))
	f.Add(^uint64(0))
	f.Fuzz(func(t *testing.T, seed uint64) {
		_, _, err := Transform(clues, solution, seed)
		if err != nil {
			t.Fatal(err)
		}
	})
}
