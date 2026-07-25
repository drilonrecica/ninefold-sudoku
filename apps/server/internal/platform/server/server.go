package server

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/go-chi/chi/v5"

	adminhttp "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/admin/http"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/migrate"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/repository"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/sqlite"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/platform/config"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/platform/maintenance"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/platform/ops"
	replayproof "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/replay/proof"
	replayhttp "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/replay/transport/http"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/room/actor"
	roomhttp "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/room/transport/http"
	roomws "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/room/transport/websocket"
	solohttp "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/solo/transport/http"
)

type Server struct {
	httpServer      *http.Server
	logger          *slog.Logger
	db              *sqlite.DB
	registry        *actor.Registry
	ready           atomic.Bool
	stopMaintenance context.CancelFunc
	maintenanceDone chan struct{}
}

func New(address, buildVersion string, cfg config.Config, db *sqlite.DB, repo *repository.Repository, logger *slog.Logger) *Server {
	metrics := ops.NewMetrics()
	repo.SetObserver(metrics)
	registry := actor.NewRegistry(repo, logger, replayproof.Signer{
		KeyID: cfg.ReplaySigningKeyID, PrivateKey: cfg.ReplaySigningKey,
	})
	registry.SetObserver(metrics)
	server := &Server{
		logger:          logger,
		db:              db,
		registry:        registry,
		maintenanceDone: make(chan struct{}),
	}
	if err := registry.RecoverNonTerminal(context.Background(), time.Now()); err != nil {
		logger.Error("startup recovery failed", "error", err)
		checkCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		if checkErr := db.IntegrityCheck(checkCtx); checkErr != nil {
			logger.Error("conditional database integrity check failed", "error", checkErr)
		}
		cancel()
	} else {
		server.ready.Store(true)
	}

	router := chi.NewRouter()
	router.Use(metrics.Middleware)
	router.Use(func(next http.Handler) http.Handler { return ops.RequestLog(logger, next) })
	router.Use(func(next http.Handler) http.Handler { return ops.BodyLimit(1<<20, next) })
	router.Use(func(next http.Handler) http.Handler {
		return ops.SecurityHeaders(cfg.Environment == config.Production, next)
	})
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			isCreate := r.Method == http.MethodPost && r.URL.Path == "/api/v1/rooms"
			isJoin := r.Method == http.MethodPost &&
				strings.HasPrefix(r.URL.Path, "/api/v1/rooms/") &&
				strings.HasSuffix(r.URL.Path, "/join")
			if !server.ready.Load() && (isCreate || isJoin) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", "5")
				w.WriteHeader(http.StatusServiceUnavailable)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"code": "SERVICE_UNAVAILABLE", "message": "server maintenance in progress",
				})
				return
			}
			next.ServeHTTP(w, r)
		})
	})
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

	wsHandler := roomws.NewHandler(repo, registry, cfg, logger, metrics)
	wsHandler.RegisterRoutes(router)

	replayHandler := replayhttp.NewHandler(repo, logger)
	replayHandler.RegisterRoutes(router)

	soloHandler := solohttp.NewHandler(repo, cfg.CookieSecret)
	soloHandler.RegisterRoutes(router)

	metrics.SetRuntimeProvider(func(ctx context.Context) ops.RuntimeMetrics {
		actors, connections, actorQueue, outboundQueue := registry.OperationalStats()
		activeMatches, err := repo.CountActiveMatches(ctx)
		if err != nil {
			logger.Warn("active match metric unavailable", "error", err)
		}
		return ops.RuntimeMetrics{
			ActiveWebSockets: connections, ActiveRoomActors: actors, ActiveMatches: activeMatches,
			ActorQueueDepth: actorQueue, OutboundQueueDepth: outboundQueue,
		}
	})

	adminHandler := adminhttp.New(repo, registry, func() map[string]any {
		current, _ := migrate.Version(db.Writer())
		return map[string]any{
			"status": "ready", "version": buildVersion, "sqliteVersion": db.Version(),
			"migrationVersion": current, "actorCount": registry.Count(), "config": cfg.Sanitized(),
		}
	})
	router.Group(func(private chi.Router) {
		private.Use(func(next http.Handler) http.Handler {
			return ops.AdminOnly(cfg.AdminProxyHeader, cfg.AdminTrustedProxies, next)
		})
		private.Handle("/internal/metrics", metrics)
		adminHandler.RegisterRoutes(private)
	})

	server.httpServer = &http.Server{
		Addr:              address,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    32 << 10,
	}
	maintenanceCtx, stopMaintenance := context.WithCancel(context.Background())
	server.stopMaintenance = stopMaintenance
	scheduler := maintenance.New(repo, db, registry, cfg, logger)
	go func() {
		defer close(server.maintenanceDone)
		scheduler.Run(maintenanceCtx)
	}()
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
	s.stopMaintenance()
	select {
	case <-s.maintenanceDone:
	case <-ctx.Done():
		return ctx.Err()
	}
	s.registry.NotifyMaintenance()
	s.registry.ShutdownAll()
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return err
	}
	return s.db.Checkpoint(ctx)
}
