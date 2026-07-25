package http

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"

	shared "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/domain"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/gen"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/repository"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/platform/config"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/platform/idgen"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/room/actor"
	roomdomain "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/room/domain"
	roomsession "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/room/session"
)

const idempotencyKeyHeader = "Idempotency-Key"

// Handler exposes the room lifecycle HTTP endpoints.
type Handler struct {
	repo     *repository.Repository
	registry *actor.Registry
	config   config.Config
	logger   *slog.Logger
	createMu sync.Mutex
	creates  map[string]creationWindow
	lookupMu sync.Mutex
	lookups  map[string]lookupFailure
}

type creationWindow struct {
	start time.Time
	count int
}

type lookupFailure struct {
	count        int
	windowStart  time.Time
	blockedUntil time.Time
}

// NewHandler creates the room HTTP handler.
func NewHandler(repo *repository.Repository, registry *actor.Registry, cfg config.Config, logger *slog.Logger) *Handler {
	return &Handler{
		repo: repo, registry: registry, config: cfg, logger: logger,
		creates: make(map[string]creationWindow), lookups: make(map[string]lookupFailure),
	}
}

// RegisterRoutes wires room endpoints into the supplied router.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/api/v1/rooms", h.CreateRoom)
	r.Get("/api/v1/rooms/{code}", h.PreviewRoom)
	r.Post("/api/v1/rooms/{code}/join", h.JoinRoom)
	r.Post("/api/v1/rooms/{code}/resume", h.ResumeRoom)
	r.Post("/api/v1/rooms/{code}/leave", h.LeaveRoom)
}

// --- request and response types ---

type createRoomRequest struct {
	DisplayName string `json:"displayName"`
	Mode        string `json:"mode"`
	Difficulty  string `json:"difficulty"`
}

type joinRoomRequest struct {
	DisplayName string `json:"displayName"`
	Role        string `json:"role,omitempty"`
}

type leaveRoomRequest struct {
	Intent string `json:"intent"`
}

// --- handlers ---

func (h *Handler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID, err := readRequestID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, shared.ErrRoomStateInvalid, "missing idempotency key")
		return
	}
	var req createRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, shared.ErrRoomStateInvalid, "invalid request body")
		return
	}
	name, err := shared.NewDisplayName(req.DisplayName)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, shared.ErrNameInvalid, "invalid display name")
		return
	}
	mode, err := shared.ParseMode(req.Mode)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, shared.ErrRoomStateInvalid, "unsupported mode")
		return
	}
	difficulty, err := shared.ParseDifficulty(req.Difficulty)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, shared.ErrRoomStateInvalid, "unsupported difficulty")
		return
	}

	if err := h.rejectActiveSession(ctx, r, ""); err != nil {
		writeDomainError(w, err)
		return
	}

	fingerprint := fmt.Sprintf("%s|%s|%s", name.ComparisonKey(), mode, difficulty)
	if resp, ok := h.checkReceipt(ctx, requestID, []byte{}, "CreateRoom", fingerprint); ok {
		writeJSON(w, http.StatusOK, resp)
		return
	}
	if !h.allowRoomCreation(r.RemoteAddr, time.Now()) {
		writeError(w, http.StatusTooManyRequests, shared.ErrRateLimited, "room creation rate limit exceeded")
		return
	}

	roomID, err := idgen.Generator{}.RoomID()
	if err != nil {
		writeError(w, http.StatusInternalServerError, shared.ErrPersistenceFailed, "id generation failed")
		return
	}
	participantID, err := idgen.Generator{}.ParticipantID()
	if err != nil {
		writeError(w, http.StatusInternalServerError, shared.ErrPersistenceFailed, "id generation failed")
		return
	}

	code, err := roomdomain.GenerateCode()
	if err != nil {
		writeError(w, http.StatusInternalServerError, shared.ErrPersistenceFailed, "code generation failed")
		return
	}
	token, err := roomsession.Generate()
	if err != nil {
		writeError(w, http.StatusInternalServerError, shared.ErrPersistenceFailed, "session generation failed")
		return
	}

	host := roomdomain.Participant{
		ID:       shared.ParticipantID(participantID),
		Name:     name,
		Role:     shared.RolePlayer,
		JoinedAt: nowTimestamp(),
	}
	rules := roomdomain.MatchRules{
		Mode:              mode,
		Difficulty:        difficulty,
		ErrorPreset:       shared.ErrorPresetCasual,
		HintsEnabled:      true,
		SharedNotes:       true,
		AutoRemoveNotes:   true,
		SpectatorsAllowed: true,
		RuleVersion:       1,
	}
	room, err := roomdomain.NewRoom(shared.RoomID(roomID), code, host, rules, time.Now())
	if err != nil {
		writeDomainError(w, err)
		return
	}

	tx, txRepo, err := h.repo.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		writeError(w, http.StatusInternalServerError, shared.ErrPersistenceFailed, "transaction failed")
		return
	}
	defer repository.TxRollback(tx)

	genRoom := roomToGen(room)
	genParticipants := []gen.RoomParticipant{participantToGen(host, room.ID)}
	genSession := gen.RoomSession{
		TokenHash:     token.Hash,
		RoomID:        room.ID.String(),
		ParticipantID: host.ID.String(),
		CreatedAtMs:   time.Now().UnixMilli(),
		ExpiresAtMs:   time.Now().Add(7 * 24 * time.Hour).UnixMilli(),
	}
	if err := txRepo.CreateRoomTx(ctx, tx, genRoom, genParticipants, genSession); err != nil {
		if isUniqueViolation(err) {
			writeError(w, http.StatusConflict, shared.ErrDuplicateRequest, "code collision")
			return
		}
		writeError(w, http.StatusInternalServerError, shared.ErrPersistenceFailed, "room creation failed")
		return
	}

	response := buildCreateResponse(room, host.ID)
	if err := h.saveReceipt(ctx, txRepo, tx, requestID, []byte{}, "CreateRoom", fingerprint, response); err != nil {
		writeError(w, http.StatusInternalServerError, shared.ErrPersistenceFailed, "receipt failed")
		return
	}

	if err := repository.TxCommit(tx); err != nil {
		writeError(w, http.StatusInternalServerError, shared.ErrPersistenceFailed, "commit failed")
		return
	}

	// Activate the actor so subsequent commands route through it.
	a := h.registry.Activate(room)
	h.registry.Release(room.ID)

	setRoomCookie(w, token.Value, h.cookieSecure())
	writeJSON(w, http.StatusCreated, response)
	_ = a
}

