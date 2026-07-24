package http

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/gen"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/repository"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/platform/idgen"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/replay/capability"
	roomsession "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/room/session"
)

const (
	capabilityLifetime = 7 * 24 * time.Hour
	maxReplayEvents    = 10_000
)

type Handler struct {
	repo   *repository.Repository
	logger *slog.Logger
}

func NewHandler(repo *repository.Repository, logger *slog.Logger) *Handler {
	return &Handler{repo: repo, logger: logger}
}

func (h *Handler) RegisterRoutes(router chi.Router) {
	router.Post("/api/v1/replays/{matchID}/capabilities", h.CreateCapability)
	router.Get("/api/v1/replays/{replayID}", h.GetReplay)
	router.Delete("/api/v1/replays/{replayID}", h.DeleteReplay)
}

type capabilityResponse struct {
	ReplayID   string `json:"replayId"`
	Capability string `json:"capability"`
	ShareURL   string `json:"shareUrl"`
	ExpiresAt  int64  `json:"expiresAt"`
}

type replayEvent struct {
	EventNumber      int64          `json:"eventNumber"`
	AggregateVersion int64          `json:"aggregateVersion"`
	ServerTimestamp  int64          `json:"serverTimestamp"`
	Type             string         `json:"type"`
	Payload          map[string]any `json:"payload"`
}

type replayParticipant struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type replayRules struct {
	Mode            string `json:"mode"`
	Difficulty      string `json:"difficulty"`
	ErrorPreset     string `json:"errorPreset"`
	HintsEnabled    bool   `json:"hintsEnabled"`
	AutoRemoveNotes bool   `json:"autoRemoveNotes"`
	RuleVersion     int64  `json:"ruleVersion"`
}

type replayDocument struct {
	SchemaVersion int                 `json:"schemaVersion"`
	ReplayID      string              `json:"replayId"`
	MatchID       string              `json:"matchId"`
	ExpiresAt     int64               `json:"expiresAt"`
	Clues         string              `json:"clues"`
	Rules         replayRules         `json:"rules"`
	Participants  []replayParticipant `json:"participants"`
	Events        []replayEvent       `json:"events"`
}

func (h *Handler) CreateCapability(w http.ResponseWriter, r *http.Request) {
	setPrivateHeaders(w)
	ctx := r.Context()
	matchID := chi.URLParam(r, "matchID")
	match, err := h.repo.GetMatchByID(ctx, matchID)
	if err != nil || match.State != "Completed" {
		writeUnavailable(w)
		return
	}
	rawSession := roomsession.Read(r)
	if rawSession == "" {
		writeUnavailable(w)
		return
	}
	session, err := h.repo.GetRoomSessionByHash(ctx, roomsession.Hash(rawSession))
	if err != nil || session.RoomID != match.RoomID || session.ExpiresAtMs <= time.Now().UnixMilli() {
		writeUnavailable(w)
		return
	}
	participants, err := h.repo.ListMatchParticipants(ctx, matchID)
	if err != nil || !containsParticipant(participants, session.ParticipantID) {
		writeUnavailable(w)
		return
	}

	replayID, err := idgen.Generator{}.ReplayID()
	if err != nil {
		writeInternal(w)
		return
	}
	token, err := capability.Generate()
	if err != nil {
		writeInternal(w)
		return
	}
	now := time.Now()
	expiresAt := now.Add(capabilityLifetime).UnixMilli()
	tx, _, err := h.repo.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		writeInternal(w)
		return
	}
	defer repository.TxRollback(tx)
	if err := h.repo.CreateReplayCapability(ctx, tx, gen.ReplayCapability{
		TokenHash: token.Hash, ReplayID: replayID.String(), MatchID: matchID,
		CreatedAtMs: now.UnixMilli(), ExpiresAtMs: expiresAt,
	}); err != nil || repository.TxCommit(tx) != nil {
		writeInternal(w)
		return
	}
	writeJSON(w, http.StatusCreated, capabilityResponse{
		ReplayID: replayID.String(), Capability: token.Value,
		ShareURL:  "/replay/" + replayID.String() + "#cap=" + token.Value,
		ExpiresAt: expiresAt,
	})
}

func (h *Handler) GetReplay(w http.ResponseWriter, r *http.Request) {
	setPrivateHeaders(w)
	token := bearerToken(r.Header.Get("Authorization"))
	if token == "" {
		writeUnavailable(w)
		return
	}
	stored, err := h.repo.GetReplayCapabilityByHash(r.Context(), capability.Hash(token))
	now := time.Now().UnixMilli()
	if err != nil || stored.ReplayID != chi.URLParam(r, "replayID") ||
		stored.ExpiresAtMs <= now || stored.RevokedAtMs.Valid {
		writeUnavailable(w)
		return
	}
	document, err := h.projectReplay(r, stored)
	if err != nil {
		h.logger.Warn("replay projection failed", "replayID", stored.ReplayID, "error", err)
		writeUnavailable(w)
		return
	}
	writeJSON(w, http.StatusOK, document)
}

