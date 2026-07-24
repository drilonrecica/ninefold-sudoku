package actor

import (
	"database/sql"

	shared "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/domain"
	matchdomain "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/match/domain"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/gen"
	roomdomain "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/room/domain"
)

func roomToGen(r *roomdomain.Room) gen.Room {
	hostID := ""
	if r.HostParticipantID != nil {
		hostID = r.HostParticipantID.String()
	}
	return gen.Room{
		ID:                r.ID.String(),
		Code:              r.Code.String(),
		State:             string(r.State),
		Version:           int64(r.Version),
		Mode:              string(r.Rules.Mode),
		Difficulty:        string(r.Rules.Difficulty),
		ErrorPreset:       string(r.Rules.ErrorPreset),
		HintsEnabled:      boolToInt(r.Rules.HintsEnabled),
		SharedNotes:       boolToInt(r.Rules.SharedNotes),
		AutoRemoveNotes:   boolToInt(r.Rules.AutoRemoveNotes),
		SpectatorsAllowed: boolToInt(r.Rules.SpectatorsAllowed),
		HostParticipantID: hostID,
		CurrentMatchID:    nullMatchIDString(r.CurrentMatchID),
		CreatedAtMs:       r.CreatedAt.Milliseconds(),
		LastActivityAtMs:  r.LastActivityAt.Milliseconds(),
		ExpiresAtMs:       r.ExpiresAt.Milliseconds(),
	}
}

func roomFromGen(gr gen.Room, participants []gen.RoomParticipant) (*roomdomain.Room, error) {
	mode, err := shared.ParseMode(gr.Mode)
	if err != nil {
		return nil, err
	}
	difficulty, err := shared.ParseDifficulty(gr.Difficulty)
	if err != nil {
		return nil, err
	}
	errorPreset, err := shared.ParseErrorPreset(gr.ErrorPreset)
	if err != nil {
		return nil, err
	}
	createdAt, err := shared.TimestampFromMilliseconds(gr.CreatedAtMs)
	if err != nil {
		return nil, err
	}
	lastActivity, err := shared.TimestampFromMilliseconds(gr.LastActivityAtMs)
	if err != nil {
		return nil, err
	}
	expiresAt, err := shared.TimestampFromMilliseconds(gr.ExpiresAtMs)
	if err != nil {
		return nil, err
	}
	room := &roomdomain.Room{
		ID:    shared.RoomID(gr.ID),
		Code:  shared.RoomCode(gr.Code),
		State: shared.RoomState(gr.State),
		Rules: roomdomain.MatchRules{
			Mode:              mode,
			Difficulty:        difficulty,
			ErrorPreset:       errorPreset,
			HintsEnabled:      gr.HintsEnabled != 0,
			SharedNotes:       gr.SharedNotes != 0,
			AutoRemoveNotes:   gr.AutoRemoveNotes != 0,
			SpectatorsAllowed: gr.SpectatorsAllowed != 0,
		},
		CreatedAt:      createdAt,
		LastActivityAt: lastActivity,
		ExpiresAt:      expiresAt,
	}
	room.Version = shared.RoomVersion(gr.Version)
	if gr.HostParticipantID != "" {
		id := shared.ParticipantID(gr.HostParticipantID)
		room.HostParticipantID = &id
	}
	if gr.CurrentMatchID.Valid {
		id := shared.MatchID(gr.CurrentMatchID.String)
		room.CurrentMatchID = &id
	}
	for _, gp := range participants {
		p, err := participantFromGen(gp)
		if err != nil {
			return nil, err
		}
		room.Participants = append(room.Participants, p)
	}
	return room, nil
}

func participantToGen(p roomdomain.Participant, roomID shared.RoomID) gen.RoomParticipant {
	return gen.RoomParticipant{
		ID:            p.ID.String(),
		RoomID:        roomID.String(),
		DisplayName:   p.Name.String(),
		Role:          string(p.Role),
		IsHost:        boolToInt(p.IsHost),
		IsReady:       boolToInt(p.IsReady),
		JoinedAtMs:    p.JoinedAt.Milliseconds(),
		LeftAtMs:      nullInt64(p.LeftAt),
		RemovedAtMs:   nullInt64(p.RemovedAt),
		RemovedReason: nullString(p.RemovedReason),
	}
}