func (h *Handler) allowRoomCreation(remoteAddress string, now time.Time) bool {
	host, _, err := net.SplitHostPort(remoteAddress)
	if err != nil {
		host = remoteAddress
	}
	h.createMu.Lock()
	defer h.createMu.Unlock()
	mac := hmac.New(sha256.New, h.config.CookieSecret)
	_, _ = mac.Write([]byte(now.UTC().Format("2006-01-02") + "|" + host))
	key := hex.EncodeToString(mac.Sum(nil))
	window := h.creates[key]
	if window.start.IsZero() || now.Sub(window.start) >= time.Hour {
		h.creates[key] = creationWindow{start: now, count: 1}
		return true
	}
	if window.count >= 10 {
		return false
	}
	window.count++
	h.creates[key] = window
	return true
}

func (h *Handler) PreviewRoom(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if retry := h.lookupRetryAfter(r.RemoteAddr, time.Now()); retry > 0 {
		w.Header().Set("Retry-After", strconv.Itoa(retry))
		writeError(w, http.StatusTooManyRequests, shared.ErrRateLimited, "room lookup temporarily blocked")
		return
	}
	codeParam := chi.URLParam(r, "code")
	code, err := shared.ParseRoomCode(codeParam)
	if err != nil {
		w.Header().Set("Retry-After", strconv.Itoa(h.recordFailedLookup(r.RemoteAddr, time.Now())))
		writeError(w, http.StatusBadRequest, shared.ErrRoomNotFound, "invalid room code")
		return
	}
	gr, err := h.repo.GetRoomByCode(ctx, code.String())
	if err != nil {
		if repository.IsNoRows(err) {
			w.Header().Set("Retry-After", strconv.Itoa(h.recordFailedLookup(r.RemoteAddr, time.Now())))
			writeError(w, http.StatusNotFound, shared.ErrRoomNotFound, "room not found")
			return
		}
		writeError(w, http.StatusInternalServerError, shared.ErrPersistenceFailed, "database error")
		return
	}
	if isTerminal(gr.State) {
		w.Header().Set("Retry-After", strconv.Itoa(h.recordFailedLookup(r.RemoteAddr, time.Now())))
		writeError(w, http.StatusNotFound, shared.ErrRoomNotFound, "room not found")
		return
	}
	h.clearLookupFailures(r.RemoteAddr, time.Now())
	participants, err := h.repo.ListActiveRoomParticipants(ctx, gr.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, shared.ErrPersistenceFailed, "database error")
		return
	}
	players, spectators := countParticipants(participants)
	cap := modeCapacity(shared.Mode(gr.Mode))

	preview := map[string]any{
		"mode":                 gr.Mode,
		"difficulty":           gr.Difficulty,
		"state":                gr.State,
		"locked":               gr.SpectatorsAllowed == 0,
		"playerSeatsTotal":     cap.MaxPlayers,
		"playerSeatsAvailable": max(0, cap.MaxPlayers-players),
		"spectatorSeatsTotal":  cap.MaxSpectators,
		"spectatorSeatsAvailable": func() int {
			if gr.SpectatorsAllowed == 0 {
				return 0
			}
			return max(0, cap.MaxSpectators-spectators)
		}(),
	}
	previewBody, _ := json.Marshal(preview)
	writeJSON(w, http.StatusOK, previewBody)
}

