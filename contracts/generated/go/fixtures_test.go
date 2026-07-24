package contracts_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	httpcontract "github.com/drilonrecica/ninefold-sudoku/contracts/generated/go/http"
	realtimecontract "github.com/drilonrecica/ninefold-sudoku/contracts/generated/go/realtime"
	replaycontract "github.com/drilonrecica/ninefold-sudoku/contracts/generated/go/replay"
)

func TestSharedFixturesDecode(t *testing.T) {
	t.Parallel()

	var success httpcontract.SuccessEnvelope
	decodeFixture(t, "http-success.json", &success)
	if success.Version != 9007199254740991 || success.Data.DisplayName != "Éva 🧩" {
		t.Fatalf("unexpected HTTP fixture: %#v", success)
	}

	for _, name := range []string{
		"error-room.json",
		"error-lifecycle.json",
		"error-gameplay.json",
		"error-concurrency.json",
		"error-replay.json",
	} {
		var envelope httpcontract.ErrorEnvelope
		decodeFixture(t, name, &envelope)
		if envelope.Error.Code == "" || envelope.Error.RequestId == "" {
			t.Fatalf("%s lost required values: %#v", name, envelope)
		}
	}

	var client realtimecontract.ClientMessage
	decodeFixture(t, "realtime-client.json", &client)
	if uint64(client.ClientSequence) != 9007199254740991 || client.Target.Kind != realtimecontract.ClientMessageTargetKindMatch {
		t.Fatalf("unexpected realtime fixture: %#v", client)
	}

	var replay replaycontract.ReplayDocument
	decodeFixture(t, "replay.json", &replay)
	if len(replay.Events) != 1 || uint64(replay.Events[0].EventNumber) != 1 {
		t.Fatalf("unexpected replay fixture: %#v", replay)
	}
}

func decodeFixture(t *testing.T, name string, destination any) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "fixtures", name))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, destination); err != nil {
		t.Fatalf("decode %s: %v", name, err)
	}
}
