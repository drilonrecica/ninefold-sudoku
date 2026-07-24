package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

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

	listener, err := net.Listen("tcp", cfg.HTTPAddress)
	if err != nil {
		logger.Error("listen failed", "error", err)
		os.Exit(1)
	}

	httpServer := server.New(cfg.HTTPAddress, buildVersion, logger)
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
