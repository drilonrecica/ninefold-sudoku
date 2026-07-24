package websocket

import (
	"testing"
	"time"

	"github.com/drilonrecica/ninefold-sudoku/contracts/generated/go/realtime"
)

func TestRealtimeRateLimits(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	c := &connection{rateWindows: make(map[string]rateWindow)}
	for i := 0; i < 10; i++ {
		if !c.allowMessage(realtime.ClientMessageTypeMatchPlaceValue, now) {
			t.Fatalf("value command %d should be allowed", i+1)
		}
	}
	if c.allowMessage(realtime.ClientMessageTypeMatchPlaceValue, now) {
		t.Fatal("eleventh value command in one second should be rate limited")
	}
	if !c.allowMessage(realtime.ClientMessageTypeMatchPlaceValue, now.Add(time.Second)) {
		t.Fatal("value window should reset")
	}

	for i := 0; i < 5; i++ {
		if !c.allowMessage(realtime.ClientMessageTypeMatchReaction, now) {
			t.Fatalf("social command %d should be allowed", i+1)
		}
	}
	if c.allowMessage(realtime.ClientMessageTypeMatchPing, now) {
		t.Fatal("ping and reaction must share the social limit")
	}
}
