package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/migrate"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/repository"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/sqlite"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/puzzle/catalog"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: maintenance <migrate|status|down|seed-e2e>\n")
		os.Exit(1)
	}

	path := os.Getenv("NINEFOLD_DATABASE_PATH")
	if path == "" {
		path = "data/ninefold.db"
	}
	if path == ":memory:" || filepath.Clean(path) == "." {
		fmt.Fprintf(os.Stderr, "invalid NINEFOLD_DATABASE_PATH\n")
		os.Exit(1)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		fmt.Fprintf(os.Stderr, "create database directory: %v\n", err)
		os.Exit(1)
	}

	db, err := sqlite.New(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "database open failed: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	switch os.Args[1] {
	case "migrate":
		if err := migrate.Up(db.Writer()); err != nil {
			fmt.Fprintf(os.Stderr, "migrate failed: %v\n", err)
			os.Exit(1)
		}
		v, err := migrate.Version(db.Writer())
		if err != nil {
			fmt.Fprintf(os.Stderr, "version failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("migrated to version %d\n", v)
	case "status":
		if err := migrate.Status(db.Writer()); err != nil {
			fmt.Fprintf(os.Stderr, "status failed: %v\n", err)
			os.Exit(1)
		}
	case "down":
		if err := migrate.Down(db.Writer()); err != nil {
			fmt.Fprintf(os.Stderr, "down failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("rolled back all migrations")
	case "seed-e2e":
		if os.Getenv("NINEFOLD_ENVIRONMENT") != "test" {
			fmt.Fprintln(os.Stderr, "seed-e2e is restricted to NINEFOLD_ENVIRONMENT=test")
			os.Exit(1)
		}
		if len(os.Args) != 3 {
			fmt.Fprintln(os.Stderr, "usage: maintenance seed-e2e <catalog.jsonl>")
			os.Exit(1)
		}
		if err := migrate.Up(db.Writer()); err != nil {
			fmt.Fprintf(os.Stderr, "migrate failed: %v\n", err)
			os.Exit(1)
		}
		if err := verifyE2ECatalog(context.Background(), repository.New(db), os.Args[2]); err != nil {
			fmt.Fprintf(os.Stderr, "verify e2e catalog failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("verified e2e puzzle catalog")
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}

func verifyE2ECatalog(ctx context.Context, repo *repository.Repository, path string) error {
	records, err := catalog.ReadFile(path)
	if err != nil {
		return err
	}
	for _, record := range records {
		clues, err := decimalGrid(record.Clues)
		if err != nil {
			return fmt.Errorf("puzzle %s clues: %w", record.ID, err)
		}
		solution, err := decimalGrid(record.Solution)
		if err != nil {
			return fmt.Errorf("puzzle %s solution: %w", record.ID, err)
		}
		puzzle, err := repo.GetPuzzle(ctx, record.ID, int64(record.Revision))
		if err != nil {
			return fmt.Errorf("get migrated puzzle %s: %w", record.ID, err)
		}
		if puzzle.State != record.Lifecycle ||
			puzzle.Difficulty != string(record.Difficulty) ||
			puzzle.HardestTechnique != record.HardestTechnique ||
			puzzle.QualityScore != float64(record.Quality.LogicalStepCount) ||
			puzzle.MultiplayerApproved != 1 ||
			puzzle.GeneratorVersion != record.GeneratorVersion ||
			puzzle.SolverVersion != record.SolverVersion ||
			puzzle.CanonicalFingerprint != record.CanonicalFingerprint ||
			!bytes.Equal(puzzle.Clues, clues) ||
			!bytes.Equal(puzzle.Solution, solution) {
			return fmt.Errorf("migrated puzzle %s does not match catalog", record.ID)
		}
	}
	return nil
}

func decimalGrid(value string) ([]byte, error) {
	if len(value) != 81 {
		return nil, fmt.Errorf("grid must contain 81 digits")
	}
	grid := make([]byte, 81)
	for index, digit := range []byte(value) {
		if digit < '0' || digit > '9' {
			return nil, fmt.Errorf("invalid digit at index %d", index)
		}
		grid[index] = digit - '0'
	}
	return grid, nil
}