func (h *Handler) lookupKey(remoteAddress string, now time.Time) string {
	host, _, err := net.SplitHostPort(remoteAddress)
	if err != nil {
		host = remoteAddress
	}
	mac := hmac.New(sha256.New, h.config.CookieSecret)
	_, _ = mac.Write([]byte(now.UTC().Format("2006-01-02") + "|" + host))
	return hex.EncodeToString(mac.Sum(nil))
}

func (h *Handler) lookupRetryAfter(remoteAddress string, now time.Time) int {
	h.lookupMu.Lock()
	defer h.lookupMu.Unlock()
	failure := h.lookups[h.lookupKey(remoteAddress, now)]
	if !failure.blockedUntil.After(now) {
		return 0
	}
	return max(1, int(failure.blockedUntil.Sub(now).Seconds()))
}

func (h *Handler) recordFailedLookup(remoteAddress string, now time.Time) int {
	h.lookupMu.Lock()
	defer h.lookupMu.Unlock()
	key := h.lookupKey(remoteAddress, now)
	if len(h.lookups) >= 10_000 {
		h.lookups = make(map[string]lookupFailure)
	}
	failure := h.lookups[key]
	if failure.windowStart.IsZero() || now.Sub(failure.windowStart) >= 10*time.Minute {
		failure = lookupFailure{windowStart: now}
	}
	failure.count++
	delay := min(30, 1<<min(failure.count-1, 5))
	if failure.count >= 5 {
		failure.blockedUntil = now.Add(time.Duration(delay) * time.Second)
	}
	h.lookups[key] = failure
	return delay
}

func (h *Handler) clearLookupFailures(remoteAddress string, now time.Time) {
	h.lookupMu.Lock()
	delete(h.lookups, h.lookupKey(remoteAddress, now))
	h.lookupMu.Unlock()
}

