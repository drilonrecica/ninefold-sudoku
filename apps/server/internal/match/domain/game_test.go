package domain

import (
	"testing"
	"time"

	shared "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/domain"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/platform/idgen"
)

// A valid completed Sudoku solution used for deterministic tests.
var testSolution = []byte{
	1, 2, 3, 4, 5, 6, 7, 8, 9,
	4, 5, 6, 7, 8, 9, 1, 2, 3,
	7, 8, 9, 1, 2, 3, 4, 5, 6,
	2, 3, 4, 5, 6, 7, 8, 9, 1,
	5, 6, 7, 8, 9, 1, 2, 3, 4,
	8, 9, 1, 2, 3, 4, 5, 6, 7,
	3, 4, 5, 6, 7, 8, 9, 1, 2,
	6, 7, 8, 9, 1, 2, 3, 4, 5,
	9, 1, 2, 3, 4, 5, 6, 7, 8,
}

func testClues(filled int) []byte {
	clues := make([]byte, 81)
	for i := 0; i < 81; i++ {
		if i < filled {
			clues[i] = testSolution[i]
		}
	}
	return clues
}

func makeMatch(t testing.TB, preset shared.ErrorPreset, clues []byte) (*Match, shared.ParticipantID, time.Time) {
	t.Helper()
	gen := idgen.Generator{}
	matchID, err := gen.MatchID()
	if err != nil {
		t.Fatal(err)
	}
	roomID, err := gen.RoomID()
	if err != nil {
		t.Fatal(err)
	}
	participantID, err := gen.ParticipantID()
	if err != nil {
		t.Fatal(err)
	}
	puzzleID, err := gen.PuzzleID()
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	m, _, err := NewPrepared(matchID, roomID, Rules{
		Mode:            shared.ModeCoop,
		Difficulty:      shared.DifficultyMedium,
		ErrorPreset:     preset,
		HintsEnabled:    true,
		AutoRemoveNotes: true,
		RuleVersion:     1,
	}, shared.AssignedPuzzle{
		PuzzleID:         puzzleID,
		Revision:         1,
		Difficulty:       shared.DifficultyMedium,
		GeneratorVersion: "1",
		SolverVersion:    "1",
		Clues:            clues,
		Solution:         testSolution,
	}, []shared.ParticipantID{participantID}, now)
	if err != nil {
		t.Fatal(err)
	}
	return m, participantID, now
}

func cmdMeta(t testing.TB, m *Match, participantID shared.ParticipantID) shared.CommandMetadata {
	t.Helper()
	gen := idgen.Generator{}
	reqID, err := gen.RequestID()
	if err != nil {
		t.Fatal(err)
	}
	seq, err := shared.NewClientSequence(1)
	if err != nil {
		t.Fatal(err)
	}
	meta, err := shared.NewCommandMetadata(reqID, participantID, seq, shared.NewMatchTarget(m.ID), uint64(m.Version))
	if err != nil {
		t.Fatal(err)
	}
	return meta
}

