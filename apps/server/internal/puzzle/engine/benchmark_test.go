package engine

import (
	"context"
	"testing"
)

var benchmarkClues, _ = ParseClues("000082900000000087409000000005040006003010590690020000006273408040000070708004300")

func BenchmarkValidation(b *testing.B) {
	encoded := benchmarkClues.String()
	for b.Loop() {
		if _, err := ParseClues(encoded); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUniqueness(b *testing.B) {
	for b.Loop() {
		if count, _ := CountSolutions(benchmarkClues); count != 1 {
			b.Fatal(count)
		}
	}
}

func BenchmarkLogicalGrading(b *testing.B) {
	for b.Loop() {
		if _, err := SolveLogical(benchmarkClues); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkHints(b *testing.B) {
	for b.Loop() {
		if _, err := Nudge(benchmarkClues); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCanonicalization(b *testing.B) {
	for b.Loop() {
		if CanonicalFingerprint(benchmarkClues) == "" {
			b.Fatal("empty fingerprint")
		}
	}
}

func BenchmarkGeneration(b *testing.B) {
	for iteration := 0; b.Loop(); iteration++ {
		if _, err := Generate(context.Background(), uint64(iteration+1), Easy, 5); err != nil {
			b.Fatal(err)
		}
	}
}
