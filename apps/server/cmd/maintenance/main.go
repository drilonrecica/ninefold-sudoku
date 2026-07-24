package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/gen"
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
		if err := seedE2ECatalog(context.Background(), repository.New(db), os.Args[2]); err != nil {
			fmt.Fprintf(os.Stderr, "seed e2e catalog failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("seeded e2e puzzle catalog")
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}

func seedE2ECatalog(ctx context.Context, repo *repository.Repository, path string) error {
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
		if err := repo.CreatePuzzle(ctx, gen.Puzzle{
			ID:                   record.ID,
			Revision:             int64(record.Revision),
			State:                "Active",
			Difficulty:           string(record.Difficulty),
			HardestTechnique:     record.HardestTechnique,
			QualityScore:         float64(record.Quality.LogicalStepCount),
			MultiplayerApproved:  1,
			GeneratorVersion:     record.GeneratorVersion,
			SolverVersion:        record.SolverVersion,
			CanonicalFingerprint: record.CanonicalFingerprint,
			Clues:                clues,
			Solution:             solution,
			CreatedAtMs:          time.Now().UnixMilli(),
		}); err != nil {
			return err
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
