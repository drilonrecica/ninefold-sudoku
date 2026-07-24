package main

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"flag"
	"log/slog"
	"os"
	"strings"

	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/platform/idgen"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/puzzle/catalog"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/puzzle/engine"
)

func main() {
	var (
		output       = flag.String("output", "", "required JSONL output path")
		validate     = flag.String("validate", "", "validate an existing JSONL catalog")
		grade        = flag.String("grade", "Easy", "target grade")
		seed         = flag.Uint64("seed", 0, "SplitMix64 seed; cryptographically random when omitted")
		maxAttempts  = flag.Int("max-attempts", 100, "bounded generation attempts")
		approved     = flag.Bool("multiplayer-approved", false, "record explicit reviewed approval")
		reviewReason = flag.String("review-reason", "", "required reason for multiplayer approval")
	)
	flag.Parse()
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	if *validate != "" {
		records, err := catalog.ReadFile(*validate)
		if err != nil {
			logger.Error("catalog validation failed", "error", err)
			os.Exit(1)
		}
		logger.Info("catalog valid", "records", len(records))
		return
	}
	if *output == "" {
		logger.Error("generation requires an explicit -output path")
		os.Exit(2)
	}
	if *approved && strings.TrimSpace(*reviewReason) == "" {
		logger.Error("multiplayer approval requires -review-reason")
		os.Exit(2)
	}
	var target engine.Grade
	switch strings.ToLower(*grade) {
	case "easy":
		target = engine.Easy
	case "medium":
		target = engine.Medium
	case "hard":
		target = engine.Hard
	case "expert":
		target = engine.Expert
	default:
		logger.Error("grade must be Easy, Medium, Hard, or Expert")
		os.Exit(2)
	}
	if *seed == 0 {
		var bytes [8]byte
		if _, err := rand.Read(bytes[:]); err != nil {
			logger.Error("seed generation failed", "error", err)
			os.Exit(1)
		}
		*seed = binary.LittleEndian.Uint64(bytes[:])
	}
	generated, err := engine.Generate(context.Background(), *seed, target, *maxAttempts)
	if err != nil {
		logger.Error("puzzle generation failed", "grade", target, "seed", *seed, "error", err)
		os.Exit(1)
	}
	puzzleID, err := (idgen.Generator{}).PuzzleID()
	if err != nil {
		logger.Error("identifier generation failed", "error", err)
		os.Exit(1)
	}
	record := catalog.Record{
		ID: puzzleID.String(), Revision: 1, Lifecycle: "Verified",
		Clues: generated.Clues.String(), Solution: generated.Solution.String(),
		Difficulty: generated.Grade, HardestTechnique: generated.Hardest.String(),
		CanonicalFingerprint: engine.CanonicalFingerprint(generated.Clues),
		CanonicalVersion:     engine.CanonicalVersion, TransformationVersion: engine.TransformationVersion,
		GeneratorVersion: engine.GeneratorVersion, SolverVersion: engine.SolverVersion,
		GraderVersion: engine.GraderVersion, Seed: generated.Seed, Attempt: generated.Attempt,
		SolverPath:        catalog.SolverPath(generated.Steps),
		Quality:           generated.Metrics,
		MultiplayerReview: catalog.Review{Approved: *approved, Reason: strings.TrimSpace(*reviewReason)},
	}
	file, err := os.OpenFile(*output, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		logger.Error("output creation failed", "error", err)
		os.Exit(1)
	}
	if err := json.NewEncoder(file).Encode(record); err != nil {
		_ = file.Close()
		logger.Error("catalog write failed", "error", err)
		os.Exit(1)
	}
	if err := file.Close(); err != nil {
		logger.Error("catalog close failed", "error", err)
		os.Exit(1)
	}
	logger.Info("puzzle generated", "id", puzzleID.String(), "grade", generated.Grade, "seed", *seed, "attempt", generated.Attempt, "output", *output)
}
