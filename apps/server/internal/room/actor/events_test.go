package actor

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	shared "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/domain"
	matchdomain "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/match/domain"
)

func TestReplayHashFixture(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "..", "contracts", "fixtures", "replay-hash-chain.json"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	var fixture struct {
		Envelope  matchEventPublicEnvelope `json:"envelope"`
		EventHash string                   `json:"eventHash"`
	}
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatalf("decode fixture: %v", err)
	}
	got, err := hashEnvelope(fixture.Envelope)
	if err != nil {
		t.Fatalf("hash envelope: %v", err)
	}
	if hex.EncodeToString(got) != fixture.EventHash {
		t.Fatalf("hash mismatch: got %x want %s", got, fixture.EventHash)
	}
}

func TestNotePayloadSerializesDigitsAsNumbers(t *testing.T) {
	payload, err := matchEventPublicPayload(matchdomain.NotesAddedEvent{
		Cell:   shared.CellIndex(4),
		Digits: []shared.Digit{2, 7},
	})
	if err != nil {
		t.Fatalf("serialize note event: %v", err)
	}
	if string(payload) != `{"schemaVersion":1,"cell":4,"digits":[2,7],"participantId":""}` {
		t.Fatalf("note payload = %s", payload)
	}
}

func TestMatchViewNeverContainsSolution(t *testing.T) {
	match := &matchdomain.Match{
		ID:            shared.MatchID("01900000-0000-7000-8000-000000000010"),
		Mistakes:      map[shared.ParticipantID]uint32{},
		Contributions: map[shared.ParticipantID]uint32{},
		Puzzle: shared.AssignedPuzzle{
			Solution: bytes.Repeat([]byte{9}, 81),
		},
	}
	encoded, err := json.Marshal(buildMatchView(match))
	if err != nil {
		t.Fatalf("encode match view: %v", err)
	}
	if bytes.Contains(encoded, []byte("solution")) || bytes.Contains(encoded, bytes.Repeat([]byte("9"), 20)) {
		t.Fatalf("match view exposed a standalone solution artifact")
	}
}
