package actor

import (
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/gowebpki/jcs"

	shared "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/domain"
	matchdomain "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/match/domain"
	"github.com/drilonrecica/ninefold-sudoku/apps/server/internal/persistence/gen"
)

const proofVersion = 1

// genesisHash is the previous-hash value for the first event in a Match.
var genesisHash = make([]byte, sha256.Size)

// matchEventPublicEnvelope is the canonical public envelope hashed for replay integrity.
type matchEventPublicEnvelope struct {
	ProofVersion         int             `json:"proofVersion"`
	MatchID              string          `json:"matchId"`
	EventNumber          uint64          `json:"eventNumber"`
	AggregateVersion     uint64          `json:"aggregateVersion"`
	PublicEventType      string          `json:"publicEventType"`
	PublicActorID        string          `json:"publicActorId"`
	OccurredAtMs         int64           `json:"occurredAtMs"`
	PublicPayload        json.RawMessage `json:"publicPayload"`
	PrivatePayloadDigest string          `json:"privatePayloadDigest"`
	PreviousEventHash    []byte          `json:"previousEventHash"`
}

// public payload structs used for persistence and broadcast.
type valuePlacedPayload struct {
	SchemaVersion uint8  `json:"schemaVersion"`
	Cell          uint8  `json:"cell"`
	Value         uint8  `json:"value"`
	ParticipantID string `json:"participantId"`
	Correct       *bool  `json:"correct,omitempty"`
	Conflict      bool   `json:"conflict"`
	ReplacesValue bool   `json:"replacesValue"`
	IsHint        bool   `json:"isHint"`
}

type valueRejectedPayload struct {
	SchemaVersion uint8  `json:"schemaVersion"`
	Cell          uint8  `json:"cell"`
	Value         uint8  `json:"value"`
	ParticipantID string `json:"participantId"`
	Reason        string `json:"reason"`
	PenaltyMs     uint64 `json:"penaltyMs"`
}

type valueErasedPayload struct {
	SchemaVersion uint8  `json:"schemaVersion"`
	Cell          uint8  `json:"cell"`
	ParticipantID string `json:"participantId"`
}

type notesAddedPayload struct {
	SchemaVersion uint8  `json:"schemaVersion"`
	Cell          uint8  `json:"cell"`
	Digits        []int  `json:"digits"`
	ParticipantID string `json:"participantId"`
}

type notesRemovedPayload struct {
	SchemaVersion uint8  `json:"schemaVersion"`
	Cell          uint8  `json:"cell"`
	Digits        []int  `json:"digits"`
	ParticipantID string `json:"participantId"`
}

type notesAutoRemovedPayload struct {
	SchemaVersion uint8 `json:"schemaVersion"`
	Cell          uint8 `json:"cell"`
	Digits        []int `json:"digits"`
	CausedBy      uint8 `json:"causedBy"`
}

type hintUsedPayload struct {
	SchemaVersion uint8  `json:"schemaVersion"`
	Level         string `json:"level"`
	TargetCell    *uint8 `json:"targetCell,omitempty"`
	Value         *uint8 `json:"value,omitempty"`
	ParticipantID string `json:"participantId"`
}

type pingPayload struct {
	SchemaVersion uint8  `json:"schemaVersion"`
	Cell          uint8  `json:"cell"`
	Intent        string `json:"intent"`
	ParticipantID string `json:"participantId"`
}

type participantDisconnectedPayload struct {
	SchemaVersion uint8  `json:"schemaVersion"`
	ParticipantID string `json:"participantId"`
}

type participantReconnectedPayload struct {
	SchemaVersion uint8  `json:"schemaVersion"`
	ParticipantID string `json:"participantId"`
}

type matchCompletedPayload struct {
	SchemaVersion     uint8             `json:"schemaVersion"`
	ElapsedMs         uint64            `json:"elapsedMs"`
	Assisted          bool              `json:"assisted"`
	MistakesByPlayer  map[string]uint32 `json:"mistakesByPlayer"`
	HintCount         uint32            `json:"hintCount"`
	ContributionCount uint32            `json:"contributionCount"`
}

