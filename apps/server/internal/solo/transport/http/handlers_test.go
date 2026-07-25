package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/gen"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/migrate"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/repository"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/sqlite"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/platform/idgen"
)

const (
	testPuzzleID       = "019f93fe-59f5-7e2d-a543-c697931f766b"
	otherEasyPuzzleID1 = "019f962c-2545-766f-b66f-81f8e200f266"
	otherEasyPuzzleID2 = "019f9b5c-8c94-785b-9947-72c6f5998fde"
)

func TestSoloAssignmentProofAndVisibility(t *testing.T) {
	router, record := testRouter(t)
	assignment := createTestAssignment(t, router, "Guided")
	_, solution, err := transformed(record, assignment.TransformationSeed)
	if err != nil {
		t.Fatal(err)
	}
	if assignment.PuzzleID != testPuzzleID || assignment.Clues == solution.String() ||
		strings.Contains(assignment.AssignmentProof, solution.String()) {
		t.Fatalf("assignment exposed invalid data: %+v", assignment)
	}
	values := []byte(assignment.Clues)
	wrongCell := strings.IndexByte(assignment.Clues, '0')
	expected, _ := solution.Value(wrongCell)
	values[wrongCell] = '0' + ((expected % 9) + 1)
	response := post(t, router, "/api/v1/solo/attempts/"+assignment.AttemptID+"/complete", map[string]any{
		"assignmentProof": assignment.AssignmentProof,
		"values":          string(values),
	})
	if response.Code != http.StatusOK || !bytes.Contains(response.Body.Bytes(), []byte(`"incorrectCells"`)) {
		t.Fatalf("guided check = %d %s", response.Code, response.Body.String())
	}

	tampered := assignment.AssignmentProof[:len(assignment.AssignmentProof)-1] + "A"
	response = post(t, router, "/api/v1/solo/attempts/"+assignment.AttemptID+"/complete", map[string]any{
		"assignmentProof": tampered, "values": assignment.Clues,
	})
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("tampered proof status = %d", response.Code)
	}
	otherID, _ := idgen.Generator{}.RequestID()
	response = post(t, router, "/api/v1/solo/attempts/"+otherID.String()+"/complete", map[string]any{
		"assignmentProof": assignment.AssignmentProof, "values": assignment.Clues,
	})
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("mismatched attempt status = %d", response.Code)
	}
}

func TestClassicHidesSolutionCorrectnessAndRevealIsBounded(t *testing.T) {
	router, _ := testRouter(t)
	assignment := createTestAssignment(t, router, "Classic")
	response := post(t, router, "/api/v1/solo/attempts/"+assignment.AttemptID+"/complete", map[string]any{
		"assignmentProof": assignment.AssignmentProof, "values": assignment.Clues,
	})
	if response.Code != http.StatusOK || bytes.Contains(response.Body.Bytes(), []byte("incorrectCells")) {
		t.Fatalf("classic response = %d %s", response.Code, response.Body.String())
	}
	response = post(t, router, "/api/v1/solo/attempts/"+assignment.AttemptID+"/hint", map[string]any{
		"assignmentProof": assignment.AssignmentProof, "values": assignment.Clues, "level": "Reveal",
	})
	if response.Code != http.StatusOK || !bytes.Contains(response.Body.Bytes(), []byte(`"penaltyMs":20000`)) {
		t.Fatalf("reveal response = %d %s", response.Code, response.Body.String())
	}
}

func testRouter(t *testing.T) (http.Handler, gen.Puzzle) {
	t.Helper()
	db, err := sqlite.New(t.TempDir() + "/solo.db")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := migrate.Up(db.Writer()); err != nil {
		t.Fatal(err)
	}
	repo := repository.New(db)
	record, err := repo.GetPuzzle(context.Background(), testPuzzleID, 1)
	if err != nil {
		t.Fatal(err)
	}
	router := chi.NewRouter()
	NewHandler(repo, []byte(strings.Repeat("s", 32))).RegisterRoutes(router)
	return router, record
}

func createTestAssignment(t *testing.T, router http.Handler, style string) assignmentResponse {
	t.Helper()
	response := post(t, router, "/api/v1/solo/puzzles", map[string]any{
		"difficulty": "Easy", "playStyle": style,
		"recentPuzzleIds": []string{otherEasyPuzzleID1, otherEasyPuzzleID2},
	})
	if response.Code != http.StatusCreated {
		t.Fatalf("assignment = %d %s", response.Code, response.Body.String())
	}
	var assignment assignmentResponse
	if err := json.Unmarshal(response.Body.Bytes(), &assignment); err != nil {
		t.Fatal(err)
	}
	return assignment
}

func post(t *testing.T, router http.Handler, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	encoded, _ := json.Marshal(body)
	request := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(encoded))
	request.Header.Set("Content-Type", "application/json")
	requestID, _ := idgen.Generator{}.RequestID()
	request.Header.Set("Idempotency-Key", requestID.String())
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}