func (h *Handler) JoinRoom(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID, err := readRequestID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, shared.ErrRoomStateInvalid, "missing idempotency key")
		return
	}
	codeParam := chi.URLParam(r, "code")
	code, err := shared.ParseRoomCode(codeParam)
	if err != nil {
		writeError(w, http.StatusBadRequest, shared.ErrRoomNotFound, "invalid room code")
		return
	}
	var req joinRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, shared.ErrRoomStateInvalid, "invalid request body")
		return
	}
	name, err := shared.NewDisplayName(req.DisplayName)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, shared.ErrNameInvalid, "invalid display name")
		return
	}
	role := shared.RolePlayer
	if req.Role != "" {
		role, err = shared.ParseParticipationRole(req.Role)
		if err != nil {
			writeError(w, http.StatusUnprocessableEntity, shared.ErrRoleChangeInvalid, "invalid role")
			return
		}
	}

	existing, err := h.activeSession(r)
	if err != nil {
		writeError(w, http.StatusInternalServerError, shared.ErrSessionInvalid, "session lookup failed")
		return
	}
	if existing != nil && existing.RoomID != grFromCode(ctx, h.repo, code) {
		details := map[string]any{}
		if currentRoom, lookupErr := h.repo.GetRoomByID(ctx, existing.RoomID); lookupErr == nil {
			details["roomCode"] = currentRoom.Code
		}
		writeErrorWithDetails(w, http.StatusConflict, shared.ErrActiveRoomSessionExists, details)
		return
	}
	if existing != nil && existing.RoomID == grFromCode(ctx, h.repo, code) {
		// Treat as resume for the same room.
		h.rotateAndResume(ctx, w, r, existing)
		return
	}

	gr, err := h.repo.GetRoomByCode(ctx, code.String())
	if err != nil {
		if repository.IsNoRows(err) {
			writeError(w, http.StatusNotFound, shared.ErrRoomNotFound, "room not found")
			return
		}
		writeError(w, http.StatusInternalServerError, shared.ErrPersistenceFailed, "database error")
		return
	}
	if isTerminal(gr.State) {
		writeError(w, http.StatusNotFound, shared.ErrRoomNotFound, "room not found")
		return
	}

	fingerprint := fmt.Sprintf("%s|%s|%s", code.String(), name.ComparisonKey(), role)
	if resp, ok := h.checkReceipt(ctx, requestID, nil, "RequestJoin", fingerprint); ok {
		writeJSON(w, http.StatusOK, resp)
		return
	}

	participantID, err := idgen.Generator{}.ParticipantID()
	if err != nil {
		writeError(w, http.StatusInternalServerError, shared.ErrPersistenceFailed, "id generation failed")
		return
	}
	token, err := roomsession.Generate()
	if err != nil {
		writeError(w, http.StatusInternalServerError, shared.ErrPersistenceFailed, "session generation failed")
		return
	}

	cmd := roomdomain.RequestJoinCommand{
		Meta: shared.CommandMetadata{
			RequestID:                  requestID,
			AuthenticatedParticipantID: shared.ParticipantID(participantID),
			ClientSequence:             1,
			Target:                     shared.NewRoomTarget(shared.RoomID(gr.ID)),
			ExpectedVersion:            uint64(gr.Version),
		},
		Code:          code,
		DisplayName:   name,
		Role:          role,
		ParticipantID: shared.ParticipantID(participantID),
	}

	a, err := h.registry.Acquire(ctx, shared.RoomID(gr.ID))
	if err != nil {
		writeError(w, http.StatusInternalServerError, shared.ErrPersistenceFailed, "actor unavailable")
		return
	}
	defer h.registry.Release(shared.RoomID(gr.ID))

	result, err := a.Submit(ctx, actor.Envelope{
		RequestID:      requestID,
		CommandType:    "RequestJoin",
		ScopeHash:      token.Hash,
		Fingerprint:    fingerprint,
		Command:        cmd,
		NewSessionHash: token.Hash,
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}

	setRoomCookie(w, token.Value, h.cookieSecure())
	writeJSON(w, http.StatusOK, result.Response)
}

func (h *Handler) ResumeRoom(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, err := readRequestID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, shared.ErrRoomStateInvalid, "missing idempotency key")
		return
	}
	existing, err := h.activeSession(r)
	if err != nil || existing == nil {
		writeError(w, http.StatusUnauthorized, shared.ErrSessionInvalid, "invalid session")
		return
	}
	codeParam := chi.URLParam(r, "code")
	code, err := shared.ParseRoomCode(codeParam)
	if err != nil || existing.RoomID != grFromCode(ctx, h.repo, code) {
		writeError(w, http.StatusNotFound, shared.ErrRoomNotFound, "room not found")
		return
	}
	h.rotateAndResume(ctx, w, r, existing)
}

func (h *Handler) LeaveRoom(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID, err := readRequestID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, shared.ErrRoomStateInvalid, "missing idempotency key")
		return
	}
	existing, err := h.activeSession(r)
	if err != nil || existing == nil {
		writeError(w, http.StatusUnauthorized, shared.ErrSessionInvalid, "invalid session")
		return
	}
	codeParam := chi.URLParam(r, "code")
	code, err := shared.ParseRoomCode(codeParam)
	if err != nil {
		writeError(w, http.StatusBadRequest, shared.ErrRoomNotFound, "invalid room code")
		return
	}
	gr, err := h.repo.GetRoomByCode(ctx, code.String())
	if err != nil {
		if repository.IsNoRows(err) {
			writeError(w, http.StatusNotFound, shared.ErrRoomNotFound, "room not found")
			return
		}
		writeError(w, http.StatusInternalServerError, shared.ErrPersistenceFailed, "database error")
		return
	}
	if existing.RoomID != gr.ID {
		writeError(w, http.StatusNotFound, shared.ErrRoomNotFound, "room not found")
		return
	}

	var req leaveRoomRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	fingerprint := fmt.Sprintf("%s|%s", existing.ParticipantID, req.Intent)
	if resp, ok := h.checkReceipt(ctx, requestID, existing.TokenHash, "LeaveRoom", fingerprint); ok {
		clearRoomCookie(w, h.cookieSecure())
		writeJSON(w, http.StatusOK, resp)
		return
	}

	cmd := roomdomain.LeaveRoomCommand{
		Meta: shared.CommandMetadata{
			RequestID:                  requestID,
			AuthenticatedParticipantID: shared.ParticipantID(existing.ParticipantID),
			ClientSequence:             1,
			Target:                     shared.NewRoomTarget(shared.RoomID(gr.ID)),
			ExpectedVersion:            uint64(gr.Version),
		},
		Intent: req.Intent,
	}

	a, err := h.registry.Acquire(ctx, shared.RoomID(gr.ID))
	if err != nil {
		writeError(w, http.StatusInternalServerError, shared.ErrPersistenceFailed, "actor unavailable")
		return
	}
	defer h.registry.Release(shared.RoomID(gr.ID))

	_, err = a.Submit(ctx, actor.Envelope{
		RequestID:   requestID,
		CommandType: "LeaveRoom",
		ScopeHash:   existing.TokenHash,
		Fingerprint: fingerprint,
		Command:     cmd,
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}

	clearRoomCookie(w, h.cookieSecure())
	w.WriteHeader(http.StatusNoContent)
}

