package actor

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"io"
	"time"

	shared "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/domain"
	matchdomain "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/match/domain"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/gen"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/repository"
)

const snapshotStateFormat = "match-state-v1+gzip"

type matchSnapshotState struct {
	MatchID               shared.MatchID
	RoomID                shared.RoomID
	Version               shared.MatchVersion
	State                 shared.MatchState
	Cells                 [81]matchdomain.Cell
	Values                map[shared.CellIndex]shared.Digit
	Attribution           map[shared.CellIndex]shared.ParticipantID
	Notes                 [81]shared.CandidateSet
	Mistakes              map[shared.ParticipantID]uint32
	Contributions         map[shared.ParticipantID]uint32
	Disconnects           map[shared.ParticipantID]uint32
	Connected             map[shared.ParticipantID]bool
	HintsByPlayer         map[shared.ParticipantID]uint32
	HintsUsed             uint32
	PenaltiesMs           uint64
	Assisted              bool
	StartedAtMs           *int64
	CompletedAtMs         *int64
	PausedMilliseconds    uint64
	RecoveryStartedAtMs   *int64
	RecoveryGeneration    uint64
	RecoveryPreviousState shared.MatchState
	CreatedAtMs           int64
	Result                *matchSnapshotResult
}

type matchSnapshotResult struct {
	Reason                string
	CompletedAtMs         int64
	ElapsedMilliseconds   uint64
	PenaltyMilliseconds   uint64
	Assisted              bool
	MistakesByPlayer      map[shared.ParticipantID]uint32
	ContributionsByPlayer map[shared.ParticipantID]uint32
	DisconnectsByPlayer   map[shared.ParticipantID]uint32
	HintCount             uint32
	ContributionCount     uint32
}

func snapshotStateFromMatch(match *matchdomain.Match) matchSnapshotState {
	state := matchSnapshotState{
		MatchID:               match.ID,
		RoomID:                match.RoomID,
		Version:               match.Version,
		State:                 match.State,
		Cells:                 match.Cells,
		Values:                match.Values,
		Attribution:           match.Attribution,
		Notes:                 match.Notes,
		Mistakes:              match.Mistakes,
		Contributions:         match.Contributions,
		Disconnects:           match.Disconnects,
		Connected:             match.Connected,
		HintsByPlayer:         match.HintsByPlayer,
		HintsUsed:             match.HintsUsed,
		PenaltiesMs:           match.PenaltiesMs,
		Assisted:              match.Assisted,
		PausedMilliseconds:    match.PausedMilliseconds,
		RecoveryGeneration:    match.RecoveryGeneration,
		RecoveryPreviousState: match.RecoveryPreviousState,
		CreatedAtMs:           match.CreatedAt.Milliseconds(),
	}
	state.StartedAtMs = timestampMilliseconds(match.StartedAt)
	state.CompletedAtMs = timestampMilliseconds(match.CompletedAt)
	state.RecoveryStartedAtMs = timestampMilliseconds(match.RecoveryStartedAt)
	if match.Result != nil {
		state.Result = &matchSnapshotResult{
			Reason:                match.Result.Reason,
			CompletedAtMs:         match.Result.CompletedAt.Milliseconds(),
			ElapsedMilliseconds:   match.Result.ElapsedMilliseconds,
			PenaltyMilliseconds:   match.Result.PenaltyMilliseconds,
			Assisted:              match.Result.Assisted,
			MistakesByPlayer:      match.Result.MistakesByPlayer,
			ContributionsByPlayer: match.Result.ContributionsByPlayer,
			DisconnectsByPlayer:   match.Result.DisconnectsByPlayer,
			HintCount:             match.Result.HintCount,
			ContributionCount:     match.Result.ContributionCount,
		}
	}
	return state
}