func matchEventsToGen(requestID shared.RequestID, m *matchdomain.Match, events []matchdomain.Event, previousHash []byte) ([]gen.MatchEvent, []byte, error) {
	if len(previousHash) == 0 {
		previousHash = append([]byte(nil), genesisHash...)
	}
	out := make([]gen.MatchEvent, 0, len(events))
	for index, e := range events {
		meta := e.Metadata()
		eventType := eventTypeName(e)
		payload, err := matchEventPublicPayload(e)
		if err != nil {
			return nil, nil, err
		}
		actorID := ""
		if actor := publicActorID(e); actor != "" {
			actorID = actor.String()
		}
		env := matchEventPublicEnvelope{
			ProofVersion:      proofVersion,
			MatchID:           m.ID.String(),
			EventNumber:       uint64(meta.EventNumber),
			AggregateVersion:  uint64(meta.AggregateVersion),
			PublicEventType:   eventType,
			PublicActorID:     actorID,
			OccurredAtMs:      meta.OccurredAt.Milliseconds(),
			PublicPayload:     payload,
			PreviousEventHash: previousHash,
		}
		hash, err := hashEnvelope(env)
		if err != nil {
			return nil, nil, err
		}
		eventRequestID := sql.NullString{}
		if index == 0 {
			eventRequestID = sqlNullString(requestID.String())
		}
		out = append(out, gen.MatchEvent{
			MatchID:              m.ID.String(),
			EventNumber:          int64(meta.EventNumber),
			AggregateVersion:     int64(meta.AggregateVersion),
			PublicEventType:      eventType,
			PublicActorID:        sqlNullString(actorID),
			RequestID:            eventRequestID,
			OccurredAtMs:         meta.OccurredAt.Milliseconds(),
			PublicPayloadJson:    string(payload),
			PrivatePayloadBlob:   nil,
			PrivatePayloadSalt:   nil,
			PrivatePayloadDigest: nil,
			PreviousHash:         previousHash,
			EventHash:            hash,
		})
		previousHash = hash
	}
	return out, previousHash, nil
}

