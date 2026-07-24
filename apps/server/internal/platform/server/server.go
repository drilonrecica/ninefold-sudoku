package server

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/migrate"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/repository"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/sqlite"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/platform/config"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/room/actor"
	roomhttp "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/room/transport/http"
)

type Server struct {
	httpServer *http.Server
	logger     *slog.Logger
	db         *sqlite.DB
	registry   *actor.Registry
}

func New(address, buildVersion string, cfg config.Config, db *sqlite.DB, repo *repository.Repository, logger *slog.Logger) *Server {
	router := chi.NewRouter()
	router.Get("/health/live", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status":  "live",
			"version": buildVersion,
		})
	})
	router.Get("/health/ready", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		if err := db.Health(ctx); err != nil {
			logger.Warn("readiness check failed", "error", err)
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "not_ready", "reason": "database_unhealthy"})
			return
		}
		current, err := migrate.Version(db.Writer())
		if err != nil {
			logger.Warn("readiness check failed", "error", err)
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "not_ready", "reason": "migration_version_unavailable"})
			return
		}
		expected, err := migrate.CurrentVersion()
		if err != nil {
			logger.Warn("readiness check failed", "error", err)
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "not_ready", "reason": "migration_version_unknown"})
			return
		}
		if current != expected {
			logger.Warn("readiness check failed", "current", current, "expected", expected)
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "not_ready", "reason": "migration_version_mismatch"})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ready", "version": buildVersion})
	})

	registry := actor.NewRegistry(repo, logger)
	roomHandler := roomhttp.NewHandler(repo, registry, cfg, logger)
	roomHandler.RegisterRoutes(router)

	return &Server{
		httpServer: &http.Server{
			Addr:              address,
			Handler:           router,
			ReadHeaderTimeout: 5 * time.Second,
			ReadTimeout:       10 * time.Second,
			WriteTimeout:      10 * time.Second,
			IdleTimeout:       60 * time.Second,
			MaxHeaderBytes:    1 << 20,
		},
		logger:   logger,
		db:       db,
		registry: registry,
	}
}

func (s *Server) Serve(listener net.Listener) error {
	err := s.httpServer.Serve(listener)
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.registry.ShutdownAll()
	return s.httpServer.Shutdown(ctx)
}
