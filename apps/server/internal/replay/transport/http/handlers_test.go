package http

import (
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/gen"
	replayproof "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/replay/proof"
)

func TestVerifyReplayEventRejectsTampering(t *testing.T) {
	const matchID = "01900000-0000-7000-8000-000000000001"
	row := gen.MatchEvent{
		MatchID: matchID, EventNumber: 1, AggregateVersion: 2,
		PublicEventType: "ValuePlaced", PublicActorID: sql.NullString{
			String: "01900000-0000-7000-8000-000000000002", Valid: true,
		},
		OccurredAtMs: 1_800_000_000_000, PublicPayloadJson: `{"cell":4,"value":7}`,
		PrivatePayloadDigest: make([]byte, 32), PreviousHash: replayproof.GenesisHash,
	}
	hash, err := replayproof.HashEnvelope(replayproof.Envelope{
		ProofVersion: replayproof.Version, MatchID: matchID,
		EventNumber: uint64(row.EventNumber), AggregateVersion: uint64(row.AggregateVersion),
		PublicEventType: row.PublicEventType, PublicActorID: row.PublicActorID.String,
		OccurredAtMs: row.OccurredAtMs, PublicPayload: json.RawMessage(row.PublicPayloadJson),
		PrivatePayloadDigest: hex.EncodeToString(row.PrivatePayloadDigest),
		PreviousEventHash:    row.PreviousHash,
	})
	if err != nil {
		t.Fatal(err)
	}
	row.EventHash = hash
	if err := verifyReplayEvent(matchID, row); err != nil {
		t.Fatalf("valid event rejected: %v", err)
	}
	row.PublicPayloadJson = `{"cell":4,"value":8}`
	if err := verifyReplayEvent(matchID, row); err == nil {
		t.Fatal("tampered event verified")
	}
}
