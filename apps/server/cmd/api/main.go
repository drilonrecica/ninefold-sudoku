package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/migrate"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/repository"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/sqlite"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/platform/config"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/platform/server"
)

var buildVersion = "unknown"

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	cfg, err := config.Load()
	if err != nil {
		logger.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

	db, err := sqlite.New(cfg.DatabasePath)
	if err != nil {
		logger.Error("database open failed", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("database close failed", "error", err)
		}
	}()

	if err := migrate.Up(db.Writer()); err != nil {
		logger.Error("database migration failed", "error", err)
		os.Exit(1)
	}
	logger.Info("database migrated", "version", db.Version())

	listener, err := net.Listen("tcp", cfg.HTTPAddress)
	if err != nil {
		logger.Error("listen failed", "error", err)
		os.Exit(1)
	}

	repo := repository.New(db)

	httpServer := server.New(cfg.HTTPAddress, buildVersion, cfg, db, repo, logger)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errs := make(chan error, 1)
	go func() {
		logger.Info("server started", "address", listener.Addr().String(), "version", buildVersion)
		errs <- httpServer.Serve(listener)
	}()

	select {
	case err = <-errs:
		if err != nil {
			logger.Error("server stopped unexpectedly", "error", err)
			os.Exit(1)
		}
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()
		if err = httpServer.Shutdown(shutdownCtx); err != nil && !errors.Is(err, context.Canceled) {
			logger.Error("graceful shutdown failed", "error", err)
			os.Exit(1)
		}
		if err = <-errs; err != nil {
			logger.Error("server shutdown failed", "error", err)
			os.Exit(1)
		}
	}
	logger.Info("server stopped")
}
