package maintenance

import (
	"context"
	"log/slog"
	"time"

	shared "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/domain"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/repository"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/sqlite"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/platform/config"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/platform/idgen"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/room/actor"
	roomdomain "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/room/domain"
)

const batchSize = 100

type Scheduler struct {
	repo     *repository.Repository
	db       *sqlite.DB
	registry *actor.Registry
	cfg      config.Config
	logger   *slog.Logger
}

func New(repo *repository.Repository, db *sqlite.DB, registry *actor.Registry, cfg config.Config, logger *slog.Logger) *Scheduler {
	return &Scheduler{repo: repo, db: db, registry: registry, cfg: cfg, logger: logger}
}

func (s *Scheduler) Run(ctx context.Context) {
	s.runObserved(ctx)
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.runObserved(ctx)
		}
	}
}

func (s *Scheduler) runObserved(parent context.Context) {
	ctx, cancel := context.WithTimeout(parent, 30*time.Second)
	defer cancel()
	started := time.Now()
	if err := s.RunOnce(ctx, time.Now()); err != nil {
		s.logger.Error("maintenance failed", "error", err, "latencyMs", time.Since(started).Milliseconds())
		return
	}
	s.logger.Info("maintenance completed", "latencyMs", time.Since(started).Milliseconds())
}

func (s *Scheduler) RunOnce(ctx context.Context, now time.Time) error {
	expired, err := s.repo.ListExpiredRooms(ctx, now.UnixMilli())
	if err != nil {
		return err
	}
	if len(expired) > batchSize {
		expired = expired[:batchSize]
	}
	for _, room := range expired {
		if err := s.expireRoom(ctx, room.ID, room.HostParticipantID, uint64(room.Version)); err != nil {
			return err
		}
	}
	if err := s.repo.DeleteExpiredSessions(ctx, now.UnixMilli()); err != nil {
		return err
	}
	if err := s.repo.DeleteExpiredCommandReceipts(ctx, now.UnixMilli()); err != nil {
		return err
	}
	if err := s.repo.DeleteExpiredReplayCapabilities(ctx, now.UnixMilli()); err != nil {
		return err
	}
	if _, err := s.repo.ScrubTerminalMatches(ctx, now.Add(-7*24*time.Hour).UnixMilli(), now.UnixMilli(), batchSize); err != nil {
		return err
	}
	if err := s.repo.DeleteExpiredMatchTombstones(ctx, now.Add(-s.cfg.MatchTombstoneRetention).UnixMilli()); err != nil {
		return err
	}
	if err := s.repo.DeleteExpiredAdminAuditLog(ctx, now.Add(-365*24*time.Hour).UnixMilli()); err != nil {
		return err
	}
	if err := s.db.Optimize(ctx); err != nil {
		return err
	}
	return s.db.Checkpoint(ctx)
}

func (s *Scheduler) expireRoom(ctx context.Context, roomIDValue, hostIDValue string, version uint64) error {
	roomID, err := shared.ParseRoomID(roomIDValue)
	if err != nil {
		return err
	}
	hostID, err := shared.ParseParticipantID(hostIDValue)
	if err != nil {
		return err
	}
	requestID, err := idgen.Generator{}.RequestID()
	if err != nil {
		return err
	}
	meta, err := shared.NewCommandMetadata(requestID, hostID, 1, shared.NewRoomTarget(roomID), version)
	if err != nil {
		return err
	}
	roomActor, err := s.registry.Acquire(ctx, roomID)
	if err != nil {
		return err
	}
	defer s.registry.Release(roomID)
	_, err = roomActor.Submit(ctx, actor.Envelope{
		RequestID: requestID, CommandType: "room.expire", Fingerprint: "room.expire:" + roomID.String(),
		Command: roomdomain.ExpireRoomCommand{Meta: meta},
	})
	return err
}
