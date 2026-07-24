package catalog

import (
	"bytes"
	"context"
	"os"
	"testing"

	shared "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/domain"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/puzzle/engine"
)

func TestCommittedCatalog(t *testing.T) {
	records, err := ReadFile("catalog.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 4 {
		t.Fatalf("catalog record count = %d", len(records))
	}
	for _, difficulty := range []shared.Difficulty{
		shared.DifficultyEasy,
		shared.DifficultyMedium,
		shared.DifficultyHard,
		shared.DifficultyExpert,
	} {
		puzzle, err := NewSelector(records).Select(difficulty, nil)
		if err != nil || puzzle.Difficulty != difficulty {
			t.Fatalf("select %s: %#v %v", difficulty, puzzle, err)
		}
		excluded := map[string]struct{}{puzzle.CanonicalFingerprint: {}}
		if _, err := NewSelector(records).Select(difficulty, excluded); err == nil {
			t.Fatalf("excluded %s puzzle was selected", difficulty)
		}
	}
	for _, record := range records {
		generated, err := engine.Generate(context.Background(), record.Seed, record.Difficulty, record.Attempt)
		if err != nil {
			t.Fatalf("reproduce %s: %v", record.ID, err)
		}
		if generated.Attempt != record.Attempt || generated.Clues.String() != record.Clues ||
			generated.Solution.String() != record.Solution {
			t.Fatalf("fixed seed did not reproduce catalog record %s", record.ID)
		}
	}
}

func TestReviewedApprovalRequiresReason(t *testing.T) {
	records, err := ReadFile("catalog.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	record := records[0]
	record.MultiplayerReview.Approved = true
	record.MultiplayerReview.Reason = ""
	if err := ValidateRecord(record); err == nil {
		t.Fatal("approval without reason accepted")
	}
}

func TestCatalogRejectsCanonicalDuplicates(t *testing.T) {
	data, err := os.ReadFile("catalog.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	firstLine := data[:bytes.IndexByte(data, '\n')+1]
	duplicate := append(append([]byte(nil), firstLine...), firstLine...)
	if _, err := Read(bytes.NewReader(duplicate)); err == nil {
		t.Fatal("canonical-equivalent duplicate accepted")
	}
}

func FuzzCatalogInput(f *testing.F) {
	f.Add([]byte("{"))
	f.Add([]byte("{}\n"))
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 2<<20 {
			t.Skip()
		}
		_, _ = Read(bytes.NewReader(data))
	})
}
