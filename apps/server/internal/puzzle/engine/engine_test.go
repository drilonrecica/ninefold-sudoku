package engine

import (
	"testing"
)

const standardSolution = "534678912672195348198342567859761423426853791713924856961537284287419635345286179"

func TestGridValidationAndUniqueness(t *testing.T) {
	t.Parallel()
	clues, err := ParseClues("530070000600195000098000060800060003400803001700020006060000280000419005000080079")
	if err != nil {
		t.Fatal(err)
	}
	solution, err := ParseSolution(standardSolution)
	if err != nil {
		t.Fatal(err)
	}
	if err := NewPuzzle(clues, solution); err != nil {
		t.Fatal(err)
	}
	count, found := CountSolutions(clues)
	if count != 1 || found.String() != solution.String() {
		t.Fatalf("solutions=%d found=%s", count, found.String())
	}
	empty, _ := ParseClues("000000000000000000000000000000000000000000000000000000000000000000000000000000000")
	if count, _ := CountSolutions(empty); count != 2 {
		t.Fatalf("solution counting did not stop after non-uniqueness: %d", count)
	}
	if candidates := Candidates(clues, 2); candidates != (1<<1 | 1<<2 | 1<<4) {
		t.Fatalf("unexpected candidates for cell 2: %09b", candidates)
	}
	disagreement := solution
	disagreement.cells[0], disagreement.cells[1] = disagreement.cells[1], disagreement.cells[0]
	if err := NewPuzzle(clues, disagreement); err == nil {
		t.Fatal("clue/solution disagreement accepted")
	}
	for _, invalid := range []string{"", "53007000x" + string(make([]byte, 72)), "113456789" + standardSolution[9:]} {
		if _, err := ParseClues(invalid); err == nil {
			t.Fatalf("invalid clues accepted: %q", invalid)
		}
	}
}

func TestLogicalSolverRejectsPuzzleBeyondTechniqueLimit(t *testing.T) {
	clues, err := ParseClues("005300000800000020070010500400005300010070006003200080060500009004000030000009700")
	if err != nil {
		t.Fatal(err)
	}
	if count, _ := CountSolutions(clues); count != 1 {
		t.Fatalf("solution count = %d", count)
	}
	if _, err := SolveLogical(clues); err == nil {
		t.Fatal("puzzle requiring search was assigned a logical grade")
	}
}

func TestTechniqueOrderAndGradeBoundaries(t *testing.T) {
	expected := []struct {
		technique Technique
		name      string
		grade     Grade
	}{
		{NakedSingle, "naked_single", Easy},
		{HiddenSingle, "hidden_single", Easy},
		{LockedCandidates, "locked_candidates", Medium},
		{NakedPair, "naked_pair", Medium},
		{HiddenPair, "hidden_pair", Medium},
		{NakedTriple, "naked_triple", Hard},
		{HiddenTriple, "hidden_triple", Hard},
		{XWing, "x_wing", Hard},
		{Swordfish, "swordfish", Expert},
		{XYWing, "xy_wing", Expert},
		{SimpleColoring, "simple_coloring", Expert},
	}
	for index, item := range expected {
		if int(item.technique) != index+1 || item.technique.String() != item.name ||
			gradeFor(item.technique) != item.grade {
			t.Fatalf("technique boundary %d is inconsistent: %#v", index+1, item)
		}
	}
}

func TestHintsAndTransform(t *testing.T) {
	t.Parallel()
	clues, _ := ParseClues("530070000600195000098000060800060003400803001700020006060000280000419005000080079")
	solution, _ := ParseSolution(standardSolution)
	nudge, err := Nudge(clues)
	if err != nil || len(nudge.AffectedCells) == 0 {
		t.Fatalf("invalid nudge: %#v %v", nudge, err)
	}
	reveal, err := Reveal(clues, solution)
	if err != nil || clues.cells[reveal.Cell] != 0 || solution.cells[reveal.Cell] != reveal.Value {
		t.Fatalf("invalid reveal: %#v %v", reveal, err)
	}
	firstClues, firstSolution, err := Transform(clues, solution, 42)
	if err != nil {
		t.Fatal(err)
	}
	secondClues, secondSolution, _ := Transform(clues, solution, 42)
	if firstClues.String() != secondClues.String() || firstSolution.String() != secondSolution.String() {
		t.Fatal("transformation is not deterministic")
	}
	if count, _ := CountSolutions(firstClues); count != 1 {
		t.Fatalf("transformation changed uniqueness: %d", count)
	}
	if CanonicalFingerprint(clues) != CanonicalFingerprint(firstClues) {
		t.Fatal("transformation changed canonical identity")
	}
}