func (h *Handler) DeleteReplay(w http.ResponseWriter, r *http.Request) {
	setPrivateHeaders(w)
	ctx := r.Context()
	stored, err := h.repo.GetReplayCapabilityByReplayID(ctx, chi.URLParam(r, "replayID"))
	if err != nil {
		writeUnavailable(w)
		return
	}
	match, err := h.repo.GetMatchByID(ctx, stored.MatchID)
	if err != nil {
		writeUnavailable(w)
		return
	}
	rawSession := roomsession.Read(r)
	if rawSession == "" {
		writeUnavailable(w)
		return
	}
	session, err := h.repo.GetRoomSessionByHash(ctx, roomsession.Hash(rawSession))
	if err != nil || session.RoomID != match.RoomID || session.ExpiresAtMs <= time.Now().UnixMilli() {
		writeUnavailable(w)
		return
	}
	participants, err := h.repo.ListMatchParticipants(ctx, stored.MatchID)
	if err != nil || !containsParticipant(participants, session.ParticipantID) {
		writeUnavailable(w)
		return
	}
	tx, _, err := h.repo.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		writeInternal(w)
		return
	}
	defer repository.TxRollback(tx)
	if err := h.repo.RevokeReplayCapabilitiesByMatch(ctx, tx, stored.MatchID, time.Now().UnixMilli()); err != nil ||
		repository.TxCommit(tx) != nil {
		writeInternal(w)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) projectReplay(r *http.Request, stored gen.ReplayCapability) (replayDocument, error) {
	ctx := r.Context()
	match, err := h.repo.GetMatchByID(ctx, stored.MatchID)
	if err != nil || match.State != "Completed" {
		return replayDocument{}, errors.New("completed match unavailable")
	}
	puzzle, err := h.repo.GetPuzzle(ctx, match.PuzzleID, match.PuzzleRevision)
	if err != nil {
		return replayDocument{}, err
	}
	clues := decimalGridString(puzzle.Clues)
	if len(clues) != 81 {
		return replayDocument{}, errors.New("invalid puzzle assignment")
	}
	rows, err := h.repo.GetMatchEvents(ctx, stored.MatchID)
	if err != nil {
		return replayDocument{}, err
	}
	if len(rows) > maxReplayEvents {
		return replayDocument{}, errors.New("replay event limit exceeded")
	}
	events := make([]replayEvent, 0, len(rows))
	var expected int64 = 1
	for _, row := range rows {
		if row.EventNumber != expected {
			return replayDocument{}, errors.New("event sequence gap")
		}
		var payload map[string]any
		if err := json.Unmarshal([]byte(row.PublicPayloadJson), &payload); err != nil {
			return replayDocument{}, err
		}
		events = append(events, replayEvent{
			EventNumber: row.EventNumber, AggregateVersion: row.AggregateVersion,
			ServerTimestamp: row.OccurredAtMs, Type: row.PublicEventType, Payload: payload,
		})
		expected++
	}
	resultPlayers, err := h.repo.GetMatchResultPlayers(ctx, stored.MatchID)
	if err != nil {
		return replayDocument{}, err
	}
	participants := make([]replayParticipant, 0, len(resultPlayers))
	for _, participant := range resultPlayers {
		participants = append(participants, replayParticipant{ID: participant.ParticipantID, Name: participant.DisplayName})
	}
	return replayDocument{
		SchemaVersion: 1, ReplayID: stored.ReplayID, MatchID: stored.MatchID,
		ExpiresAt: stored.ExpiresAtMs, Clues: clues,
		Rules: replayRules{
			Mode: match.Mode, Difficulty: match.Difficulty, ErrorPreset: match.ErrorPreset,
			HintsEnabled: match.HintsEnabled == 1, AutoRemoveNotes: match.AutoRemoveNotes == 1,
			RuleVersion: match.RuleVersion,
		},
		Participants: participants, Events: events,
	}, nil
}

func containsParticipant(participants []gen.MatchParticipant, participantID string) bool {
	for _, participant := range participants {
		if participant.ParticipantID == participantID {
			return true
		}
	}
	return false
}

func decimalGridString(values []byte) string {
	encoded := make([]byte, len(values))
	for index, value := range values {
		if value <= 9 {
			encoded[index] = '0' + value
		} else {
			encoded[index] = value
		}
	}
	return string(encoded)
}

func bearerToken(header string) string {
	scheme, token, ok := strings.Cut(header, " ")
	if !ok || !strings.EqualFold(scheme, "Bearer") || token == "" || strings.ContainsAny(token, " \t\r\n") {
		return ""
	}
	return token
}

func setPrivateHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("X-Robots-Tag", "noindex, nofollow")
}

func writeUnavailable(w http.ResponseWriter) {
	writeJSON(w, http.StatusNotFound, map[string]any{"error": map[string]any{
		"code": "REPLAY_UNAVAILABLE", "messageKey": "error.replay_unavailable",
		"requestId": "", "details": map[string]any{},
	}})
}

func writeInternal(w http.ResponseWriter) {
	writeJSON(w, http.StatusInternalServerError, map[string]any{"error": map[string]any{
		"code": "PERSISTENCE_FAILED", "messageKey": "error.persistence_failed",
		"requestId": "", "details": map[string]any{},
	}})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
