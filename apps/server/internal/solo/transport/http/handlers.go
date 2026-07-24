package http

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	shared "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/domain"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/gen"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/repository"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/platform/idgen"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/puzzle/engine"
	soloproof "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/solo/proof"
)

const maxRecentPuzzles = 50

type Handler struct {
	repo   *repository.Repository
	secret []byte
}

func NewHandler(repo *repository.Repository, secret []byte) *Handler {
	return &Handler{repo: repo, secret: append([]byte(nil), secret...)}
}

func (h *Handler) RegisterRoutes(router chi.Router) {
	router.Post("/api/v1/solo/puzzles", h.CreateAssignment)
	router.Post("/api/v1/solo/attempts/{attemptID}/hint", h.Hint)
	router.Post("/api/v1/solo/attempts/{attemptID}/complete", h.Complete)
}

type assignmentRequest struct {
	Difficulty      string   `json:"difficulty"`
	PlayStyle       string   `json:"playStyle"`
	RecentPuzzleIDs []string `json:"recentPuzzleIds"`
}

type assignmentResponse struct {
	AttemptID          string `json:"attemptId"`
	AssignmentProof    string `json:"assignmentProof"`
	Clues              string `json:"clues"`
	PuzzleID           string `json:"puzzleId"`
	Revision           int64  `json:"revision"`
	Difficulty         string `json:"difficulty"`
	GeneratorVersion   string `json:"generatorVersion"`
	SolverVersion      string `json:"solverVersion"`
	Transformation     string `json:"transformationVersion"`
	TransformationSeed uint64 `json:"transformationSeed"`
	IssuedAtMs         int64  `json:"issuedAtMs"`
}

type boardRequest struct {
	AssignmentProof string `json:"assignmentProof"`
	Values          string `json:"values"`
	Level           string `json:"level,omitempty"`
}

func (h *Handler) CreateAssignment(w http.ResponseWriter, r *http.Request) {
	var request assignmentRequest
	if _, err := shared.ParseRequestID(r.Header.Get("Idempotency-Key")); err != nil {
		writeError(w, http.StatusBadRequest, "REQUEST_ID_INVALID")
		return
	}
	if !decodeJSON(w, r, &request) ||
		(request.PlayStyle != "Guided" && request.PlayStyle != "Classic") ||
		len(request.RecentPuzzleIDs) > maxRecentPuzzles {
		writeError(w, http.StatusBadRequest, "SOLO_REQUEST_INVALID")
		return
	}
	difficulty := request.Difficulty
	if difficulty != "Random" {
		if _, err := shared.ParseDifficulty(difficulty); err != nil {
			writeError(w, http.StatusBadRequest, "SOLO_REQUEST_INVALID")
			return
		}
	} else {
		difficulty = ""
	}
	excluded := make([]string, 0, len(request.RecentPuzzleIDs))
	for _, id := range request.RecentPuzzleIDs {
		if _, err := shared.ParsePuzzleID(id); err != nil {
			writeError(w, http.StatusBadRequest, "SOLO_REQUEST_INVALID")
			return
		}
		excluded = append(excluded, id)
	}
	record, err := h.repo.SelectPuzzleForAssignment(r.Context(), difficulty, excluded, false)
	if err != nil && len(excluded) > 0 {
		record, err = h.repo.SelectPuzzleForAssignment(r.Context(), difficulty, nil, false)
	}
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, "PUZZLE_UNAVAILABLE")
		return
	}
	seed, err := randomUint64()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "PERSISTENCE_FAILED")
		return
	}
	clues, solution, err := transformed(record, seed)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "PUZZLE_INVALID")
		return
	}
	_ = solution
	attemptID, err := idgen.Generator{}.RequestID()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "PERSISTENCE_FAILED")
		return
	}
	now := time.Now().UnixMilli()
	claims := soloproof.Claims{
		Version: soloproof.Version, AttemptID: attemptID.String(), PuzzleID: record.ID,
		Revision: record.Revision, TransformationSeed: seed, IssuedAtMs: now,
		PlayStyle: request.PlayStyle,
	}
	token, err := soloproof.Sign(h.secret, claims)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "PERSISTENCE_FAILED")
		return
	}
	writeJSON(w, http.StatusCreated, assignmentResponse{
		AttemptID: attemptID.String(), AssignmentProof: token, Clues: clues.String(),
		PuzzleID: record.ID, Revision: record.Revision, Difficulty: record.Difficulty,
		GeneratorVersion: record.GeneratorVersion, SolverVersion: record.SolverVersion,
		Transformation: engine.TransformationVersion, TransformationSeed: seed, IssuedAtMs: now,
	})
}

