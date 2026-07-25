package websocket

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"nhooyr.io/websocket"

	shared "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/domain"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/repository"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/platform/config"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/platform/idgen"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/room/actor"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/room/session"
)

const maxMessageBytes = 64 << 10

// Handler upgrades HTTP connections to WebSocket and owns the connection lifecycle.
type Handler struct {
	repo     *repository.Repository
	registry *actor.Registry
	cfg      config.Config
	logger   *slog.Logger
	observer Observer
}

type Observer interface {
	ObserveReconnect()
}

// NewHandler creates a WebSocket handler.
func NewHandler(repo *repository.Repository, registry *actor.Registry, cfg config.Config, logger *slog.Logger, observers ...Observer) *Handler {
	handler := &Handler{
		repo:     repo,
		registry: registry,
		cfg:      cfg,
		logger:   logger,
	}
	if len(observers) > 0 {
		handler.observer = observers[0]
	}
	return handler
}

// RegisterRoutes registers the WebSocket endpoint.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/ws", h.ServeWS)
}

// ServeWS authenticates the session, validates origin, upgrades the connection,
// and attaches it to the Room actor.
func (h *Handler) ServeWS(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if !originAllowed(r.Header.Get("Origin"), h.cfg.AllowedOrigins) {
		http.Error(w, "origin not allowed", http.StatusForbidden)
		return
	}

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true, // Full origins are validated above.
		CompressionMode:    websocket.CompressionDisabled,
	})
	if err != nil {
		h.logger.Warn("websocket accept rejected", "error", err)
		return
	}
	conn.SetReadLimit(maxMessageBytes)

	token := session.Read(r)
	if token == "" {
		h.reject(conn, "SESSION_INVALID")
		return
	}
	hash := session.Hash(token)
	roomSession, err := h.repo.GetRoomSessionByHash(ctx, hash)
	if err != nil {
		if repository.IsNoRows(err) {
			h.reject(conn, "SESSION_INVALID")
			return
		}
		h.reject(conn, "SERVER_BUSY")
		return
	}
	if roomSession.ExpiresAtMs < time.Now().UnixMilli() {
		h.reject(conn, "SESSION_EXPIRED")
		return
	}

	roomID, err := shared.ParseRoomID(roomSession.RoomID)
	if err != nil {
		h.reject(conn, "SESSION_INVALID")
		return
	}
	participantID, err := shared.ParseParticipantID(roomSession.ParticipantID)
	if err != nil {
		h.reject(conn, "SESSION_INVALID")
		return
	}

	roomActor, err := h.registry.Acquire(ctx, roomID)
	if err != nil {
		h.reject(conn, "ROOM_NOT_FOUND")
		return
	}

	connID, err := idgen.Generator{}.ConnectionID()
	if err != nil {
		h.reject(conn, "SERVER_BUSY")
		h.registry.Release(roomID)
		return
	}

	c := newConnection(conn, roomActor, h.registry, roomID, participantID, connID, hash, h.logger, h.observer)
	// Use a detached lifetime context; the HTTP request context is canceled
	// when the upgrade handler returns.
	go c.run(context.Background())
}

func originAllowed(raw string, allowed []string) bool {
	origin, err := url.Parse(raw)
	if err != nil || origin.Scheme == "" || origin.Host == "" || origin.User != nil ||
		origin.Path != "" || origin.RawQuery != "" || origin.Fragment != "" {
		return false
	}
	normalized := strings.TrimSuffix(origin.String(), "/")
	for _, candidate := range allowed {
		if normalized == strings.TrimSuffix(candidate, "/") {
			return true
		}
	}
	return false
}

func (h *Handler) reject(conn *websocket.Conn, code string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	msg := serverMessage("connection.rejected", 0, 0, map[string]any{"code": code})
	_ = conn.Write(ctx, websocket.MessageText, msg)
	_ = conn.Close(websocket.StatusPolicyViolation, code)
}

func serverMessage(msgType string, eventNumber uint64, aggregateVersion uint64, payload map[string]any) []byte {
	view := map[string]any{
		"schemaVersion":    1,
		"eventNumber":      eventNumber,
		"aggregateVersion": aggregateVersion,
		"serverTimestamp":  time.Now().UnixMilli(),
		"type":             msgType,
		"payload":          payload,
	}
	b, err := json.Marshal(view)
	if err != nil {
		return []byte("{}")
	}
	return b
}
