package http

import (
	"bytes"
	"compress/gzip"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
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
	replayproof "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/replay/proof"
	roomsession "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/room/session"
)

const (
	capabilityLifetime = 7 * 24 * time.Hour
	maxReplayEvents    = 10_000
)

type Handler struct {
	repo     *repository.Repository
	logger   *slog.Logger
	observer Observer
}

type Observer interface {
	ObserveReplayVerificationFailure()
}

func NewHandler(repo *repository.Repository, logger *slog.Logger, observers ...Observer) *Handler {
	handler := &Handler{repo: repo, logger: logger}
	if len(observers) > 0 {
		handler.observer = observers[0]
	}
	return handler
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
	ProofVersion         int            `json:"proofVersion"`
	EventNumber          int64          `json:"eventNumber"`
	AggregateVersion     int64          `json:"aggregateVersion"`
	PublicEventType      string         `json:"publicEventType"`
	PublicActorID        string         `json:"publicActorId"`
	OccurredAtMs         int64          `json:"occurredAtMs"`
	PublicPayload        map[string]any `json:"publicPayload"`
	PrivatePayloadDigest string         `json:"privatePayloadDigest"`
	PreviousEventHash    string         `json:"previousEventHash"`
	EventHash            string         `json:"eventHash"`
}

type replayProof struct {
	ProofVersion     int    `json:"proofVersion"`
	MatchID          string `json:"matchId"`
	FinalEventNumber int64  `json:"finalEventNumber"`
	FinalEventHash   string `json:"finalEventHash"`
	TerminalAtMs     int64  `json:"terminalAtMs"`
	KeyID            string `json:"keyId"`
	Signature        string `json:"signature"`
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
	Proof         replayProof         `json:"proof"`
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
	if err != nil || stored.ReplayID != chi.URLParam(r, "replayID") {
		writeUnavailable(w)
		return
	}
	if stored.RevokedAtMs.Valid {
		writeReplayError(w, http.StatusGone, "REPLAY_DELETED", "error.replay_deleted")
		return
	}
	if stored.ExpiresAtMs <= now {
		writeReplayError(w, http.StatusGone, "REPLAY_EXPIRED", "error.replay_expired")
		return
	}
	document, err := h.projectReplay(r, stored)
	if err != nil {
		if h.observer != nil {
			h.observer.ObserveReplayVerificationFailure()
		}
		h.logger.Warn("replay projection failed", "replayID", stored.ReplayID, "error", err)
		writeUnavailable(w)
		return
	}
	writeCompressedJSON(w, r, document)
}

func (h *Handler) DeleteReplay(w http.ResponseWriter, r *http.Request) {
	setPrivateHeaders(w)
	ctx := r.Context()
	var confirmation struct {
		Confirm bool `json:"confirm"`
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1024)
	if r.Header.Get("Content-Type") != "application/json" ||
		json.NewDecoder(r.Body).Decode(&confirmation) != nil || !confirmation.Confirm {
		writeReplayError(w, http.StatusBadRequest, "CONFIRMATION_REQUIRED", "error.confirmation_required")
		return
	}
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
	previousHash := replayproof.GenesisHash
	for _, row := range rows {
		if row.EventNumber != expected || !bytes.Equal(row.PreviousHash, previousHash) {
			return replayDocument{}, errors.New("event sequence gap")
		}
		var payload map[string]any
		if err := json.Unmarshal([]byte(row.PublicPayloadJson), &payload); err != nil {
			return replayDocument{}, err
		}
		if err := verifyReplayEvent(stored.MatchID, row); err != nil {
			return replayDocument{}, err
		}
		events = append(events, replayEvent{
			ProofVersion: 1, EventNumber: row.EventNumber, AggregateVersion: row.AggregateVersion,
			PublicEventType: row.PublicEventType, PublicActorID: row.PublicActorID.String,
			OccurredAtMs: row.OccurredAtMs, PublicPayload: payload,
			PrivatePayloadDigest: hex.EncodeToString(row.PrivatePayloadDigest),
			PreviousEventHash:    base64.StdEncoding.EncodeToString(row.PreviousHash),
			EventHash:            hex.EncodeToString(row.EventHash),
		})
		previousHash = row.EventHash
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
	seal, err := h.repo.GetReplaySeal(ctx, stored.MatchID)
	if err != nil {
		return replayDocument{}, errors.New("replay seal unavailable")
	}
	if seal.FinalEventNumber != int64(len(rows)) || !bytes.Equal(seal.FinalEventHash, previousHash) {
		return replayDocument{}, errors.New("replay seal boundary mismatch")
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
		Proof: replayProof{
			ProofVersion: 1, MatchID: seal.MatchID, FinalEventNumber: seal.FinalEventNumber,
			FinalEventHash: hex.EncodeToString(seal.FinalEventHash), TerminalAtMs: seal.TerminalAtMs,
			KeyID: seal.SigningKeyID, Signature: base64.StdEncoding.EncodeToString(seal.Signature),
		},
	}, nil
}

func verifyReplayEvent(matchID string, row gen.MatchEvent) error {
	computedHash, err := replayproof.HashEnvelope(replayproof.Envelope{
		ProofVersion: replayproof.Version, MatchID: matchID,
		EventNumber: uint64(row.EventNumber), AggregateVersion: uint64(row.AggregateVersion),
		PublicEventType: row.PublicEventType, PublicActorID: row.PublicActorID.String,
		OccurredAtMs: row.OccurredAtMs, PublicPayload: json.RawMessage(row.PublicPayloadJson),
		PrivatePayloadDigest: hex.EncodeToString(row.PrivatePayloadDigest),
		PreviousEventHash:    row.PreviousHash,
	})
	if err != nil || !bytes.Equal(computedHash, row.EventHash) {
		return errors.New("replay event hash mismatch")
	}
	return nil
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

func writeReplayError(w http.ResponseWriter, status int, code, messageKey string) {
	writeJSON(w, status, map[string]any{"error": map[string]any{
		"code": code, "messageKey": messageKey, "requestId": "", "details": map[string]any{},
	}})
}

func writeCompressedJSON(w http.ResponseWriter, r *http.Request, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Vary", "Accept-Encoding")
	if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		writeJSON(w, http.StatusOK, value)
		return
	}
	w.Header().Set("Content-Encoding", "gzip")
	w.WriteHeader(http.StatusOK)
	writer := gzip.NewWriter(w)
	defer writer.Close()
	_ = json.NewEncoder(writer).Encode(value)
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
