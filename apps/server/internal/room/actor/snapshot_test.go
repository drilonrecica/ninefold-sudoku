package actor

import (
	"testing"

	shared "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/domain"
	matchdomain "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/match/domain"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/gen"
)

func TestSnapshotRoundTripAndCorruptionFallbackSignal(t *testing.T) {
	match := &matchdomain.Match{
		ID:      shared.MatchID("01900000-0000-7000-8000-000000000001"),
		RoomID:  shared.RoomID("01900000-0000-7000-8000-000000000002"),
		Version: 1,
		State:   shared.MatchActive,
		Values:  map[shared.CellIndex]shared.Digit{3: 7},
	}
	blob, integrity, err := encodeSnapshot(match)
	if err != nil {
		t.Fatal(err)
	}
	snapshot := gen.MatchSnapshot{
		MatchID:          match.ID.String(),
		EventNumber:      1,
		AggregateVersion: 1,
		StateFormat:      snapshotStateFormat,
		StateBlob:        blob,
		IntegrityHash:    integrity,
	}
	decoded, err := decodeSnapshot(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	if decoded.State != shared.MatchActive || decoded.Values[3] != 7 {
		t.Fatalf("unexpected decoded snapshot: %#v", decoded)
	}

	snapshot.StateBlob[0] ^= 0xff
	if _, err := decodeSnapshot(snapshot); err == nil {
		t.Fatal("expected corrupted snapshot to fail integrity validation")
	}
}

func TestEventBufferCoverage(t *testing.T) {
	actor := &Actor{lastEventNumber: 502, eventBuffer: make([]bufferedEvent, 0, eventBufferCapacity)}
	for number := uint64(3); number <= 502; number++ {
		actor.appendBufferedEvent(number, []byte{byte(number)})
	}
	if actor.eventBuffer[0].number != 3 || len(actor.eventBuffer) != eventBufferCapacity {
		t.Fatalf("unexpected buffer boundary: first=%d len=%d", actor.eventBuffer[0].number, len(actor.eventBuffer))
	}
	if actor.eventBuffer[len(actor.eventBuffer)-1].number != 502 {
		t.Fatal("event buffer boundaries are invalid")
	}
}
