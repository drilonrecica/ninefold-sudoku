package domain

import (
	"errors"
	"testing"
	"time"

	shared "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/domain"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/platform/idgen"
)

func newRoom(t *testing.T, hostName string) *Room {
	t.Helper()
	name, err := shared.NewDisplayName(hostName)
	if err != nil {
		t.Fatalf("name: %v", err)
	}
	g := idgen.Generator{}
	hostID, _ := g.ParticipantID()
	roomID, _ := g.RoomID()
	host := Participant{
		ID:       hostID,
		Name:     name,
		Role:     shared.RolePlayer,
		JoinedAt: nowTs(t),
	}
	rules := MatchRules{
		Mode:              shared.ModeCoop,
		Difficulty:        shared.DifficultyMedium,
		ErrorPreset:       shared.ErrorPresetCasual,
		HintsEnabled:      true,
		SharedNotes:       true,
		AutoRemoveNotes:   true,
		SpectatorsAllowed: true,
		RuleVersion:       1,
	}
	room, err := NewRoom(roomID, shared.RoomCode("7KMP4R"), host, rules, time.Now())
	if err != nil {
		t.Fatalf("new room: %v", err)
	}
	return room
}

func nowTs(t *testing.T) shared.Timestamp {
	ts, err := shared.NewTimestamp(time.Now())
	if err != nil {
		t.Fatalf("timestamp: %v", err)
	}
	return ts
}

func meta(t *testing.T, participant shared.ParticipantID, roomID shared.RoomID, version uint64) shared.CommandMetadata {
	t.Helper()
	g := idgen.Generator{}
	reqID, _ := g.RequestID()
	m, err := shared.NewCommandMetadata(
		reqID,
		participant,
		1,
		shared.NewRoomTarget(roomID),
		version,
	)
	if err != nil {
		t.Fatalf("meta: %v", err)
	}
	return m
}

var testGen idgen.Generator

func newParticipantID(t *testing.T) shared.ParticipantID {
	id, err := testGen.ParticipantID()
	if err != nil {
		t.Fatalf("participant id: %v", err)
	}
	return id
}

func newMatchID(t *testing.T) shared.MatchID {
	id, err := testGen.MatchID()
	if err != nil {
		t.Fatalf("match id: %v", err)
	}
	return id
}

func newPuzzleID(t *testing.T) shared.PuzzleID {
	id, err := testGen.PuzzleID()
	if err != nil {
		t.Fatalf("puzzle id: %v", err)
	}
	return id
}

func TestNewRoomStartsInLobbyWithHost(t *testing.T) {
	room := newRoom(t, "Mila")
	if room.State != shared.RoomLobby {
		t.Fatalf("expected Lobby, got %s", room.State)
	}
	if room.HostParticipantID == nil || *room.HostParticipantID != room.Participants[0].ID {
		t.Fatalf("expected host to be first participant")
	}
}

func TestSettingsChangeResetsReadiness(t *testing.T) {
	room := newRoom(t, "Mila")
	mila := room.Participants[0].ID
	room.Participants[0].IsReady = true

	difficulty := shared.DifficultyHard
	_, err := room.Apply(ChangeRoomSettingsCommand{
		Meta: meta(t, mila, room.ID, 1),
		Settings: RoomSettingsPatch{
			Difficulty: &difficulty,
		},
	}, time.Now())
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if room.Participants[0].IsReady {
		t.Fatalf("expected readiness reset after settings change")
	}
	if room.Rules.Difficulty != shared.DifficultyHard {
		t.Fatalf("expected difficulty Hard, got %s", room.Rules.Difficulty)
	}
}

func TestNonHostCannotChangeSettings(t *testing.T) {
	room := newRoom(t, "Mila")
	other, _ := shared.NewDisplayName("Noah")
	room.Participants = append(room.Participants, Participant{
		ID:       newParticipantID(t),
		Name:     other,
		Role:     shared.RolePlayer,
		JoinedAt: nowTs(t),
	})
	difficulty := shared.DifficultyHard
	_, err := room.Apply(ChangeRoomSettingsCommand{
		Meta: meta(t, room.Participants[1].ID, room.ID, 1),
		Settings: RoomSettingsPatch{
			Difficulty: &difficulty,
		},
	}, time.Now())
	if !isCode(err, shared.ErrNotRoomHost) {
		t.Fatalf("expected NOT_ROOM_HOST, got %v", err)
	}
}

