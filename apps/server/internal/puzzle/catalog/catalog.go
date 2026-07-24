package catalog

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	shared "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/domain"
	puzzledomain "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/puzzle/domain"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/puzzle/engine"
)

type Review struct {
	Approved bool   `json:"approved"`
	Reason   string `json:"reason,omitempty"`
}

type PathSegment struct {
	Technique string `json:"technique"`
	Count     int    `json:"count"`
}

type Record struct {
	ID                    string                `json:"id"`
	Revision              uint32                `json:"revision"`
	Lifecycle             string                `json:"lifecycle"`
	Clues                 string                `json:"clues"`
	Solution              string                `json:"solution"`
	Difficulty            engine.Grade          `json:"difficulty"`
	HardestTechnique      string                `json:"hardestTechnique"`
	CanonicalFingerprint  string                `json:"canonicalFingerprint"`
	CanonicalVersion      string                `json:"canonicalVersion"`
	TransformationVersion string                `json:"transformationVersion"`
	GeneratorVersion      string                `json:"generatorVersion"`
	SolverVersion         string                `json:"solverVersion"`
	GraderVersion         string                `json:"graderVersion"`
	Seed                  uint64                `json:"seed"`
	Attempt               int                   `json:"attempt"`
	SolverPath            []PathSegment         `json:"solverPath"`
	Quality               engine.QualityMetrics `json:"quality"`
	MultiplayerReview     Review                `json:"multiplayerReview"`
}

func Read(reader io.Reader) ([]Record, error) {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	var records []Record
	identifiers := make(map[string]struct{})
	fingerprints := make(map[string]struct{})
	line := 0
	for scanner.Scan() {
		line++
		if strings.TrimSpace(scanner.Text()) == "" {
			continue
		}
		var record Record
		decoder := json.NewDecoder(strings.NewReader(scanner.Text()))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&record); err != nil {
			return nil, fmt.Errorf("catalog line %d: %w", line, err)
		}
		if err := ValidateRecord(record); err != nil {
			return nil, fmt.Errorf("catalog line %d: %w", line, err)
		}
		if _, duplicate := identifiers[record.ID]; duplicate {
			return nil, fmt.Errorf("catalog line %d: duplicate puzzle identifier", line)
		}
		if _, duplicate := fingerprints[record.CanonicalFingerprint]; duplicate {
			return nil, fmt.Errorf("catalog line %d: canonical-equivalent duplicate", line)
		}
		identifiers[record.ID] = struct{}{}
		fingerprints[record.CanonicalFingerprint] = struct{}{}
		records = append(records, record)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return records, nil
}

func ReadFile(path string) ([]Record, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return Read(file)
}

func ValidateRecord(record Record) error {
	if _, err := shared.ParsePuzzleID(record.ID); err != nil {
		return err
	}
	if record.Revision == 0 || (record.Lifecycle != "Verified" && record.Lifecycle != "Active") {
		return errors.New("revision or lifecycle is invalid")
	}
	if record.MultiplayerReview.Approved && strings.TrimSpace(record.MultiplayerReview.Reason) == "" {
		return errors.New("multiplayer approval requires a review reason")
	}
	clues, err := engine.ParseClues(record.Clues)
	if err != nil {
		return err
	}
	solution, err := engine.ParseSolution(record.Solution)
	if err != nil {
		return err
	}
	if err := engine.NewPuzzle(clues, solution); err != nil {
		return err
	}
	count, unique := engine.CountSolutions(clues)
	if count != 1 || unique.String() != solution.String() {
		return errors.New("puzzle does not have the recorded unique solution")
	}
	logical, err := engine.SolveLogical(clues)
	if err != nil || logical.Solution.String() != solution.String() || logical.Grade != record.Difficulty ||
		logical.HardestTechnique.String() != record.HardestTechnique {
		return errors.New("logical grade or solver path is invalid")
	}
	if engine.CanonicalFingerprint(clues) != record.CanonicalFingerprint {
		return errors.New("canonical fingerprint is invalid")
	}
	if expected := engine.EvaluateQuality(clues, logical.Steps); expected != record.Quality {
		return errors.New("quality metrics are invalid")
	}
	expectedPath := SolverPath(logical.Steps)
	if len(expectedPath) != len(record.SolverPath) {
		return errors.New("solver path is invalid")
	}
	for index := range expectedPath {
		if expectedPath[index] != record.SolverPath[index] {
			return errors.New("solver path is invalid")
		}
	}
	if record.CanonicalVersion != engine.CanonicalVersion ||
		record.TransformationVersion != engine.TransformationVersion ||
		record.GeneratorVersion != engine.GeneratorVersion ||
		record.SolverVersion != engine.SolverVersion ||
		record.GraderVersion != engine.GraderVersion {
		return errors.New("puzzle tooling version is unsupported")
	}
	return nil
}

func SolverPath(steps []engine.Step) []PathSegment {
	var path []PathSegment
	for _, step := range steps {
		name := step.Technique.String()
		if len(path) == 0 || path[len(path)-1].Technique != name {
			path = append(path, PathSegment{Technique: name, Count: 1})
		} else {
			path[len(path)-1].Count++
		}
	}
	return path
}

type Selector struct{ records []Record }

var _ puzzledomain.CatalogSelector = Selector{}

func NewSelector(records []Record) Selector {
	return Selector{records: append([]Record(nil), records...)}
}

func (selector Selector) Select(difficulty shared.Difficulty, excluded map[string]struct{}) (puzzledomain.Puzzle, error) {
	for _, record := range selector.records {
		if record.Lifecycle != "Active" || string(record.Difficulty) != string(difficulty) {
			continue
		}
		if _, skip := excluded[record.CanonicalFingerprint]; skip {
			continue
		}
		id, err := shared.ParsePuzzleID(record.ID)
		if err != nil {
			return puzzledomain.Puzzle{}, err
		}
		return puzzledomain.Puzzle{
			ID: id, Revision: record.Revision, State: shared.PuzzleActive, Difficulty: difficulty,
			HardestTechnique: record.HardestTechnique, QualityScore: uint16(record.Quality.ClueCount),
			MultiplayerApproved: record.MultiplayerReview.Approved, GeneratorVersion: record.GeneratorVersion,
			SolverVersion: record.SolverVersion, CanonicalFingerprint: record.CanonicalFingerprint,
		}, nil
	}
	return puzzledomain.Puzzle{}, errors.New("no eligible puzzle is available")
}