func matchEventPublicPayload(e matchdomain.Event) ([]byte, error) {
	var payload any
	switch ev := e.(type) {
	case matchdomain.MatchPreparedEvent:
		payload = map[string]uint8{"schemaVersion": 1}
	case matchdomain.MatchCountdownStartedEvent:
		payload = map[string]any{"schemaVersion": 1, "deadlineAtMs": ev.DeadlineAt.Milliseconds()}
	case matchdomain.MatchStartedEvent:
		payload = map[string]uint8{"schemaVersion": 1}
	case matchdomain.ValuePlacedEvent:
		var correct *bool
		if ev.Correct != nil {
			v := *ev.Correct
			correct = &v
		}
		payload = valuePlacedPayload{
			SchemaVersion: 1,
			Cell:          uint8(ev.Cell),
			Value:         uint8(ev.Digit),
			ParticipantID: ev.ParticipantID.String(),
			Correct:       correct,
			Conflict:      ev.Conflict,
			ReplacesValue: ev.ReplacesValue,
			IsHint:        ev.IsHint,
		}
	case matchdomain.ValueRejectedEvent:
		payload = valueRejectedPayload{
			SchemaVersion: 1,
			Cell:          uint8(ev.Cell),
			Value:         uint8(ev.Digit),
			ParticipantID: ev.ParticipantID.String(),
			Reason:        ev.Reason,
			PenaltyMs:     ev.PenaltyMs,
		}
	case matchdomain.ValueErasedEvent:
		payload = valueErasedPayload{
			SchemaVersion: 1,
			Cell:          uint8(ev.Cell),
			ParticipantID: ev.ParticipantID.String(),
		}
	case matchdomain.NotesAddedEvent:
		payload = notesAddedPayload{
			SchemaVersion: 1,
			Cell:          uint8(ev.Cell),
			Digits:        digitsToInts(ev.Digits),
			ParticipantID: ev.ParticipantID.String(),
		}
	case matchdomain.NotesRemovedEvent:
		payload = notesRemovedPayload{
			SchemaVersion: 1,
			Cell:          uint8(ev.Cell),
			Digits:        digitsToInts(ev.Digits),
			ParticipantID: ev.ParticipantID.String(),
		}
	case matchdomain.NotesAutoRemovedEvent:
		payload = notesAutoRemovedPayload{
			SchemaVersion: 1,
			Cell:          uint8(ev.Cell),
			Digits:        digitsToInts(ev.Digits),
			CausedBy:      uint8(ev.CausedBy),
		}
	case matchdomain.HintUsedEvent:
		var targetCell *uint8
		if ev.TargetCell != nil {
			c := uint8(*ev.TargetCell)
			targetCell = &c
		}
		var value *uint8
		if ev.Digit != nil {
			v := uint8(*ev.Digit)
			value = &v
		}
		payload = hintUsedPayload{
			SchemaVersion: 1,
			Level:         string(ev.Level),
			TargetCell:    targetCell,
			Value:         value,
			ParticipantID: ev.ParticipantID.String(),
		}
	case matchdomain.PingEvent:
		payload = pingPayload{
			SchemaVersion: 1,
			Cell:          uint8(ev.Cell),
			Intent:        ev.Intent,
			ParticipantID: ev.ParticipantID.String(),
		}
	case matchdomain.ParticipantDisconnectedEvent:
		payload = participantDisconnectedPayload{
			SchemaVersion: 1,
			ParticipantID: ev.ParticipantID.String(),
		}
	case matchdomain.ParticipantReconnectedEvent:
		payload = participantReconnectedPayload{
			SchemaVersion: 1,
			ParticipantID: ev.ParticipantID.String(),
		}
	case matchdomain.MatchCompletedEvent:
		mistakes := make(map[string]uint32)
		for p, n := range ev.Result.MistakesByPlayer {
			mistakes[p.String()] = n
		}
		payload = matchCompletedPayload{
			SchemaVersion:     1,
			ElapsedMs:         ev.Result.ElapsedMilliseconds,
			Assisted:          ev.Result.Assisted,
			MistakesByPlayer:  mistakes,
			HintCount:         ev.Result.HintCount,
			ContributionCount: ev.Result.ContributionCount,
		}
	default:
		return nil, fmt.Errorf("unknown match event type %T", e)
	}
	return json.Marshal(payload)
}

func eventTypeName(e matchdomain.Event) string {
	switch e.(type) {
	case matchdomain.MatchPreparedEvent:
		return "MatchPrepared"
	case matchdomain.MatchCountdownStartedEvent:
		return "MatchCountdownStarted"
	case matchdomain.MatchStartedEvent:
		return "MatchStarted"
	case matchdomain.ValuePlacedEvent:
		return "ValuePlaced"
	case matchdomain.ValueRejectedEvent:
		return "ValueRejected"
	case matchdomain.ValueErasedEvent:
		return "ValueErased"
	case matchdomain.NotesAddedEvent:
		return "NotesAdded"
	case matchdomain.NotesRemovedEvent:
		return "NotesRemoved"
	case matchdomain.NotesAutoRemovedEvent:
		return "NotesAutoRemoved"
	case matchdomain.HintUsedEvent:
		return "HintUsed"
	case matchdomain.PingEvent:
		return "Ping"
	case matchdomain.ParticipantDisconnectedEvent:
		return "ParticipantDisconnected"
	case matchdomain.ParticipantReconnectedEvent:
		return "ParticipantReconnected"
	case matchdomain.MatchCompletedEvent:
		return "MatchCompleted"
	default:
		return "Unknown"
	}
}