func TestActivateStartsMatch(t *testing.T) {
	m, _, now := makeMatch(t, shared.ErrorPresetCasual, testClues(0))
	if m.State != shared.MatchPrepared {
		t.Fatalf("expected Prepared, got %s", m.State)
	}
	evts, err := m.Activate(3, now.Add(time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if len(evts) != 1 {
		t.Fatalf("expected 1 event, got %d", len(evts))
	}
	if m.State != shared.MatchActive {
		t.Fatalf("expected Active, got %s", m.State)
	}
	if m.StartedAt == nil {
		t.Fatal("expected StartedAt to be set")
	}
}

func TestPlaceCorrectValue(t *testing.T) {
	clues := testClues(1) // cell 0 is fixed
	m, p, now := makeMatch(t, shared.ErrorPresetCasual, clues)
	m.Activate(3, now.Add(time.Second))

	meta := cmdMeta(t, m, p)
	events, err := m.Apply(PlaceValueCommand{Meta: meta, Cell: 1, Digit: 2}, 3, now.Add(2*time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	placed := events[0].(ValuePlacedEvent)
	if placed.Cell != 1 || placed.Digit != 2 {
		t.Fatalf("unexpected placed event: %+v", placed)
	}
	if placed.Correct == nil || !*placed.Correct {
		t.Fatal("expected correct=true")
	}
	if m.Contributions[p] != 1 {
		t.Fatalf("expected 1 contribution, got %d", m.Contributions[p])
	}
}

func TestPlaceWrongValueCasual(t *testing.T) {
	m, p, now := makeMatch(t, shared.ErrorPresetCasual, testClues(0))
	m.Activate(3, now.Add(time.Second))

	meta := cmdMeta(t, m, p)
	events, err := m.Apply(PlaceValueCommand{Meta: meta, Cell: 0, Digit: 9}, 3, now.Add(2*time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	placed := events[0].(ValuePlacedEvent)
	if placed.Correct == nil || *placed.Correct {
		t.Fatal("expected correct=false")
	}
	if m.Mistakes[p] != 1 {
		t.Fatalf("expected 1 mistake, got %d", m.Mistakes[p])
	}
}

func TestPlaceWrongValueChallenge(t *testing.T) {
	m, p, now := makeMatch(t, shared.ErrorPresetChallenge, testClues(0))
	m.Activate(3, now.Add(time.Second))

	meta := cmdMeta(t, m, p)
	events, err := m.Apply(PlaceValueCommand{Meta: meta, Cell: 0, Digit: 9}, 3, now.Add(2*time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	rejected := events[0].(ValueRejectedEvent)
	if rejected.PenaltyMs != 5000 {
		t.Fatalf("expected 5000ms penalty, got %d", rejected.PenaltyMs)
	}
	if m.Mistakes[p] != 1 {
		t.Fatalf("expected 1 mistake, got %d", m.Mistakes[p])
	}
	if _, ok := m.Values[0]; ok {
		t.Fatal("expected value not to be placed")
	}
	if m.PenaltiesMs != 5000 {
		t.Fatalf("expected 5000ms total penalty, got %d", m.PenaltiesMs)
	}
}

func TestPlaceWrongValueClean(t *testing.T) {
	m, p, now := makeMatch(t, shared.ErrorPresetClean, testClues(0))
	m.Activate(3, now.Add(time.Second))

	meta := cmdMeta(t, m, p)
	events, err := m.Apply(PlaceValueCommand{Meta: meta, Cell: 0, Digit: 9}, 3, now.Add(2*time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	rejected := events[0].(ValueRejectedEvent)
	if rejected.PenaltyMs != 0 {
		t.Fatalf("expected no penalty, got %d", rejected.PenaltyMs)
	}
}

func TestPlaceWrongValueBlind(t *testing.T) {
	m, p, now := makeMatch(t, shared.ErrorPresetBlind, testClues(0))
	m.Activate(3, now.Add(time.Second))

	meta := cmdMeta(t, m, p)
	events, err := m.Apply(PlaceValueCommand{Meta: meta, Cell: 0, Digit: 9}, 3, now.Add(2*time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	placed := events[0].(ValuePlacedEvent)
	if placed.Correct != nil {
		t.Fatal("expected hidden correctness")
	}
}

func TestPlaceValueOnFixedCellRejected(t *testing.T) {
	clues := testClues(1)
	m, p, now := makeMatch(t, shared.ErrorPresetCasual, clues)
	m.Activate(3, now.Add(time.Second))

	meta := cmdMeta(t, m, p)
	_, err := m.Apply(PlaceValueCommand{Meta: meta, Cell: 0, Digit: 1}, 3, now.Add(2*time.Second))
	if err == nil {
		t.Fatal("expected error for fixed cell")
	}
	if err.(shared.DomainError).Code != shared.ErrCellFixed {
		t.Fatalf("expected CELL_FIXED, got %s", err.(shared.DomainError).Code)
	}
}

func TestEraseValue(t *testing.T) {
	m, p, now := makeMatch(t, shared.ErrorPresetCasual, testClues(0))
	m.Activate(3, now.Add(time.Second))

	meta := cmdMeta(t, m, p)
	_, err := m.Apply(PlaceValueCommand{Meta: meta, Cell: 0, Digit: 1}, 3, now.Add(2*time.Second))
	if err != nil {
		t.Fatal(err)
	}

	meta2 := cmdMeta(t, m, p)
	events, err := m.Apply(EraseValueCommand{Meta: meta2, Cell: 0}, 4, now.Add(3*time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if _, ok := m.Values[0]; ok {
		t.Fatal("expected value to be erased")
	}
}

func TestNotesAndAutoRemoval(t *testing.T) {
	m, p, now := makeMatch(t, shared.ErrorPresetCasual, testClues(0))
	m.Activate(3, now.Add(time.Second))

	meta := cmdMeta(t, m, p)
	events, err := m.Apply(AddNoteCommand{Meta: meta, Cell: 1, Digits: []shared.Digit{2, 3}}, 3, now.Add(2*time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 note event, got %d", len(events))
	}
	if !m.Notes[1].Contains(2) || !m.Notes[1].Contains(3) {
		t.Fatal("expected notes to be added")
	}

	// Place value 2 at cell 0; peer cell 1 should auto-remove 2.
	meta2 := cmdMeta(t, m, p)
	events, err = m.Apply(PlaceValueCommand{Meta: meta2, Cell: 0, Digit: 2}, 4, now.Add(3*time.Second))
	if err != nil {
		t.Fatal(err)
	}
	autoRemoved := false
	for _, e := range events {
		if ev, ok := e.(NotesAutoRemovedEvent); ok && ev.Cell == 1 {
			autoRemoved = true
		}
	}
	if !autoRemoved {
		t.Fatalf("expected auto-remove event, got %+v", events)
	}
	if m.Notes[1].Contains(2) {
		t.Fatal("expected digit 2 to be removed from peer notes")
	}
	if !m.Notes[1].Contains(3) {
		t.Fatal("expected digit 3 to remain in peer notes")
	}
}

func TestDuplicateRequestIdempotent(t *testing.T) {
	m, p, now := makeMatch(t, shared.ErrorPresetCasual, testClues(0))
	m.Activate(3, now.Add(time.Second))

	meta := cmdMeta(t, m, p)
	reqID := meta.RequestID
	seq := meta.ClientSequence
	target := meta.Target

	cmd := PlaceValueCommand{Meta: meta, Cell: 0, Digit: 1}
	_, err := m.Apply(cmd, 3, now.Add(2*time.Second))
	if err != nil {
		t.Fatal(err)
	}

	meta2, err := shared.NewCommandMetadata(reqID, p, seq, target, uint64(m.Version))
	if err != nil {
		t.Fatal(err)
	}
	cmd2 := PlaceValueCommand{Meta: meta2, Cell: 0, Digit: 2}
	events, err := m.Apply(cmd2, 99, now.Add(3*time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 0 {
		t.Fatalf("expected no events for duplicate request, got %d", len(events))
	}
}

func TestStaleVersionRejected(t *testing.T) {
	m, p, now := makeMatch(t, shared.ErrorPresetCasual, testClues(0))
	m.Activate(3, now.Add(time.Second))

	meta := cmdMeta(t, m, p)
	meta.ExpectedVersion = 0
	_, err := m.Apply(PlaceValueCommand{Meta: meta, Cell: 0, Digit: 1}, 3, now.Add(2*time.Second))
	if err == nil {
		t.Fatal("expected stale version error")
	}
	if err.(shared.DomainError).Code != shared.ErrStaleVersion {
		t.Fatalf("expected STALE_VERSION, got %s", err.(shared.DomainError).Code)
	}
}

func TestHintRevealCompletesPuzzle(t *testing.T) {
	clues := testClues(80) // only cell 80 empty
	m, p, now := makeMatch(t, shared.ErrorPresetCasual, clues)
	m.Activate(3, now.Add(time.Second))

	meta := cmdMeta(t, m, p)
	events, err := m.Apply(UseHintCommand{Meta: meta, Level: shared.HintReveal}, 3, now.Add(2*time.Second))
	if err != nil {
		t.Fatal(err)
	}
	completed := false
	for _, e := range events {
		if _, ok := e.(MatchCompletedEvent); ok {
			completed = true
		}
	}
	if !completed {
		t.Fatalf("expected MatchCompleted event, got %+v", events)
	}
	if m.State != shared.MatchCompleted {
		t.Fatalf("expected Completed, got %s", m.State)
	}
	if m.Result == nil || !m.Result.Assisted {
		t.Fatal("expected assisted result")
	}
}

func TestNudgeHint(t *testing.T) {
	m, p, now := makeMatch(t, shared.ErrorPresetCasual, testClues(0))
	m.Activate(3, now.Add(time.Second))

	meta := cmdMeta(t, m, p)
	events, err := m.Apply(UseHintCommand{Meta: meta, Level: shared.HintNudge}, 3, now.Add(2*time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	hint := events[0].(HintUsedEvent)
	if hint.Level != shared.HintNudge || hint.TargetCell == nil {
		t.Fatalf("unexpected hint event: %+v", hint)
	}
}

func TestFullCompletion(t *testing.T) {
	m, p, now := makeMatch(t, shared.ErrorPresetCasual, testClues(0))
	m.Activate(3, now.Add(time.Second))

	next := uint64(3)
	for i := 0; i < 81; i++ {
		meta := cmdMeta(t, m, p)
		events, err := m.Apply(PlaceValueCommand{Meta: meta, Cell: shared.CellIndex(i), Digit: shared.Digit(testSolution[i])}, next, now.Add(time.Duration(i+2)*time.Second))
		if err != nil {
			t.Fatalf("placing cell %d: %v", i, err)
		}
		next += uint64(len(events))
	}
	if m.State != shared.MatchCompleted {
		t.Fatalf("expected Completed, got %s", m.State)
	}
	if m.Result == nil {
		t.Fatal("expected result")
	}
	if m.Result.ContributionCount != 81 {
		t.Fatalf("expected 81 contributions, got %d", m.Result.ContributionCount)
	}
}

func TestPingIsDurableButNotVersionBump(t *testing.T) {
	m, p, now := makeMatch(t, shared.ErrorPresetCasual, testClues(0))
	m.Activate(3, now.Add(time.Second))
	v := m.Version

	meta := cmdMeta(t, m, p)
	events, err := m.Apply(PingCommand{Meta: meta, Cell: 0, Intent: "look_here"}, 3, now.Add(2*time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 ping event, got %d", len(events))
	}
	if m.Version != v {
		t.Fatalf("ping must not bump aggregate version: expected %d, got %d", v, m.Version)
	}
}

func BenchmarkPlaceValue(b *testing.B) {
	m, p, now := makeMatch(b, shared.ErrorPresetCasual, testClues(0))
	m.Activate(3, now.Add(time.Second))
	next := uint64(3)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cell := shared.CellIndex(i % 5)
		meta := cmdMeta(b, m, p)
		events, err := m.Apply(PlaceValueCommand{Meta: meta, Cell: cell, Digit: shared.Digit(testSolution[cell])}, next, now.Add(time.Duration(i+2)*time.Second))
		if err != nil {
			b.Fatal(err)
		}
		next += uint64(len(events))
	}
}

func TestReconstructFromEvents(t *testing.T) {
	m, p, now := makeMatch(t, shared.ErrorPresetCasual, testClues(0))
	startEvents, _ := m.Activate(3, now.Add(time.Second))

	next := uint64(3)
	var allEvents []Event
	allEvents = append(allEvents, startEvents...)

	meta := cmdMeta(t, m, p)
	events, err := m.Apply(PlaceValueCommand{Meta: meta, Cell: 0, Digit: 1}, next, now.Add(2*time.Second))
	if err != nil {
		t.Fatal(err)
	}
	allEvents = append(allEvents, events...)
	next += uint64(len(events))

	meta2 := cmdMeta(t, m, p)
	events, err = m.Apply(AddNoteCommand{Meta: meta2, Cell: 1, Digits: []shared.Digit{2, 3}}, next, now.Add(3*time.Second))
	if err != nil {
		t.Fatal(err)
	}
	allEvents = append(allEvents, events...)

	reconstructed, err := ReconstructMatch(m.Puzzle, m.Rules, m.Participants, allEvents)
	if err != nil {
		t.Fatal(err)
	}
	if reconstructed.State != shared.MatchActive {
		t.Fatalf("expected Active after reconstruction, got %s", reconstructed.State)
	}
	if reconstructed.Values[0] != 1 {
		t.Fatalf("expected reconstructed value 1 at cell 0, got %v", reconstructed.Values[0])
	}
	if !reconstructed.Notes[1].Contains(2) || !reconstructed.Notes[1].Contains(3) {
		t.Fatalf("expected reconstructed notes at cell 1, got %v", reconstructed.Notes[1])
	}
	if reconstructed.Contributions[p] != 1 {
		t.Fatalf("expected 1 contribution after reconstruction, got %d", reconstructed.Contributions[p])
	}
}