func (state matchSnapshotState) restore(puzzle shared.AssignedPuzzle, rules matchdomain.Rules, participants []shared.ParticipantID) (*matchdomain.Match, error) {
	match, err := matchdomain.ReconstructMatch(puzzle, rules, participants, nil)
	if err != nil {
		return nil, err
	}
	match.ID = state.MatchID
	match.RoomID = state.RoomID
	match.Version = state.Version
	match.State = state.State
	match.Cells = state.Cells
	match.Values = state.Values
	match.Attribution = state.Attribution
	match.Notes = state.Notes
	match.Mistakes = state.Mistakes
	match.Contributions = state.Contributions
	match.Disconnects = state.Disconnects
	match.Connected = state.Connected
	match.HintsByPlayer = state.HintsByPlayer
	match.HintsUsed = state.HintsUsed
	match.PenaltiesMs = state.PenaltiesMs
	match.Assisted = state.Assisted
	match.PausedMilliseconds = state.PausedMilliseconds
	match.RecoveryGeneration = state.RecoveryGeneration
	match.RecoveryPreviousState = state.RecoveryPreviousState
	if match.StartedAt, err = timestampFromMilliseconds(state.StartedAtMs); err != nil {
		return nil, err
	}
	if match.CompletedAt, err = timestampFromMilliseconds(state.CompletedAtMs); err != nil {
		return nil, err
	}
	if match.RecoveryStartedAt, err = timestampFromMilliseconds(state.RecoveryStartedAtMs); err != nil {
		return nil, err
	}
	if state.CreatedAtMs > 0 {
		if match.CreatedAt, err = shared.TimestampFromMilliseconds(state.CreatedAtMs); err != nil {
			return nil, err
		}
	}
	if state.Result != nil {
		completedAt, timestampErr := shared.TimestampFromMilliseconds(state.Result.CompletedAtMs)
		if timestampErr != nil {
			return nil, timestampErr
		}
		match.Result = &matchdomain.Result{
			Reason:                state.Result.Reason,
			CompletedAt:           completedAt,
			ElapsedMilliseconds:   state.Result.ElapsedMilliseconds,
			PenaltyMilliseconds:   state.Result.PenaltyMilliseconds,
			Assisted:              state.Result.Assisted,
			MistakesByPlayer:      state.Result.MistakesByPlayer,
			ContributionsByPlayer: state.Result.ContributionsByPlayer,
			DisconnectsByPlayer:   state.Result.DisconnectsByPlayer,
			HintCount:             state.Result.HintCount,
			ContributionCount:     state.Result.ContributionCount,
		}
	}
	return match, nil
}

func encodeSnapshot(match *matchdomain.Match) ([]byte, []byte, error) {
	var compressed bytes.Buffer
	writer := gzip.NewWriter(&compressed)
	if err := json.NewEncoder(writer).Encode(snapshotStateFromMatch(match)); err != nil {
		return nil, nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, nil, err
	}
	blob := compressed.Bytes()
	digest := sha256.Sum256(blob)
	return blob, digest[:], nil
}

func decodeSnapshot(snapshot gen.MatchSnapshot) (matchSnapshotState, error) {
	if snapshot.StateFormat != snapshotStateFormat {
		return matchSnapshotState{}, errors.New("unsupported snapshot format")
	}
	digest := sha256.Sum256(snapshot.StateBlob)
	if !bytes.Equal(digest[:], snapshot.IntegrityHash) {
		return matchSnapshotState{}, errors.New("snapshot integrity check failed")
	}
	reader, err := gzip.NewReader(bytes.NewReader(snapshot.StateBlob))
	if err != nil {
		return matchSnapshotState{}, err
	}
	defer reader.Close()
	limited := io.LimitReader(reader, 16<<20)
	var state matchSnapshotState
	decoder := json.NewDecoder(limited)
	if err := decoder.Decode(&state); err != nil {
		return matchSnapshotState{}, err
	}
	if state.MatchID.String() != snapshot.MatchID || uint64(state.Version) != uint64(snapshot.AggregateVersion) {
		return matchSnapshotState{}, errors.New("snapshot aggregate identity is invalid")
	}
	return state, nil
}

func (a *Actor) shouldSnapshot(eventNumber uint64, state string, now time.Time) bool {
	if eventNumber == 0 || eventNumber == a.lastSnapshotEvent {
		return false
	}
	if a.lastSnapshotEvent == 0 || eventNumber-a.lastSnapshotEvent >= 50 {
		return true
	}
	if state == "Completed" || state == "Cancelled" {
		return true
	}
	return !a.lastSnapshotAt.IsZero() && now.Sub(a.lastSnapshotAt) >= 30*time.Second
}

func (a *Actor) createSnapshot(ctx context.Context, repo *repository.Repository, match *matchdomain.Match, eventNumber uint64, now time.Time) error {
	blob, digest, err := encodeSnapshot(match)
	if err != nil {
		return err
	}
	if err := repo.CreateMatchSnapshot(ctx, gen.MatchSnapshot{
		MatchID:          match.ID.String(),
		EventNumber:      int64(eventNumber),
		AggregateVersion: int64(match.Version),
		StateFormat:      snapshotStateFormat,
		StateBlob:        blob,
		IntegrityHash:    digest,
		CreatedAtMs:      now.UnixMilli(),
	}); err != nil {
		return err
	}
	a.lastSnapshotEvent = eventNumber
	a.lastSnapshotAt = now
	return nil
}

func timestampMilliseconds(timestamp *shared.Timestamp) *int64 {
	if timestamp == nil {
		return nil
	}
	value := timestamp.Milliseconds()
	return &value
}

func timestampFromMilliseconds(milliseconds *int64) (*shared.Timestamp, error) {
	if milliseconds == nil {
		return nil, nil
	}
	timestamp, err := shared.TimestampFromMilliseconds(*milliseconds)
	if err != nil {
		return nil, err
	}
	return &timestamp, nil
}
