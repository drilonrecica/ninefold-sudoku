package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/migrate"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/sqlite"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: maintenance <migrate|status|down>\n")
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
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}
