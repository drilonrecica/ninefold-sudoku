package migrate

import (
	"context"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/pressly/goose/v3"

	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/sqlite"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/puzzle/catalog"
)

func TestUpAndVersion(t *testing.T) {
	db, err := sqlite.New(filepath.Join(t.TempDir(), "migrate.db"))
	if err != nil {
		t.Fatalf("sqlite new: %v", err)
	}
	defer db.Close()
	if err := Up(db.Writer()); err != nil {
		t.Fatalf("migrate up: %v", err)
	}
	v, err := Version(db.Writer())
	if err != nil {
		t.Fatalf("version: %v", err)
	}
	if v != 17 {
		t.Fatalf("expected version 17, got %d", v)
	}
}

func TestUpgradeRepresentativePriorSchemas(t *testing.T) {
	for _, priorVersion := range []int64{1, 5, 8, 12, 15, 16} {
		t.Run("from_"+strconv.FormatInt(priorVersion, 10), func(t *testing.T) {
			db, err := sqlite.New(filepath.Join(t.TempDir(), "upgrade.db"))
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			if err := goose.UpTo(db.Writer(), "migrations", priorVersion); err != nil {
				t.Fatalf("migrate to %d: %v", priorVersion, err)
			}
			if err := Up(db.Writer()); err != nil {
				t.Fatalf("upgrade from %d: %v", priorVersion, err)
			}
			version, err := Version(db.Writer())
			if err != nil || version != 17 {
				t.Fatalf("version=%d err=%v", version, err)
			}
			if err := db.Health(context.Background()); err != nil {
				t.Fatalf("database health after upgrade: %v", err)
			}
			var foreignKeys int
			if err := db.Writer().QueryRow("PRAGMA foreign_keys").Scan(&foreignKeys); err != nil || foreignKeys != 1 {
				t.Fatalf("foreign_keys=%d err=%v", foreignKeys, err)
			}
		})
	}
}

func TestClosedBetaCatalogSeedMatchesVerifiedCatalog(t *testing.T) {
	db, err := sqlite.New(filepath.Join(t.TempDir(), "catalog.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := Up(db.Writer()); err != nil {
		t.Fatal(err)
	}
	if err := Up(db.Writer()); err != nil {
		t.Fatalf("repeated migration: %v", err)
	}

	records, err := catalog.ReadFile("../../puzzle/catalog/catalog.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 10 {
		t.Fatalf("catalog records=%d want=10", len(records))
	}
	expected := make(map[string]catalog.Record, len(records))
	for _, record := range records {
		expected[record.ID] = record
		if record.Lifecycle != "Active" || !record.MultiplayerReview.Approved ||
			record.MultiplayerReview.Reason != "Approved for closed beta deployment testing only" {
			t.Fatalf("puzzle %s lacks closed-beta approval", record.ID)
		}
	}

	rows, err := db.Writer().Query(`
		SELECT id, revision, state, difficulty, hardest_technique, quality_score,
		       multiplayer_approved, generator_version, solver_version,
		       canonical_fingerprint, clues, solution
		FROM puzzles
		ORDER BY id
	`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	counts := map[string]int{}
	seen := 0
	for rows.Next() {
		var (
			id, state, difficulty, technique, generator, solver, fingerprint string
			revision, approved                                               int64
			quality                                                          float64
			clues, solution                                                  []byte
		)
		if err := rows.Scan(&id, &revision, &state, &difficulty, &technique, &quality,
			&approved, &generator, &solver, &fingerprint, &clues, &solution); err != nil {
			t.Fatal(err)
		}
		record, ok := expected[id]
		if !ok {
			t.Fatalf("unexpected seeded puzzle %s", id)
		}
		if revision != int64(record.Revision) || state != record.Lifecycle ||
			difficulty != string(record.Difficulty) || technique != record.HardestTechnique ||
			quality != float64(record.Quality.LogicalStepCount) || approved != 1 ||
			generator != record.GeneratorVersion || solver != record.SolverVersion ||
			fingerprint != record.CanonicalFingerprint ||
			decimalBlob(clues) != record.Clues || decimalBlob(solution) != record.Solution {
			t.Fatalf("seeded puzzle %s does not match catalog", id)
		}
		counts[difficulty]++
		seen++
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if err := rows.Close(); err != nil {
		t.Fatal(err)
	}
	if seen != 10 {
		t.Fatalf("seeded puzzles=%d want=10", seen)
	}
	for difficulty, want := range map[string]int{"Easy": 3, "Medium": 3, "Hard": 2, "Expert": 2} {
		if counts[difficulty] != want {
			t.Fatalf("%s puzzles=%d want=%d", difficulty, counts[difficulty], want)
		}
	}

	if err := goose.Down(db.Writer(), "migrations"); err != nil {
		t.Fatalf("roll back seed migration: %v", err)
	}
	var retained int
	if err := db.Writer().QueryRow("SELECT COUNT(*) FROM puzzles").Scan(&retained); err != nil {
		t.Fatal(err)
	}
	if retained != 10 {
		t.Fatalf("seeded puzzles retained after rollback=%d want=10", retained)
	}
}

func decimalBlob(value []byte) string {
	encoded := make([]byte, len(value))
	for index, digit := range value {
		if digit > 9 {
			return ""
		}
		encoded[index] = '0' + digit
	}
	return string(encoded)
}
