package idgen

import (
	"fmt"

	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/domain"
	"github.com/google/uuid"
)

type Generator struct{}

func (Generator) RoomID() (domain.RoomID, error) {
	value, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("generate RoomID: %w", err)
	}
	return domain.ParseRoomID(value.String())
}

func (Generator) MatchID() (domain.MatchID, error) {
	value, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("generate MatchID: %w", err)
	}
	return domain.ParseMatchID(value.String())
}

func (Generator) ParticipantID() (domain.ParticipantID, error) {
	value, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("generate ParticipantID: %w", err)
	}
	return domain.ParseParticipantID(value.String())
}

func (Generator) PuzzleID() (domain.PuzzleID, error) {
	value, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("generate PuzzleID: %w", err)
	}
	return domain.ParsePuzzleID(value.String())
}

func (Generator) ReplayID() (domain.ReplayID, error) {
	value, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("generate ReplayID: %w", err)
	}
	return domain.ParseReplayID(value.String())
}

func (Generator) RequestID() (domain.RequestID, error) {
	value, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("generate RequestID: %w", err)
	}
	return domain.ParseRequestID(value.String())
}

func (Generator) ConnectionID() (domain.ConnectionID, error) {
	value, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("generate ConnectionID: %w", err)
	}
	return domain.ParseConnectionID(value.String())
}