func publicActorID(e matchdomain.Event) shared.ParticipantID {
	switch ev := e.(type) {
	case matchdomain.ValuePlacedEvent:
		return ev.ParticipantID
	case matchdomain.ValueRejectedEvent:
		return ev.ParticipantID
	case matchdomain.ValueErasedEvent:
		return ev.ParticipantID
	case matchdomain.NotesAddedEvent:
		return ev.ParticipantID
	case matchdomain.NotesRemovedEvent:
		return ev.ParticipantID
	case matchdomain.HintUsedEvent:
		return ev.ParticipantID
	case matchdomain.PingEvent:
		return ev.ParticipantID
	case matchdomain.ParticipantDisconnectedEvent:
		return ev.ParticipantID
	case matchdomain.ParticipantReconnectedEvent:
		return ev.ParticipantID
	default:
		return ""
	}
}

func hashEnvelope(env matchEventPublicEnvelope) ([]byte, error) {
	canonical, err := jcs.Transform(json.RawMessage(mustJSON(env)))
	if err != nil {
		return nil, err
	}
	h := sha256.Sum256(canonical)
	return h[:], nil
}

func mustJSON(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}

func digitsToInts(ds []shared.Digit) []int {
	out := make([]int, len(ds))
	for i, d := range ds {
		out[i] = int(d)
	}
	return out
}

func sqlNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