// --- helpers ---

func (h *Handler) rotateAndResume(ctx context.Context, w http.ResponseWriter, r *http.Request, existing *gen.RoomSession) {
	newToken, err := roomsession.Generate()
	if err != nil {
		writeError(w, http.StatusInternalServerError, shared.ErrPersistenceFailed, "session generation failed")
		return
	}
	gr, err := h.repo.GetRoomByID(ctx, existing.RoomID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, shared.ErrPersistenceFailed, "database error")
		return
	}
	participants, err := h.repo.ListActiveRoomParticipants(ctx, gr.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, shared.ErrPersistenceFailed, "database error")
		return
	}
	room, err := roomFromGen(gr, participants)
	if err != nil {
		writeError(w, http.StatusInternalServerError, shared.ErrPersistenceFailed, "mapping error")
		return
	}

	tx, txRepo, err := h.repo.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		writeError(w, http.StatusInternalServerError, shared.ErrPersistenceFailed, "transaction failed")
		return
	}
	defer repository.TxRollback(tx)
	if err := txRepo.RevokeRoomSession(ctx, tx, existing.TokenHash, time.Now().UnixMilli()); err != nil {
		writeError(w, http.StatusInternalServerError, shared.ErrPersistenceFailed, "session rotation failed")
		return
	}
	if err := txRepo.CreateRoomSessionTx(ctx, tx, gen.RoomSession{
		TokenHash:     newToken.Hash,
		RoomID:        existing.RoomID,
		ParticipantID: existing.ParticipantID,
		CreatedAtMs:   time.Now().UnixMilli(),
		ExpiresAtMs:   time.Now().Add(7 * 24 * time.Hour).UnixMilli(),
	}); err != nil {
		writeError(w, http.StatusInternalServerError, shared.ErrPersistenceFailed, "session rotation failed")
		return
	}
	if err := repository.TxCommit(tx); err != nil {
		writeError(w, http.StatusInternalServerError, shared.ErrPersistenceFailed, "commit failed")
		return
	}

	setRoomCookie(w, newToken.Value, h.cookieSecure())
	writeJSON(w, http.StatusOK, buildResumeResponse(room, shared.ParticipantID(existing.ParticipantID)))
}

func (h *Handler) activeSession(r *http.Request) (*gen.RoomSession, error) {
	ctx := r.Context()
	token := roomsession.Read(r)
	if token == "" {
		return nil, nil
	}
	hash := roomsession.Hash(token)
	session, err := h.repo.GetRoomSessionByHash(ctx, hash)
	if err != nil {
		if repository.IsNoRows(err) {
			return nil, nil
		}
		return nil, err
	}
	now := time.Now().UnixMilli()
	if session.ExpiresAtMs < now {
		return nil, nil
	}
	if session.RevokedAtMs.Valid && session.RevokedAtMs.Int64 <= now {
		return nil, nil
	}
	return &session, nil
}

func (h *Handler) rejectActiveSession(ctx context.Context, r *http.Request, allowedRoomID string) error {
	existing, err := h.activeSession(r)
	if err != nil {
		return err
	}
	if existing == nil {
		return nil
	}
	if allowedRoomID != "" && existing.RoomID == allowedRoomID {
		return nil
	}
	details := map[string]any{}
	if room, lookupErr := h.repo.GetRoomByID(ctx, existing.RoomID); lookupErr == nil {
		details["roomCode"] = room.Code
	}
	return shared.DomainError{Code: shared.ErrActiveRoomSessionExists, Details: details}
}

