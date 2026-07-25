package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

const minimumSQLiteVersion = "3.40.0"

// DB wraps the SQLite driver with one writer and one reader pool.
type DB struct {
	writer  *sql.DB
	readers *sql.DB
	version string
	path    string

	closeOnce sync.Once
	closeErr  error
}

// New opens the SQLite database at path, verifies the driver version, and applies the
// required PRAGMAs on both the writer and the reader pool.
func New(path string) (*DB, error) {
	if path == "" {
		return nil, errors.New("database path is required")
	}
	if path == ":memory:" {
		return nil, errors.New("in-memory SQLite is not allowed")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil && filepath.Dir(path) != "." {
		return nil, fmt.Errorf("create database directory: %w", err)
	}

	writer, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open writer: %w", err)
	}
	writer.SetMaxOpenConns(1)
	writer.SetMaxIdleConns(1)
	writer.SetConnMaxIdleTime(0)
	writer.SetConnMaxLifetime(0)

	readers, err := sql.Open("sqlite", path)
	if err != nil {
		_ = writer.Close()
		return nil, fmt.Errorf("open reader pool: %w", err)
	}
	readers.SetMaxOpenConns(4)
	readers.SetMaxIdleConns(4)
	readers.SetConnMaxIdleTime(10 * time.Minute)
	readers.SetConnMaxLifetime(0)

	db := &DB{writer: writer, readers: readers, path: path}
	if err := db.applyPragmas(writer); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := db.applyPragmas(readers); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := db.verifyVersion(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

func (d *DB) applyPragmas(pool *sql.DB) error {
	pragmas := []string{
		"PRAGMA journal_mode = WAL;",
		"PRAGMA foreign_keys = ON;",
		"PRAGMA busy_timeout = 5000;",
		"PRAGMA synchronous = FULL;",
		"PRAGMA temp_store = MEMORY;",
	}
	for _, p := range pragmas {
		if _, err := pool.Exec(p); err != nil {
			return fmt.Errorf("apply %s: %w", p, err)
		}
	}
	return nil
}

func (d *DB) verifyVersion() error {
	if err := d.writer.QueryRow("SELECT sqlite_version()").Scan(&d.version); err != nil {
		return fmt.Errorf("read sqlite_version: %w", err)
	}
	cmp, err := compareSQLiteVersions(d.version, minimumSQLiteVersion)
	if err != nil {
		return fmt.Errorf("parse sqlite_version %q: %w", d.version, err)
	}
	if cmp < 0 {
		return fmt.Errorf("sqlite version %s is below required %s", d.version, minimumSQLiteVersion)
	}
	return nil
}

// compareSQLiteVersions parses a.b.c strings and returns -1/0/1.
func compareSQLiteVersions(a, b string) (int, error) {
	pa, err := parseVersion(a)
	if err != nil {
		return 0, err
	}
	pb, err := parseVersion(b)
	if err != nil {
		return 0, err
	}
	for i := 0; i < 3; i++ {
		if pa[i] < pb[i] {
			return -1, nil
		}
		if pa[i] > pb[i] {
			return 1, nil
		}
	}
	return 0, nil
}

func parseVersion(v string) ([3]int, error) {
	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		return [3]int{}, fmt.Errorf("expected major.minor.patch, got %q", v)
	}
	var out [3]int
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil {
			return [3]int{}, err
		}
		out[i] = n
	}
	return out, nil
}

// Writer returns the single-authoritative-writer connection pool.
func (d *DB) Writer() *sql.DB { return d.writer }

// Readers returns the read-only query pool.
func (d *DB) Readers() *sql.DB { return d.readers }

// Version returns the reported SQLite runtime version.
func (d *DB) Version() string { return d.version }

// Health checks that the writer responds and required PRAGMAs are in effect.
func (d *DB) Health(ctx context.Context) error {
	if err := d.writer.PingContext(ctx); err != nil {
		return err
	}
	var journal string
	if err := d.writer.QueryRowContext(ctx, "PRAGMA journal_mode;").Scan(&journal); err != nil {
		return err
	}
	if !strings.EqualFold(journal, "wal") {
		return fmt.Errorf("journal_mode=%s, want wal", journal)
	}
	var fk int
	if err := d.writer.QueryRowContext(ctx, "PRAGMA foreign_keys;").Scan(&fk); err != nil {
		return err
	}
	if fk != 1 {
		return fmt.Errorf("foreign_keys=%d, want 1", fk)
	}
	return nil
}

func (d *DB) Checkpoint(ctx context.Context) error {
	_, err := d.writer.ExecContext(ctx, "PRAGMA wal_checkpoint(TRUNCATE);")
	return err
}

func (d *DB) Optimize(ctx context.Context) error {
	_, err := d.writer.ExecContext(ctx, "PRAGMA optimize;")
	return err
}

// IntegrityCheck is intentionally not part of routine hourly maintenance. It
// is reserved for recovery failures and explicit operational diagnosis.
func (d *DB) IntegrityCheck(ctx context.Context) error {
	var result string
	if err := d.writer.QueryRowContext(ctx, "PRAGMA integrity_check;").Scan(&result); err != nil {
		return err
	}
	if result != "ok" {
		return fmt.Errorf("sqlite integrity_check: %s", result)
	}
	return nil
}

// Close closes both pools.
func (d *DB) Close() error {
	d.closeOnce.Do(func() {
		if err := d.writer.Close(); err != nil {
			d.closeErr = err
		}
		if err := d.readers.Close(); err != nil && d.closeErr == nil {
			d.closeErr = err
		}
	})
	return d.closeErr
}
