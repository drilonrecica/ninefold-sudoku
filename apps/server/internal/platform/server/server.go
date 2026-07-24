package server

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/migrate"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/repository"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/sqlite"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/platform/config"
	replayproof "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/replay/proof"
	replayhttp "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/replay/transport/http"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/room/actor"
	roomhttp "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/room/transport/http"
	roomws "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/room/transport/websocket"
	solohttp "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/solo/transport/http"
)

type Server struct {
	httpServer *http.Server
	logger     *slog.Logger
	db         *sqlite.DB
	registry   *actor.Registry
	ready      atomic.Bool
}

func New(address, buildVersion string, cfg config.Config, db *sqlite.DB, repo *repository.Repository, logger *slog.Logger) *Server {
	registry := actor.NewRegistry(repo, logger, replayproof.Signer{
		KeyID: cfg.ReplaySigningKeyID, PrivateKey: cfg.ReplaySigningKey,
	})
	server := &Server{
		logger:   logger,
		db:       db,
		registry: registry,
	}
	if err := registry.RecoverNonTerminal(context.Background(), time.Now()); err != nil {
		logger.Error("startup recovery failed", "error", err)
	} else {
		server.ready.Store(true)
	}

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

		if !server.ready.Load() {
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "not_ready", "reason": "recovery_failed"})
			return
		}

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

	roomHandler := roomhttp.NewHandler(repo, registry, cfg, logger)
	roomHandler.RegisterRoutes(router)

	wsHandler := roomws.NewHandler(repo, registry, cfg, logger)
	wsHandler.RegisterRoutes(router)

	replayHandler := replayhttp.NewHandler(repo, logger)
	replayHandler.RegisterRoutes(router)

	soloHandler := solohttp.NewHandler(repo, cfg.CookieSecret)
	soloHandler.RegisterRoutes(router)

	server.httpServer = &http.Server{
		Addr:              address,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}
	return server
}

func (s *Server) Serve(listener net.Listener) error {
	err := s.httpServer.Serve(listener)
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.ready.Store(false)
	s.registry.NotifyMaintenance()
	s.registry.ShutdownAll()
	return s.httpServer.Shutdown(ctx)
}