func (h *Handler) checkReceipt(ctx context.Context, requestID shared.RequestID, scopeHash []byte, commandType, fingerprint string) ([]byte, bool) {
	receipt, err := h.repo.GetCommandReceipt(ctx, requestID.String())
	if err != nil {
		return nil, false
	}
	if !bytesEqual(receipt.AuthenticatedScopeHash, scopeHash) || receipt.CommandType != commandType || receipt.RequestFingerprint != fingerprint {
		return nil, false
	}
	if !receipt.SafeResponseJson.Valid {
		return nil, true
	}
	return []byte(receipt.SafeResponseJson.String), true
}

func (h *Handler) saveReceipt(ctx context.Context, txRepo *repository.Repository, tx *sql.Tx, requestID shared.RequestID, scopeHash []byte, commandType, fingerprint string, response []byte) error {
	receipt := gen.CommandReceipt{
		RequestID:              requestID.String(),
		AuthenticatedScopeHash: scopeHash,
		CommandType:            commandType,
		RequestFingerprint:     fingerprint,
		TerminalStatus:         "ok",
		SafeResponseJson:       sql.NullString{String: string(response), Valid: true},
		CreatedAtMs:            time.Now().UnixMilli(),
		ExpiresAtMs:            time.Now().Add(24 * time.Hour).UnixMilli(),
	}
	return txRepo.CreateCommandReceipt(ctx, tx, receipt)
}

func (h *Handler) cookieSecure() bool {
	return h.config.Environment == config.Production || h.config.PublicURL.Scheme == "https"
}

// --- response builders ---

func buildCreateResponse(room *roomdomain.Room, selfID shared.ParticipantID) []byte {
	return buildRoomResponse(room, selfID)
}

func buildResumeResponse(room *roomdomain.Room, selfID shared.ParticipantID) []byte {
	return buildRoomResponse(room, selfID)
}

func buildRoomResponse(room *roomdomain.Room, selfID shared.ParticipantID) []byte {
	view := map[string]any{
		"room": map[string]any{
			"id":             room.ID.String(),
			"code":           room.Code.String(),
			"state":          string(room.State),
			"version":        uint64(room.Version),
			"hostId":         nullParticipantID(room.HostParticipantID),
			"currentMatchId": nullMatchID(room.CurrentMatchID),
			"participants":   participantsToView(room.Participants),
			"settings": map[string]any{
				"mode":              string(room.Rules.Mode),
				"difficulty":        string(room.Rules.Difficulty),
				"errorPreset":       string(room.Rules.ErrorPreset),
				"hintsEnabled":      room.Rules.HintsEnabled,
				"sharedNotes":       room.Rules.SharedNotes,
				"autoRemoveNotes":   room.Rules.AutoRemoveNotes,
				"spectatorsAllowed": room.Rules.SpectatorsAllowed,
			},
		},
		"participants": participantsToView(room.Participants),
		"self": map[string]any{
			"participantId": selfID.String(),
		},
	}
	if room.Countdown != nil {
		view["countdown"] = map[string]any{
			"matchId":    room.Countdown.MatchID.String(),
			"generation": room.Countdown.Generation,
			"deadlineAt": room.Countdown.DeadlineAt.Milliseconds(),
		}
	}
	b, _ := json.Marshal(view)
	return b
}

// --- transport helpers ---

func readRequestID(r *http.Request) (shared.RequestID, error) {
	raw := r.Header.Get(idempotencyKeyHeader)
	if raw == "" {
		return "", errors.New("missing idempotency key")
	}
	return shared.ParseRequestID(raw)
}

func setRoomCookie(w http.ResponseWriter, value string, secure bool) {
	http.SetCookie(w, roomsession.Cookie(value, secure, time.Now().Add(7*24*time.Hour)))
}

func clearRoomCookie(w http.ResponseWriter, secure bool) {
	http.SetCookie(w, roomsession.ClearCookie(secure))
}

func writeJSON(w http.ResponseWriter, status int, body []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(body)
}

func writeError(w http.ResponseWriter, status int, code shared.ErrorCode, message string) {
	writeErrorWithDetails(w, status, code, map[string]any{"message": message})
}

func writeErrorWithDetails(w http.ResponseWriter, status int, code shared.ErrorCode, details map[string]any) {
	envelope := map[string]any{
		"error": map[string]any{
			"code":       string(code),
			"messageKey": "error." + strings.ToLower(string(code)),
			"requestId":  "",
			"details":    details,
		},
	}
	b, _ := json.Marshal(envelope)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(b)
}