func TestStartCountdownRequiresAllPlayersReady(t *testing.T) {
	room := newRoom(t, "Mila")
	other, _ := shared.NewDisplayName("Noah")
	room.Participants = append(room.Participants, Participant{
		ID:       newParticipantID(t),
		Name:     other,
		Role:     shared.RolePlayer,
		JoinedAt: nowTs(t),
		IsReady:  true,
	})
	mila := room.Participants[0].ID

	_, err := room.Apply(StartCountdownCommand{
		Meta:    meta(t, mila, room.ID, 1),
		MatchID: newMatchID(t),
		Puzzle:  shared.AssignedPuzzle{PuzzleID: newPuzzleID(t), Revision: 1},
	}, time.Now())
	if !isCode(err, shared.ErrPlayersNotReady) {
		t.Fatalf("expected PLAYERS_NOT_READY, got %v", err)
	}
}

func TestStartCountdownAndActivate(t *testing.T) {
	room := newRoom(t, "Mila")
	room.Participants[0].IsReady = true
	mila := room.Participants[0].ID

	matchID := newMatchID(t)
	_, err := room.Apply(StartCountdownCommand{
		Meta:    meta(t, mila, room.ID, 1),
		MatchID: matchID,
		Puzzle:  shared.AssignedPuzzle{PuzzleID: newPuzzleID(t), Revision: 1},
	}, time.Now())
	if err != nil {
		t.Fatalf("start countdown: %v", err)
	}
	if room.State != shared.RoomCountdown {
		t.Fatalf("expected Countdown, got %s", room.State)
	}
	if room.Countdown == nil || room.Countdown.MatchID != matchID {
		t.Fatalf("expected countdown match reference")
	}

	_, err = room.Apply(ActivateMatchCommand{
		Meta:       meta(t, mila, room.ID, 2),
		Generation: room.Countdown.Generation,
	}, time.Now())
	if err != nil {
		t.Fatalf("activate: %v", err)
	}
	if room.State != shared.RoomInMatch {
		t.Fatalf("expected InMatch, got %s", room.State)
	}
	if room.CurrentMatchID == nil || *room.CurrentMatchID != matchID {
		t.Fatalf("expected current match id")
	}
	if room.Countdown != nil {
		t.Fatalf("expected countdown cleared")
	}
}

func TestHostTransfer(t *testing.T) {
	room := newRoom(t, "Mila")
	other, _ := shared.NewDisplayName("Noah")
	room.Participants = append(room.Participants, Participant{
		ID:       newParticipantID(t),
		Name:     other,
		Role:     shared.RolePlayer,
		JoinedAt: nowTs(t),
	})
	mila := room.Participants[0].ID
	noah := room.Participants[1].ID

	_, err := room.Apply(TransferHostCommand{Meta: meta(t, mila, room.ID, 1), Participant: noah}, time.Now())
	if err != nil {
		t.Fatalf("transfer: %v", err)
	}
	if *room.HostParticipantID != noah {
		t.Fatalf("expected Noah to be host")
	}
	if room.Participants[0].IsHost {
		t.Fatalf("expected Mila to lose host")
	}
}

func TestDuplicateNameRejected(t *testing.T) {
	room := newRoom(t, "Mila")
	name, _ := shared.NewDisplayName("Mila")
	_, err := room.Apply(RequestJoinCommand{
		Meta:          meta(t, newParticipantID(t), room.ID, 1),
		Code:          room.Code,
		DisplayName:   name,
		Role:          shared.RolePlayer,
		ParticipantID: newParticipantID(t),
	}, time.Now())
	if !isCode(err, shared.ErrNameAlreadyUsed) {
		t.Fatalf("expected NAME_ALREADY_USED, got %v", err)
	}
}

func TestLeaveTransfersHostToLongestPresent(t *testing.T) {
	room := newRoom(t, "Mila")
	other, _ := shared.NewDisplayName("Noah")
	room.Participants = append(room.Participants, Participant{
		ID:       newParticipantID(t),
		Name:     other,
		Role:     shared.RolePlayer,
		JoinedAt: nowTs(t),
	})
	mila := room.Participants[0].ID

	_, err := room.Apply(LeaveRoomCommand{Meta: meta(t, mila, room.ID, 1), Intent: "leave_lobby"}, time.Now())
	if err != nil {
		t.Fatalf("leave: %v", err)
	}
	if *room.HostParticipantID != room.Participants[1].ID {
		t.Fatalf("expected host transfer to Noah")
	}
}

func isCode(err error, code shared.ErrorCode) bool {
	var d shared.DomainError
	if !errors.As(err, &d) {
		return false
	}
	return d.Code == code
}
