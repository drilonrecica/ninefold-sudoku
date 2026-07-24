// Package migrate embeds and applies Goose migrations.
package migrate

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func init() {
	goose.SetBaseFS(migrationsFS)
	goose.SetDialect("sqlite3")
	goose.SetLogger(goose.NopLogger())
}

// Up applies all pending migrations.
func Up(db *sql.DB) error {
	return goose.Up(db, "migrations")
}

// Version returns the current migration version.
func Version(db *sql.DB) (int64, error) {
	return goose.GetDBVersion(db)
}

// Down rolls back all migrations (intended for tests).
func Down(db *sql.DB) error {
	return goose.DownTo(db, "migrations", 0)
}

// Status returns verbose migration status.
func Status(db *sql.DB) error {
	return goose.Status(db, "migrations")
}

// CurrentVersion returns the highest numeric migration prefix present in the embedded FS.
func CurrentVersion() (int64, error) {
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return 0, err
	}
	var maxVersion int64
	for _, entry := range entries {
		name := entry.Name()
		var version int64
		if _, scanErr := fmt.Sscanf(name, "%05d", &version); scanErr != nil {
			continue
		}
		if version > maxVersion {
			maxVersion = version
		}
	}
	if maxVersion == 0 {
		return 0, fmt.Errorf("no migrations found in embedded FS")
	}
	return maxVersion, nil
}