func writeDomainError(w http.ResponseWriter, err error) {
	var domainErr shared.DomainError
	if errors.As(err, &domainErr) {
		status := http.StatusUnprocessableEntity
		switch domainErr.Code {
		case shared.ErrRoomNotFound:
			status = http.StatusNotFound
		case shared.ErrSessionInvalid, shared.ErrSessionExpired:
			status = http.StatusUnauthorized
		case shared.ErrNotRoomHost, shared.ErrActionNotAllowedForRole:
			status = http.StatusForbidden
		case shared.ErrActiveRoomSessionExists, shared.ErrStaleVersion, shared.ErrDuplicateRequest:
			status = http.StatusConflict
		case shared.ErrRoomLocked, shared.ErrRoomFull, shared.ErrSpectatorCapacityReached:
			status = http.StatusConflict
		case shared.ErrServerBusy:
			status = http.StatusServiceUnavailable
		}
		details := domainErr.Details
		if details == nil {
			details = map[string]any{}
		}
		writeErrorWithDetails(w, status, domainErr.Code, details)
		return
	}
	writeError(w, http.StatusInternalServerError, shared.ErrPersistenceFailed, err.Error())
}

func bytesEqual(a, b []byte) bool {
	return sha256.Sum256(a) == sha256.Sum256(b)
}

func nowTimestamp() shared.Timestamp {
	ts, _ := shared.NewTimestamp(time.Now())
	return ts
}

func isTerminal(state string) bool {
	switch state {
	case string(shared.RoomExpired), string(shared.RoomCancelled), string(shared.RoomTerminatedByAdmin):
		return true
	}
	return false
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "UNIQUE constraint failed")
}

func grFromCode(ctx context.Context, repo *repository.Repository, code shared.RoomCode) string {
	gr, err := repo.GetRoomByCode(ctx, code.String())
	if err != nil {
		return ""
	}
	return gr.ID
}

func countParticipants(participants []gen.RoomParticipant) (players, spectators int) {
	for _, p := range participants {
		if p.Role == string(shared.RolePlayer) {
			players++
		} else {
			spectators++
		}
	}
	return
}

func modeCapacity(mode shared.Mode) capacity {
	switch mode {
	case shared.ModeCoop:
		return capacity{MaxPlayers: 6, MinPlayers: 1, MaxSpectators: 10}
	default:
		return capacity{MaxPlayers: 6, MinPlayers: 1, MaxSpectators: 10}
	}
}

