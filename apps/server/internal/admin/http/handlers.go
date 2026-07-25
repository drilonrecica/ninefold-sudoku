package adminhttp

import (
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	shared "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/domain"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/gen"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/repository"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/platform/ops"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/room/actor"
	roomdomain "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/room/domain"
)

type Handler struct {
	repo     *repository.Repository
	registry *actor.Registry
	status   func() map[string]any
}

func New(repo *repository.Repository, registry *actor.Registry, status func() map[string]any) *Handler {
	return &Handler{repo: repo, registry: registry, status: status}
}

func (h *Handler) RegisterRoutes(router chi.Router) {
	router.Get("/health/status", h.health)
	router.Get("/internal/admin/rooms/{code}", h.room)
	router.Post("/internal/admin/rooms/{code}/terminate", h.terminateRoom)
	router.Delete("/internal/admin/replays/{replayID}", h.deleteReplay)
	router.Post("/internal/admin/puzzles/{puzzleID}/{revision}/retire", h.retirePuzzle)
}

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, h.status())
}

func (h *Handler) room(w http.ResponseWriter, r *http.Request) {
	code, err := shared.ParseRoomCode(strings.ToUpper(chi.URLParam(r, "code")))
	if err != nil {
		writeError(w, http.StatusBadRequest, "ROOM_CODE_INVALID")
		return
	}
	room, err := h.repo.GetRoomByCode(r.Context(), code.String())
	if err != nil {
		writeError(w, http.StatusNotFound, "ROOM_NOT_FOUND")
		return
	}
	participants, err := h.repo.ListActiveRoomParticipants(r.Context(), room.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ADMIN_OPERATION_FAILED")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"code": code, "state": room.State, "mode": room.Mode, "difficulty": room.Difficulty,
		"participantCount": len(participants), "hasCurrentMatch": room.CurrentMatchID.Valid,
		"expiresAtMs": room.ExpiresAtMs,
	})
}

type confirmedAction struct {
	RequestID string `json:"requestId"`
	Confirm   bool   `json:"confirm"`
	Reason    string `json:"reason"`
}

func decodeAction(w http.ResponseWriter, r *http.Request) (confirmedAction, shared.RequestID, bool) {
	r.Body = http.MaxBytesReader(w, r.Body, 2048)
	var action confirmedAction
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") ||
		json.NewDecoder(r.Body).Decode(&action) != nil || !action.Confirm || len(action.Reason) > 200 {
		writeError(w, http.StatusBadRequest, "CONFIRMATION_REQUIRED")
		return action, "", false
	}
	requestID, err := shared.ParseRequestID(action.RequestID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "REQUEST_ID_INVALID")
		return action, "", false
	}
	return action, requestID, true
}

func (h *Handler) terminateRoom(w http.ResponseWriter, r *http.Request) {
	action, requestID, ok := decodeAction(w, r)
	if !ok {
		return
	}
	code, err := shared.ParseRoomCode(strings.ToUpper(chi.URLParam(r, "code")))
	if err != nil {
		writeError(w, http.StatusBadRequest, "ROOM_CODE_INVALID")
		return
	}
	stored, err := h.repo.GetRoomByCode(r.Context(), code.String())
	if err != nil || stored.HostParticipantID == "" {
		writeError(w, http.StatusNotFound, "ROOM_NOT_FOUND")
		return
	}
	roomID, roomErr := shared.ParseRoomID(stored.ID)
	hostID, hostErr := shared.ParseParticipantID(stored.HostParticipantID)
	meta, metaErr := shared.NewCommandMetadata(requestID, hostID, 1, shared.NewRoomTarget(roomID), uint64(stored.Version))
	if roomErr != nil || hostErr != nil || metaErr != nil {
		writeError(w, http.StatusInternalServerError, "ADMIN_OPERATION_FAILED")
		return
	}
	roomActor, err := h.registry.Acquire(r.Context(), roomID)
	if err != nil {
		writeError(w, http.StatusNotFound, "ROOM_NOT_FOUND")
		return
	}
	defer h.registry.Release(roomID)
	identity := ops.AdminIdentity(r.Context())
	scope := sha256.Sum256([]byte("admin:" + identity))
	_, err = roomActor.Submit(r.Context(), actor.Envelope{
		RequestID: requestID, CommandType: "room.terminate", ScopeHash: scope[:],
		Fingerprint: "room.terminate:" + stored.ID, Command: roomdomain.TerminateRoomCommand{Meta: meta},
		AdminAudit: &gen.AdminAuditLog{Action: "room.terminate", Actor: identity, Target: stored.ID,
			Details: sql.NullString{String: action.Reason, Valid: action.Reason != ""}, CreatedAtMs: time.Now().UnixMilli()},
	})
	if err != nil {
		writeError(w, http.StatusConflict, "ADMIN_OPERATION_FAILED")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) deleteReplay(w http.ResponseWriter, r *http.Request) {
	action, _, ok := decodeAction(w, r)
	if !ok {
		return
	}
	replayID, err := shared.ParseReplayID(chi.URLParam(r, "replayID"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "REPLAY_ID_INVALID")
		return
	}
	identity := ops.AdminIdentity(r.Context())
	audit := gen.AdminAuditLog{Action: "replay.delete", Actor: identity, Target: replayID.String(),
		Details: sql.NullString{String: action.Reason, Valid: action.Reason != ""}, CreatedAtMs: time.Now().UnixMilli()}
	if err := h.repo.DeleteReplayAndAudit(r.Context(), replayID.String(), time.Now().UnixMilli(), audit); err != nil {
		writeError(w, http.StatusNotFound, "REPLAY_NOT_FOUND")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) retirePuzzle(w http.ResponseWriter, r *http.Request) {
	action, _, ok := decodeAction(w, r)
	if !ok {
		return
	}
	puzzleID, err := shared.ParsePuzzleID(chi.URLParam(r, "puzzleID"))
	revision, revisionErr := strconv.ParseInt(chi.URLParam(r, "revision"), 10, 64)
	if err != nil || revisionErr != nil || revision < 1 {
		writeError(w, http.StatusBadRequest, "PUZZLE_ID_INVALID")
		return
	}
	identity := ops.AdminIdentity(r.Context())
	audit := gen.AdminAuditLog{Action: "puzzle.retire", Actor: identity,
		Target:  puzzleID.String() + ":" + strconv.FormatInt(revision, 10),
		Details: sql.NullString{String: action.Reason, Valid: action.Reason != ""}, CreatedAtMs: time.Now().UnixMilli()}
	if err := h.repo.RetirePuzzleAndAudit(r.Context(), puzzleID.String(), revision, audit); err != nil {
		writeError(w, http.StatusConflict, "PUZZLE_RETIRE_FAILED")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeError(w http.ResponseWriter, status int, code string) {
	writeJSON(w, status, map[string]any{"error": map[string]any{"code": code}})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil && !errors.Is(err, http.ErrHandlerTimeout) {
		return
	}
}
