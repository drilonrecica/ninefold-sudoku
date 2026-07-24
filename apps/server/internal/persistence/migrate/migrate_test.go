package migrate

import (
	"os"
	"testing"

	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/sqlite"
)

func TestUpAndVersion(t *testing.T) {
	os.Remove("/tmp/migrate_test.db")
	os.Remove("/tmp/migrate_test.db-shm")
	os.Remove("/tmp/migrate_test.db-wal")
	db, err := sqlite.New("/tmp/migrate_test.db")
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
	if v != 13 {
		t.Fatalf("expected version 13, got %d", v)
	}
}
