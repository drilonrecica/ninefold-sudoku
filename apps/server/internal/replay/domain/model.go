package domain

import shared "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/domain"

type Replay struct {
	ID        shared.ReplayID
	MatchID   shared.MatchID
	CreatedAt shared.Timestamp
	ExpiresAt shared.Timestamp
}

type Event struct {
	Metadata shared.EventMetadata
	Type     string
	Payload  []byte
}
