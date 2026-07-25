package migrate

import (
	"context"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/pressly/goose/v3"

	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/sqlite"
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
	if v != 16 {
		t.Fatalf("expected version 16, got %d", v)
	}
}

func TestUpgradeRepresentativePriorSchemas(t *testing.T) {
	for _, priorVersion := range []int64{1, 5, 8, 12, 15} {
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
			if err != nil || version != 16 {
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
