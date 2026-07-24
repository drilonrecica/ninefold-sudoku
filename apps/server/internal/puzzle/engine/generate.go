package engine

import (
	"context"
	"errors"
	"fmt"
	"math/bits"
)

const (
	GeneratorVersion = "generator-v1"
	SolverVersion    = "logical-v1"
	GraderVersion    = "grader-v1"
)

type QualityMetrics struct {
	ClueCount         int     `json:"clueCount"`
	LogicalStepCount  int     `json:"logicalStepCount"`
	SingleStepRatio   float64 `json:"singleStepRatio"`
	OpeningCandidates int     `json:"openingCandidates"`
}

type GeneratedPuzzle struct {
	Clues    Clues
	Solution Solution
	Grade    Grade
	Hardest  Technique
	Steps    []Step
	Metrics  QualityMetrics
	Seed     uint64
	Attempt  int
}

func Generate(ctx context.Context, seed uint64, target Grade, maxAttempts int) (GeneratedPuzzle, error) {
	if maxAttempts <= 0 {
		return GeneratedPuzzle{}, errors.New("maximum attempts must be positive")
	}
	if target != Easy && target != Medium && target != Hard && target != Expert {
		return GeneratedPuzzle{}, errors.New("target grade is invalid")
	}
	random := splitMix64{state: seed}
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return GeneratedPuzzle{}, err
		}
		solution, ok := randomizedSolution(ctx, &random)
		if !ok {
			continue
		}
		clues := Clues{cells: solution.cells}
		order := make([]int, CellCount)
		for index := range order {
			order[index] = index
		}
		shuffle(order, &random)
		for _, cell := range order {
			if err := ctx.Err(); err != nil {
				return GeneratedPuzzle{}, err
			}
			previous := clues.cells[cell]
			clues.cells[cell] = 0
			count, _ := CountSolutions(clues)
			if count != 1 {
				clues.cells[cell] = previous
				continue
			}
			logical, err := SolveLogical(clues)
			if err != nil {
				continue
			}
			if logical.Grade == target && clueCount(clues) <= maximumClues(target) {
				return GeneratedPuzzle{
					Clues: clues, Solution: solution, Grade: logical.Grade,
					Hardest: logical.HardestTechnique, Steps: logical.Steps,
					Metrics: qualityMetrics(clues, logical.Steps), Seed: seed, Attempt: attempt,
				}, nil
			}
		}
	}
	return GeneratedPuzzle{}, fmt.Errorf("failed to generate %s puzzle within %d attempts", target, maxAttempts)
}

func randomizedSolution(ctx context.Context, random *splitMix64) (Solution, bool) {
	var cells [CellCount]uint8
	if !fillRandom(ctx, &cells, random) {
		return Solution{}, false
	}
	return Solution{cells: cells}, true
}

func fillRandom(ctx context.Context, cells *[CellCount]uint8, random *splitMix64) bool {
	if ctx.Err() != nil {
		return false
	}
	best, mask, count := -1, uint16(0), 10
	for cell, value := range cells {
		if value != 0 {
			continue
		}
		candidates := candidatesFor(*cells, cell)
		size := bits.OnesCount16(candidates)
		if size == 0 {
			return false
		}
		if size < count {
			best, mask, count = cell, candidates, size
		}
	}
	if best == -1 {
		return true
	}
	var digits []uint8
	for digit := uint8(1); digit <= 9; digit++ {
		if mask&(1<<digit) != 0 {
			digits = append(digits, digit)
		}
	}
	shuffle(digits, random)
	for _, digit := range digits {
		cells[best] = digit
		if fillRandom(ctx, cells, random) {
			return true
		}
		cells[best] = 0
	}
	return false
}

func qualityMetrics(clues Clues, steps []Step) QualityMetrics {
	singles := 0
	for _, step := range steps {
		if step.Technique <= HiddenSingle {
			singles++
		}
	}
	opening := 0
	for cell, value := range clues.cells {
		if value == 0 {
			opening += bits.OnesCount16(candidatesFor(clues.cells, cell))
		}
	}
	ratio := 0.0
	if len(steps) > 0 {
		ratio = float64(singles) / float64(len(steps))
	}
	return QualityMetrics{clueCount(clues), len(steps), ratio, opening}
}

func EvaluateQuality(clues Clues, steps []Step) QualityMetrics {
	return qualityMetrics(clues, steps)
}

func clueCount(clues Clues) int {
	count := 0
	for _, value := range clues.cells {
		if value != 0 {
			count++
		}
	}
	return count
}

func maximumClues(grade Grade) int {
	switch grade {
	case Easy:
		return 40
	case Medium:
		return 36
	case Hard:
		return 33
	default:
		return 30
	}
}