func (h *Handler) Hint(w http.ResponseWriter, r *http.Request) {
	var request boardRequest
	if !decodeJSON(w, r, &request) || (request.Level != "Nudge" && request.Level != "Reveal") {
		writeError(w, http.StatusBadRequest, "SOLO_REQUEST_INVALID")
		return
	}
	claims, clues, solution, values, ok := h.validateBoard(w, r, request)
	if !ok {
		return
	}
	current, err := mergeBoard(clues, values)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "INVALID_VALUE")
		return
	}
	response := map[string]any{"level": request.Level, "penaltyMs": 20_000}
	if request.Level == "Nudge" {
		hint, err := engine.Nudge(current)
		if err != nil {
			writeError(w, http.StatusUnprocessableEntity, "HINT_LEVEL_UNAVAILABLE")
			return
		}
		response["technique"] = hint.Technique.String()
		response["unitKind"] = hint.UnitKind
		response["unitIndex"] = hint.UnitIndex
		response["affectedCells"] = hint.AffectedCells
	} else {
		hint, err := engine.Reveal(current, solution)
		if err != nil {
			writeError(w, http.StatusUnprocessableEntity, "HINT_LEVEL_UNAVAILABLE")
			return
		}
		response["cell"] = hint.Cell
		response["value"] = hint.Value
	}
	_ = claims
	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) Complete(w http.ResponseWriter, r *http.Request) {
	var request boardRequest
	if !decodeJSON(w, r, &request) {
		writeError(w, http.StatusBadRequest, "SOLO_REQUEST_INVALID")
		return
	}
	claims, clues, solution, values, ok := h.validateBoard(w, r, request)
	if !ok {
		return
	}
	incorrect := make([]int, 0)
	full := true
	for index, value := range values {
		if _, fixed := clues.Value(index); fixed {
			continue
		}
		if value == 0 {
			full = false
		} else if expected, _ := solution.Value(index); value != expected {
			incorrect = append(incorrect, index)
		}
	}
	response := map[string]any{"complete": full && len(incorrect) == 0}
	if claims.PlayStyle == "Guided" {
		response["incorrectCells"] = incorrect
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) validateBoard(w http.ResponseWriter, r *http.Request, request boardRequest) (soloproof.Claims, engine.Clues, engine.Solution, []uint8, bool) {
	var emptyClaims soloproof.Claims
	claims, err := soloproof.Verify(h.secret, request.AssignmentProof)
	if err != nil || claims.AttemptID != chi.URLParam(r, "attemptID") {
		writeError(w, http.StatusUnauthorized, "SOLO_ASSIGNMENT_INVALID")
		return emptyClaims, engine.Clues{}, engine.Solution{}, nil, false
	}
	if _, err := shared.ParseRequestID(r.Header.Get("Idempotency-Key")); err != nil {
		writeError(w, http.StatusBadRequest, "REQUEST_ID_INVALID")
		return emptyClaims, engine.Clues{}, engine.Solution{}, nil, false
	}
	record, err := h.repo.GetPuzzle(r.Context(), claims.PuzzleID, claims.Revision)
	if err != nil {
		writeError(w, http.StatusNotFound, "PUZZLE_UNAVAILABLE")
		return emptyClaims, engine.Clues{}, engine.Solution{}, nil, false
	}
	clues, solution, err := transformed(record, claims.TransformationSeed)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "PUZZLE_INVALID")
		return emptyClaims, engine.Clues{}, engine.Solution{}, nil, false
	}
	values, err := parseValues(request.Values)
	if err != nil {
		writeError(w, http.StatusBadRequest, "SOLO_REQUEST_INVALID")
		return emptyClaims, engine.Clues{}, engine.Solution{}, nil, false
	}
	for index, value := range values {
		if clue, fixed := clues.Value(index); fixed && value != clue {
			writeError(w, http.StatusUnprocessableEntity, "CELL_FIXED")
			return emptyClaims, engine.Clues{}, engine.Solution{}, nil, false
		}
	}
	return claims, clues, solution, values, true
}

func transformed(record gen.Puzzle, seed uint64) (engine.Clues, engine.Solution, error) {
	clues, err := engine.ParseClues(decimalGrid(record.Clues))
	if err != nil {
		return engine.Clues{}, engine.Solution{}, err
	}
	solution, err := engine.ParseSolution(decimalGrid(record.Solution))
	if err != nil {
		return engine.Clues{}, engine.Solution{}, err
	}
	return engine.Transform(clues, solution, seed)
}

func mergeBoard(clues engine.Clues, values []uint8) (engine.Clues, error) {
	encoded := []byte(clues.String())
	for index, value := range values {
		if encoded[index] == '0' && value != 0 {
			encoded[index] = '0' + value
		}
	}
	return engine.ParseClues(string(encoded))
}

func parseValues(value string) ([]uint8, error) {
	if len(value) != 81 {
		return nil, errors.New("board must contain 81 digits")
	}
	values := make([]uint8, 81)
	for index, digit := range []byte(value) {
		if digit < '0' || digit > '9' {
			return nil, errors.New("invalid board digit")
		}
		values[index] = digit - '0'
	}
	return values, nil
}

func decimalGrid(values []byte) string {
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

func randomUint64() (uint64, error) {
	var raw [8]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint64(raw[:]) & ((1 << 53) - 1), nil
}

func decodeJSON(w http.ResponseWriter, r *http.Request, target any) bool {
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		return false
	}
	r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(target) == nil
}

func writeError(w http.ResponseWriter, status int, code string) {
	writeJSON(w, status, map[string]any{"error": map[string]any{
		"code": code, "messageKey": "error." + strings.ToLower(code),
		"requestId": "", "details": map[string]any{},
	}})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