func participantFromGen(gp gen.RoomParticipant) (roomdomain.Participant, error) {
	name, err := shared.NewDisplayName(gp.DisplayName)
	if err != nil {
		return roomdomain.Participant{}, err
	}
	role, err := shared.ParseParticipationRole(gp.Role)
	if err != nil {
		return roomdomain.Participant{}, err
	}
	joinedAt, err := shared.TimestampFromMilliseconds(gp.JoinedAtMs)
	if err != nil {
		return roomdomain.Participant{}, err
	}
	p := roomdomain.Participant{
		ID:       shared.ParticipantID(gp.ID),
		Name:     name,
		Role:     role,
		IsHost:   gp.IsHost != 0,
		IsReady:  gp.IsReady != 0,
		JoinedAt: joinedAt,
	}
	if gp.RemovedReason.Valid {
		p.RemovedReason = gp.RemovedReason.String
	}
	if gp.LeftAtMs.Valid {
		ts, err := shared.TimestampFromMilliseconds(gp.LeftAtMs.Int64)
		if err != nil {
			return roomdomain.Participant{}, err
		}
		p.LeftAt = &ts
	}
	if gp.RemovedAtMs.Valid {
		ts, err := shared.TimestampFromMilliseconds(gp.RemovedAtMs.Int64)
		if err != nil {
			return roomdomain.Participant{}, err
		}
		p.RemovedAt = &ts
	}
	return p, nil
}

func matchRulesToGen(rules matchdomain.Rules) (mode, difficulty, errorPreset string, hintsEnabled, autoRemoveNotes int64, ruleVersion int64) {
	return string(rules.Mode), string(rules.Difficulty), string(rules.ErrorPreset),
		boolToInt(rules.HintsEnabled), boolToInt(rules.AutoRemoveNotes), int64(rules.RuleVersion)
}

func matchToGen(m *matchdomain.Match) gen.Match {
	mode, difficulty, errorPreset, hintsEnabled, autoRemoveNotes, ruleVersion := matchRulesToGen(m.Rules)
	match := gen.Match{
		ID:                 m.ID.String(),
		RoomID:             m.RoomID.String(),
		State:              string(m.State),
		Version:            int64(m.Version),
		Mode:               mode,
		Difficulty:         difficulty,
		ErrorPreset:        errorPreset,
		HintsEnabled:       hintsEnabled,
		AutoRemoveNotes:    autoRemoveNotes,
		RuleVersion:        ruleVersion,
		PuzzleID:           m.Puzzle.PuzzleID.String(),
		PuzzleRevision:     int64(m.Puzzle.Revision),
		TransformationSeed: int64(m.Puzzle.TransformationSeed),
		PuzzleDifficulty:   string(m.Puzzle.Difficulty),
		GeneratorVersion:   m.Puzzle.GeneratorVersion,
		SolverVersion:      m.Puzzle.SolverVersion,
		CreatedAtMs:        m.CreatedAt.Milliseconds(),
	}
	if m.StartedAt != nil {
		match.StartedAtMs = sql.NullInt64{Int64: m.StartedAt.Milliseconds(), Valid: true}
	}
	if m.CompletedAt != nil {
		match.CompletedAtMs = sql.NullInt64{Int64: m.CompletedAt.Milliseconds(), Valid: true}
	}
	if m.Result != nil {
		match.ResultReason = sql.NullString{String: m.Result.Reason, Valid: true}
		match.ElapsedMs = sql.NullInt64{Int64: int64(m.Result.ElapsedMilliseconds), Valid: true}
		match.Assisted = boolToInt(m.Result.Assisted)
	}
	return match
}

func matchFromGen(gm gen.Match, participantIDs []shared.ParticipantID) (*matchdomain.Match, error) {
	mode, err := shared.ParseMode(gm.Mode)
	if err != nil {
		return nil, err
	}
	difficulty, err := shared.ParseDifficulty(gm.Difficulty)
	if err != nil {
		return nil, err
	}
	errorPreset, err := shared.ParseErrorPreset(gm.ErrorPreset)
	if err != nil {
		return nil, err
	}
	m := &matchdomain.Match{
		ID:      shared.MatchID(gm.ID),
		RoomID:  shared.RoomID(gm.RoomID),
		Version: shared.MatchVersion(gm.Version),
		State:   shared.MatchState(gm.State),
		Rules: matchdomain.Rules{
			Mode:            mode,
			Difficulty:      difficulty,
			ErrorPreset:     errorPreset,
			HintsEnabled:    gm.HintsEnabled != 0,
			AutoRemoveNotes: gm.AutoRemoveNotes != 0,
			RuleVersion:     uint16(gm.RuleVersion),
		},
		Puzzle: shared.AssignedPuzzle{
			PuzzleID:           shared.PuzzleID(gm.PuzzleID),
			Revision:           uint32(gm.PuzzleRevision),
			TransformationSeed: uint64(gm.TransformationSeed),
			Difficulty:         difficulty,
			GeneratorVersion:   gm.GeneratorVersion,
			SolverVersion:      gm.SolverVersion,
		},
		Participants: participantIDs,
	}
	return m, nil
}

func boolToInt(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

func nullMatchIDString(s *shared.MatchID) sql.NullString {
	if s == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: s.String(), Valid: true}
}

func nullInt64(ts *shared.Timestamp) sql.NullInt64 {
	if ts == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: ts.Milliseconds(), Valid: true}
}

func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}
