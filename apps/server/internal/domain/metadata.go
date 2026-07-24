package domain

import "errors"

type AggregateKind string

const (
	AggregateRoom  AggregateKind = "Room"
	AggregateMatch AggregateKind = "Match"
)

type AggregateTarget struct {
	Kind AggregateKind
	ID   string
}

func NewRoomTarget(id RoomID) AggregateTarget {
	return AggregateTarget{Kind: AggregateRoom, ID: id.String()}
}

func NewMatchTarget(id MatchID) AggregateTarget {
	return AggregateTarget{Kind: AggregateMatch, ID: id.String()}
}

type CommandMetadata struct {
	RequestID                  RequestID
	AuthenticatedParticipantID ParticipantID
	ClientSequence             ClientSequence
	Target                     AggregateTarget
	ExpectedVersion            uint64
}

func NewCommandMetadata(requestID RequestID, participantID ParticipantID, sequence ClientSequence, target AggregateTarget, expectedVersion uint64) (CommandMetadata, error) {
	if _, err := ParseRequestID(requestID.String()); err != nil {
		return CommandMetadata{}, errors.New("request ID is invalid")
	}
	if _, err := ParseParticipantID(participantID.String()); err != nil {
		return CommandMetadata{}, errors.New("authenticated participant ID is invalid")
	}
	if sequence == 0 || uint64(sequence) > MaxSafeInteger {
		return CommandMetadata{}, errors.New("client sequence is outside the JSON safe integer range")
	}
	if expectedVersion > MaxSafeInteger {
		return CommandMetadata{}, errors.New("expected version exceeds the JSON safe integer maximum")
	}
	if err := validateTarget(target); err != nil {
		return CommandMetadata{}, errors.New("aggregate target is invalid")
	}
	return CommandMetadata{
		RequestID:                  requestID,
		AuthenticatedParticipantID: participantID,
		ClientSequence:             sequence,
		Target:                     target,
		ExpectedVersion:            expectedVersion,
	}, nil
}

type EventMetadata struct {
	SchemaVersion    uint16
	EventNumber      EventNumber
	AggregateVersion uint64
	Target           AggregateTarget
	OccurredAt       Timestamp
}

func NewEventMetadata(schemaVersion uint16, eventNumber EventNumber, aggregateVersion uint64, target AggregateTarget, occurredAt Timestamp) (EventMetadata, error) {
	if schemaVersion == 0 {
		return EventMetadata{}, errors.New("schema version must be positive")
	}
	if eventNumber == 0 || uint64(eventNumber) > MaxSafeInteger {
		return EventMetadata{}, errors.New("event number is outside the JSON safe integer range")
	}
	if aggregateVersion == 0 || aggregateVersion > MaxSafeInteger {
		return EventMetadata{}, errors.New("aggregate version must be within the JSON safe integer range")
	}
	if err := validateTarget(target); err != nil {
		return EventMetadata{}, errors.New("aggregate target is invalid")
	}
	if occurredAt.Milliseconds() <= 0 {
		return EventMetadata{}, errors.New("server occurrence time is invalid")
	}
	return EventMetadata{
		SchemaVersion:    schemaVersion,
		EventNumber:      eventNumber,
		AggregateVersion: aggregateVersion,
		Target:           target,
		OccurredAt:       occurredAt,
	}, nil
}

func validateTarget(target AggregateTarget) error {
	switch target.Kind {
	case AggregateRoom:
		_, err := ParseRoomID(target.ID)
		return err
	case AggregateMatch:
		_, err := ParseMatchID(target.ID)
		return err
	default:
		return errors.New("aggregate kind is invalid")
	}
}