// matchEventFromGen reconstructs a domain event from its persisted public form.
func matchEventFromGen(e gen.MatchEvent) (matchdomain.Event, error) {
	meta, err := shared.NewEventMetadata(1, shared.EventNumber(e.EventNumber), uint64(e.AggregateVersion), shared.NewMatchTarget(shared.MatchID(e.MatchID)), mustTimestampMs(e.OccurredAtMs))
	if err != nil {
		return nil, err
	}
	switch e.PublicEventType {
	case "MatchPrepared":
		return matchdomain.MatchPreparedEvent{Meta: meta}, nil
	case "MatchCountdownStarted":
		var p struct {
			DeadlineAtMs int64 `json:"deadlineAtMs"`
		}
		if err := json.Unmarshal([]byte(e.PublicPayloadJson), &p); err != nil {
			return nil, err
		}
		ts, err := shared.TimestampFromMilliseconds(p.DeadlineAtMs)
		if err != nil {
			return nil, err
		}
		return matchdomain.MatchCountdownStartedEvent{Meta: meta, DeadlineAt: ts}, nil
	case "MatchStarted":
		return matchdomain.MatchStartedEvent{Meta: meta}, nil
	case "ValuePlaced":
		var p valuePlacedPayload
		if err := json.Unmarshal([]byte(e.PublicPayloadJson), &p); err != nil {
			return nil, err
		}
		return matchdomain.ValuePlacedEvent{
			Meta:          meta,
			Cell:          shared.CellIndex(p.Cell),
			Digit:         shared.Digit(p.Value),
			ParticipantID: shared.ParticipantID(p.ParticipantID),
			Correct:       p.Correct,
			Conflict:      p.Conflict,
			ReplacesValue: p.ReplacesValue,
			IsHint:        p.IsHint,
		}, nil
	case "ValueRejected":
		var p valueRejectedPayload
		if err := json.Unmarshal([]byte(e.PublicPayloadJson), &p); err != nil {
			return nil, err
		}
		return matchdomain.ValueRejectedEvent{
			Meta:          meta,
			Cell:          shared.CellIndex(p.Cell),
			Digit:         shared.Digit(p.Value),
			ParticipantID: shared.ParticipantID(p.ParticipantID),
			Reason:        p.Reason,
			PenaltyMs:     p.PenaltyMs,
		}, nil
	case "ValueErased":
		var p valueErasedPayload
		if err := json.Unmarshal([]byte(e.PublicPayloadJson), &p); err != nil {
			return nil, err
		}
		return matchdomain.ValueErasedEvent{
			Meta:          meta,
			Cell:          shared.CellIndex(p.Cell),
			ParticipantID: shared.ParticipantID(p.ParticipantID),
		}, nil
	case "NotesAdded":
		var p notesAddedPayload
		if err := json.Unmarshal([]byte(e.PublicPayloadJson), &p); err != nil {
			return nil, err
		}
		return matchdomain.NotesAddedEvent{
			Meta:          meta,
			Cell:          shared.CellIndex(p.Cell),
			Digits:        intsToDigits(p.Digits),
			ParticipantID: shared.ParticipantID(p.ParticipantID),
		}, nil
	case "NotesRemoved":
		var p notesRemovedPayload
		if err := json.Unmarshal([]byte(e.PublicPayloadJson), &p); err != nil {
			return nil, err
		}
		return matchdomain.NotesRemovedEvent{
			Meta:          meta,
			Cell:          shared.CellIndex(p.Cell),
			Digits:        intsToDigits(p.Digits),
			ParticipantID: shared.ParticipantID(p.ParticipantID),
		}, nil
	case "NotesAutoRemoved":
		var p notesAutoRemovedPayload
		if err := json.Unmarshal([]byte(e.PublicPayloadJson), &p); err != nil {
			return nil, err
		}
		return matchdomain.NotesAutoRemovedEvent{
			Meta:     meta,
			Cell:     shared.CellIndex(p.Cell),
			Digits:   intsToDigits(p.Digits),
			CausedBy: shared.CellIndex(p.CausedBy),
		}, nil
	case "HintUsed":
		var p hintUsedPayload
		if err := json.Unmarshal([]byte(e.PublicPayloadJson), &p); err != nil {
			return nil, err
		}
		ev := matchdomain.HintUsedEvent{
			Meta:          meta,
			Level:         shared.HintLevel(p.Level),
			ParticipantID: shared.ParticipantID(p.ParticipantID),
		}
		if p.TargetCell != nil {
			c := shared.CellIndex(*p.TargetCell)
			ev.TargetCell = &c
		}
		if p.Value != nil {
			v := shared.Digit(*p.Value)
			ev.Digit = &v
		}
		return ev, nil
	case "Ping":
		var p pingPayload
		if err := json.Unmarshal([]byte(e.PublicPayloadJson), &p); err != nil {
			return nil, err
		}
		return matchdomain.PingEvent{
			Meta:          meta,
			Cell:          shared.CellIndex(p.Cell),
			Intent:        p.Intent,
			ParticipantID: shared.ParticipantID(p.ParticipantID),
		}, nil
	case "ParticipantDisconnected":
		var p participantDisconnectedPayload
		if err := json.Unmarshal([]byte(e.PublicPayloadJson), &p); err != nil {
			return nil, err
		}
		return matchdomain.ParticipantDisconnectedEvent{
			Meta:          meta,
			ParticipantID: shared.ParticipantID(p.ParticipantID),
		}, nil
	case "ParticipantReconnected":
		var p participantReconnectedPayload
		if err := json.Unmarshal([]byte(e.PublicPayloadJson), &p); err != nil {
			return nil, err
		}
		return matchdomain.ParticipantReconnectedEvent{
			Meta:          meta,
			ParticipantID: shared.ParticipantID(p.ParticipantID),
		}, nil
	case "MatchCompleted":
		var p matchCompletedPayload
		if err := json.Unmarshal([]byte(e.PublicPayloadJson), &p); err != nil {
			return nil, err
		}
		mistakes := make(map[shared.ParticipantID]uint32)
		for id, n := range p.MistakesByPlayer {
			mistakes[shared.ParticipantID(id)] = n
		}
		return matchdomain.MatchCompletedEvent{
			Meta: meta,
			Result: matchdomain.Result{
				ElapsedMilliseconds: p.ElapsedMs,
				Assisted:            p.Assisted,
				MistakesByPlayer:    mistakes,
				HintCount:           p.HintCount,
				ContributionCount:   p.ContributionCount,
			},
		}, nil
	default:
		return nil, fmt.Errorf("unknown persisted event type %q", e.PublicEventType)
	}
}

func intsToDigits(u []int) []shared.Digit {
	out := make([]shared.Digit, len(u))
	for i, v := range u {
		out[i] = shared.Digit(v)
	}
	return out
}

func mustTimestampMs(ms int64) shared.Timestamp {
	ts, err := shared.TimestampFromMilliseconds(ms)
	if err != nil {
		panic(err)
	}
	return ts
}