type capacity struct {
	MaxPlayers    int
	MinPlayers    int
	MaxSpectators int
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// --- local mapping helpers ---

func roomToGen(room *roomdomain.Room) gen.Room {
	hostID := ""
	if room.HostParticipantID != nil {
		hostID = room.HostParticipantID.String()
	}
	return gen.Room{
		ID:                room.ID.String(),
		Code:              room.Code.String(),
		State:             string(room.State),
		Version:           int64(room.Version),
		Mode:              string(room.Rules.Mode),
		Difficulty:        string(room.Rules.Difficulty),
		ErrorPreset:       string(room.Rules.ErrorPreset),
		HintsEnabled:      boolToInt(room.Rules.HintsEnabled),
		SharedNotes:       boolToInt(room.Rules.SharedNotes),
		AutoRemoveNotes:   boolToInt(room.Rules.AutoRemoveNotes),
		SpectatorsAllowed: boolToInt(room.Rules.SpectatorsAllowed),
		HostParticipantID: hostID,
		CurrentMatchID:    nullMatchID(room.CurrentMatchID),
		CreatedAtMs:       room.CreatedAt.Milliseconds(),
		LastActivityAtMs:  room.LastActivityAt.Milliseconds(),
		ExpiresAtMs:       room.ExpiresAt.Milliseconds(),
	}
}

func participantToGen(p roomdomain.Participant, roomID shared.RoomID) gen.RoomParticipant {
	return gen.RoomParticipant{
		ID:            p.ID.String(),
		RoomID:        roomID.String(),
		DisplayName:   p.Name.String(),
		Role:          string(p.Role),
		IsHost:        boolToInt(p.IsHost),
		IsReady:       boolToInt(p.IsReady),
		JoinedAtMs:    p.JoinedAt.Milliseconds(),
		LeftAtMs:      nullInt64(p.LeftAt),
		RemovedAtMs:   nullInt64(p.RemovedAt),
		RemovedReason: nullString(p.RemovedReason),
	}
}

func roomFromGen(gr gen.Room, participants []gen.RoomParticipant) (*roomdomain.Room, error) {
	mode, err := shared.ParseMode(gr.Mode)
	if err != nil {
		return nil, err
	}
	difficulty, err := shared.ParseDifficulty(gr.Difficulty)
	if err != nil {
		return nil, err
	}
	errorPreset, err := shared.ParseErrorPreset(gr.ErrorPreset)
	if err != nil {
		return nil, err
	}
	createdAt, err := shared.TimestampFromMilliseconds(gr.CreatedAtMs)
	if err != nil {
		return nil, err
	}
	lastActivity, err := shared.TimestampFromMilliseconds(gr.LastActivityAtMs)
	if err != nil {
		return nil, err
	}
	expiresAt, err := shared.TimestampFromMilliseconds(gr.ExpiresAtMs)
	if err != nil {
		return nil, err
	}
	room := &roomdomain.Room{
		ID:    shared.RoomID(gr.ID),
		Code:  shared.RoomCode(gr.Code),
		State: shared.RoomState(gr.State),
		Rules: roomdomain.MatchRules{
			Mode:              mode,
			Difficulty:        difficulty,
			ErrorPreset:       errorPreset,
			HintsEnabled:      gr.HintsEnabled != 0,
			SharedNotes:       gr.SharedNotes != 0,
			AutoRemoveNotes:   gr.AutoRemoveNotes != 0,
			SpectatorsAllowed: gr.SpectatorsAllowed != 0,
		},
		CreatedAt:      createdAt,
		LastActivityAt: lastActivity,
		ExpiresAt:      expiresAt,
	}
	room.Version = shared.RoomVersion(gr.Version)
	if gr.HostParticipantID != "" {
		id := shared.ParticipantID(gr.HostParticipantID)
		room.HostParticipantID = &id
	}
	if gr.CurrentMatchID.Valid {
		id := shared.MatchID(gr.CurrentMatchID.String)
		room.CurrentMatchID = &id
	}
	for _, gp := range participants {
		p, err := participantFromGen(gp)
		if err != nil {
			return nil, err
		}
		room.Participants = append(room.Participants, p)
	}
	return room, nil
}

func participantFromGen(gp gen.RoomParticipant) (roomdomain.Participant, error) {
	name, err := shared.NewDisplayName(gp.DisplayName)
	if err != nil {
		return roomdomain.Participant{}, err
	}
	role, err := shared.ParseParticipationRole(gp.Role)
	if err != nil {
		return roomdomain.Participant{}, err
	}
	joinedAt, err := shared.TimestampFromMilliseconds(gp.JoinedAtMs)
	if err != nil {
		return roomdomain.Participant{}, err
	}
	p := roomdomain.Participant{
		ID:       shared.ParticipantID(gp.ID),
		Name:     name,
		Role:     role,
		IsHost:   gp.IsHost != 0,
		IsReady:  gp.IsReady != 0,
		JoinedAt: joinedAt,
	}
	if gp.LeftAtMs.Valid {
		ts, err := shared.TimestampFromMilliseconds(gp.LeftAtMs.Int64)
		if err != nil {
			return roomdomain.Participant{}, err
		}
		p.LeftAt = &ts
	}
	if gp.RemovedAtMs.Valid {
		ts, err := shared.TimestampFromMilliseconds(gp.RemovedAtMs.Int64)
		if err != nil {
			return roomdomain.Participant{}, err
		}
		p.RemovedAt = &ts
	}
	if gp.RemovedReason.Valid {
		p.RemovedReason = gp.RemovedReason.String
	}
	return p, nil
}

func participantsToView(parts []roomdomain.Participant) []map[string]any {
	out := make([]map[string]any, 0, len(parts))
	for _, p := range parts {
		if !p.IsActive() {
			continue
		}
		out = append(out, map[string]any{
			"id":       p.ID.String(),
			"name":     p.Name.String(),
			"role":     string(p.Role),
			"isHost":   p.IsHost,
			"isReady":  p.IsReady,
			"joinedAt": p.JoinedAt.Milliseconds(),
		})
	}
	return out
}

func nullParticipantID(id *shared.ParticipantID) any {
	if id == nil {
		return nil
	}
	return id.String()
}

func nullMatchID(id *shared.MatchID) sql.NullString {
	if id == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: id.String(), Valid: true}
}

func nullInt64(ts *shared.Timestamp) sql.NullInt64 {
	if ts == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: ts.Milliseconds(), Valid: true}
}

func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func boolToInt(b bool) int64 {
	if b {
		return 1
	}
	return 0
}
