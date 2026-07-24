package domain

import (
	"time"

	shared "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/domain"
)

// Apply executes a gameplay command and returns the resulting domain events.
// nextEventNumber is the first available event number for this command.
func (m *Match) Apply(cmd Command, nextEventNumber uint64, now time.Time) ([]Event, error) {
	meta := cmd.Metadata()

	if meta.ExpectedVersion != uint64(m.Version) {
		return nil, shared.DomainError{Code: shared.ErrStaleVersion}
	}
	if _, ok := m.processedRequestIDs[meta.RequestID]; ok {
		return nil, nil
	}
	m.processedRequestIDs[meta.RequestID] = struct{}{}

	switch c := cmd.(type) {
	case PlaceValueCommand:
		return m.placeValue(c, nextEventNumber, now)
	case EraseValueCommand:
		return m.eraseValue(c, nextEventNumber, now)
	case AddNoteCommand:
		return m.addNotes(c, nextEventNumber, now)
	case RemoveNoteCommand:
		return m.removeNotes(c, nextEventNumber, now)
	case UseHintCommand:
		return m.useHint(c, nextEventNumber, now)
	case PingCommand:
		return m.ping(c, nextEventNumber, now)
	case ParticipantDisconnectedCommand:
		return m.participantDisconnected(c, nextEventNumber, now)
	case ParticipantReconnectedCommand:
		return m.participantReconnected(c, nextEventNumber, now)
	case EnterRecoveryCommand:
		return m.enterRecovery(c, nextEventNumber, now)
	case RecoverMatchCommand:
		return m.recover(c, nextEventNumber, now)
	case CancelRecoveryCommand:
		return m.cancelRecovery(c, nextEventNumber, now)
	default:
		return nil, shared.DomainError{Code: shared.ErrMatchCommandInvalid}
	}
}

func (m *Match) placeValue(c PlaceValueCommand, nextEventNumber uint64, now time.Time) ([]Event, error) {
	if err := m.requireGameplay(c.Meta, now); err != nil {
		return nil, err
	}
	if int(c.Cell) >= 81 || c.Cell < 0 {
		return nil, shared.DomainError{Code: shared.ErrCellIndexInvalid}
	}
	if c.Digit < 1 || c.Digit > 9 {
		return nil, shared.DomainError{Code: shared.ErrDigitInvalid}
	}
	if m.Puzzle.Clues[c.Cell] != 0 {
		return nil, shared.DomainError{Code: shared.ErrCellFixed}
	}
	if m.Puzzle.Solution == nil || len(m.Puzzle.Solution) != 81 {
		return nil, shared.DomainError{Code: shared.ErrMatchPuzzleInvalid}
	}

	existing, hasExisting := m.Values[c.Cell]
	if hasExisting && existing == c.Digit {
		return nil, nil
	}

	participantID := c.Meta.AuthenticatedParticipantID
	solutionDigit := shared.Digit(m.Puzzle.Solution[c.Cell])
	correct := c.Digit == solutionDigit
	conflict := hasDirectConflict(m.Values, c.Cell, c.Digit)
	wrong := !correct || conflict

	var events []Event

	switch m.Rules.ErrorPreset {
	case shared.ErrorPresetChallenge:
		if wrong {
			m.recordMistake(participantID)
			m.PenaltiesMs += 5000
			m.bumpVersion()
			meta, err := m.newEventMeta(nextEventNumber, now)
			if err != nil {
				return nil, err
			}
			events = append(events, ValueRejectedEvent{
				Meta:          meta,
				Cell:          c.Cell,
				Digit:         c.Digit,
				ParticipantID: participantID,
				Reason:        "challenge",
				PenaltyMs:     5000,
			})
			return events, nil
		}
	case shared.ErrorPresetClean:
		if wrong {
			m.bumpVersion()
			meta, err := m.newEventMeta(nextEventNumber, now)
			if err != nil {
				return nil, err
			}
			events = append(events, ValueRejectedEvent{
				Meta:          meta,
				Cell:          c.Cell,
				Digit:         c.Digit,
				ParticipantID: participantID,
				Reason:        "clean",
			})
			return events, nil
		}
	case shared.ErrorPresetBlind:
		if wrong {
			m.recordMistake(participantID)
		}
	case shared.ErrorPresetCasual:
		if wrong {
			m.recordMistake(participantID)
		}
	default:
		return nil, shared.DomainError{Code: shared.ErrMatchRulesInvalid}
	}

	m.bumpVersion()
	m.Values[c.Cell] = c.Digit
	m.Attribution[c.Cell] = participantID
	m.Cells[c.Cell].Value = &c.Digit
	m.Cells[c.Cell].Attribution = participantID
	if correct && !conflict {
		m.recordContribution(participantID)
	}

	correctPtr := &correct
	if m.Rules.ErrorPreset == shared.ErrorPresetBlind {
		correctPtr = nil
	}
	m.Cells[c.Cell].Correct = correctPtr
	meta, err := m.newEventMeta(nextEventNumber, now)
	if err != nil {
		return nil, err
	}
	valueEvent := ValuePlacedEvent{
		Meta:          meta,
		Cell:          c.Cell,
		Digit:         c.Digit,
		ParticipantID: participantID,
		Correct:       correctPtr,
		Conflict:      conflict,
		ReplacesValue: hasExisting,
		IsHint:        false,
	}
	nextEventNumber++
	events = append(events, valueEvent)

	// Clear notes in this cell.
	if m.Notes[c.Cell] != 0 {
		removed := m.Notes[c.Cell]
		m.Notes[c.Cell] = 0
		m.Cells[c.Cell].Notes = 0
		notesMeta, err := m.newEventMeta(nextEventNumber, now)
		if err != nil {
			return nil, err
		}
		events = append(events, NotesRemovedEvent{
			Meta:          notesMeta,
			Cell:          c.Cell,
			Digits:        removed.Digits(),
			ParticipantID: participantID,
		})
		nextEventNumber++
	}

	// Auto-remove matching candidates from peers.
	if m.Rules.AutoRemoveNotes {
		for _, peer := range cellPeers(c.Cell) {
			if m.Puzzle.Clues[peer] != 0 {
				continue
			}
			if m.Notes[peer].Contains(c.Digit) {
				m.Notes[peer] = m.Notes[peer].Remove(c.Digit)
				m.Cells[peer].Notes = m.Notes[peer]
				notesMeta, err := m.newEventMeta(nextEventNumber, now)
				if err != nil {
					return nil, err
				}
				events = append(events, NotesAutoRemovedEvent{
					Meta:     notesMeta,
					Cell:     peer,
					Digits:   []shared.Digit{c.Digit},
					CausedBy: c.Cell,
				})
				nextEventNumber++
			}
		}
	}

	if m.checkCompleted(now) {
		result := m.buildResult(now)
		m.Result = &result
		m.State = shared.MatchCompleted
		completeMeta, err := m.newEventMeta(nextEventNumber, now)
		if err != nil {
			return nil, err
		}
		events = append(events, MatchCompletedEvent{Meta: completeMeta, Result: result})
	}

	return events, nil
}

func (m *Match) eraseValue(c EraseValueCommand, nextEventNumber uint64, now time.Time) ([]Event, error) {
	if err := m.requireGameplay(c.Meta, now); err != nil {
		return nil, err
	}
	if m.Puzzle.Clues[c.Cell] != 0 {
		return nil, shared.DomainError{Code: shared.ErrCellFixed}
	}
	if _, ok := m.Values[c.Cell]; !ok {
		return nil, nil
	}
	m.bumpVersion()
	participantID := c.Meta.AuthenticatedParticipantID
	delete(m.Values, c.Cell)
	delete(m.Attribution, c.Cell)
	m.Cells[c.Cell].Value = nil
	m.Cells[c.Cell].Attribution = ""
	m.Cells[c.Cell].Correct = nil
	meta, err := m.newEventMeta(nextEventNumber, now)
	if err != nil {
		return nil, err
	}
	return []Event{ValueErasedEvent{Meta: meta, Cell: c.Cell, ParticipantID: participantID}}, nil
}

func (m *Match) addNotes(c AddNoteCommand, nextEventNumber uint64, now time.Time) ([]Event, error) {
	if err := m.requireGameplay(c.Meta, now); err != nil {
		return nil, err
	}
	if m.Puzzle.Clues[c.Cell] != 0 {
		return nil, shared.DomainError{Code: shared.ErrCellFixed}
	}
	var added []shared.Digit
	for _, d := range c.Digits {
		if d < 1 || d > 9 {
			continue
		}
		if !m.Notes[c.Cell].Contains(d) {
			m.Notes[c.Cell] = m.Notes[c.Cell].Add(d)
			m.Cells[c.Cell].Notes = m.Notes[c.Cell]
			added = append(added, d)
		}
	}
	if len(added) == 0 {
		return nil, nil
	}
	m.bumpVersion()
	meta, err := m.newEventMeta(nextEventNumber, now)
	if err != nil {
		return nil, err
	}
	return []Event{NotesAddedEvent{
		Meta:          meta,
		Cell:          c.Cell,
		Digits:        added,
		ParticipantID: c.Meta.AuthenticatedParticipantID,
	}}, nil
}

func (m *Match) removeNotes(c RemoveNoteCommand, nextEventNumber uint64, now time.Time) ([]Event, error) {
	if err := m.requireGameplay(c.Meta, now); err != nil {
		return nil, err
	}
	if m.Puzzle.Clues[c.Cell] != 0 {
		return nil, shared.DomainError{Code: shared.ErrCellFixed}
	}
	var removed []shared.Digit
	for _, d := range c.Digits {
		if d < 1 || d > 9 {
			continue
		}
		if m.Notes[c.Cell].Contains(d) {
			m.Notes[c.Cell] = m.Notes[c.Cell].Remove(d)
			m.Cells[c.Cell].Notes = m.Notes[c.Cell]
			removed = append(removed, d)
		}
	}
	if len(removed) == 0 {
		return nil, nil
	}
	m.bumpVersion()
	meta, err := m.newEventMeta(nextEventNumber, now)
	if err != nil {
		return nil, err
	}
	return []Event{NotesRemovedEvent{
		Meta:          meta,
		Cell:          c.Cell,
		Digits:        removed,
		ParticipantID: c.Meta.AuthenticatedParticipantID,
	}}, nil
}

func (m *Match) useHint(c UseHintCommand, nextEventNumber uint64, now time.Time) ([]Event, error) {
	if err := m.requireGameplay(c.Meta, now); err != nil {
		return nil, err
	}
	if !m.Rules.HintsEnabled {
		return nil, shared.DomainError{Code: shared.ErrHintsDisabled}
	}
	if m.Puzzle.Solution == nil || len(m.Puzzle.Solution) != 81 {
		return nil, shared.DomainError{Code: shared.ErrMatchPuzzleInvalid}
	}

	m.Assisted = true
	m.HintsUsed++
	m.bumpVersion()

	var events []Event

	switch c.Level {
	case shared.HintNudge:
		empty := emptyCells(m.Values, m.Puzzle.Clues)
		if len(empty) == 0 {
			return nil, shared.DomainError{Code: shared.ErrHintLevelUnavailable}
		}
		var target shared.CellIndex
		if c.Target != nil && m.Puzzle.Clues[*c.Target] == 0 {
			target = *c.Target
		} else {
			target = empty[0]
		}
		meta, err := m.newEventMeta(nextEventNumber, now)
		if err != nil {
			return nil, err
		}
		return []Event{HintUsedEvent{Meta: meta, Level: shared.HintNudge, TargetCell: &target, ParticipantID: c.Meta.AuthenticatedParticipantID}}, nil
	case shared.HintReveal:
		empty := emptyCells(m.Values, m.Puzzle.Clues)
		if len(empty) == 0 {
			return nil, shared.DomainError{Code: shared.ErrHintLevelUnavailable}
		}
		cell := empty[0]
		digit := shared.Digit(m.Puzzle.Solution[cell])
		if digit == 0 || digit > 9 {
			return nil, shared.DomainError{Code: shared.ErrMatchPuzzleInvalid}
		}
		meta, err := m.newEventMeta(nextEventNumber, now)
		if err != nil {
			return nil, err
		}
		events = append(events, HintUsedEvent{
			Meta:          meta,
			Level:         shared.HintReveal,
			TargetCell:    &cell,
			Digit:         &digit,
			ParticipantID: c.Meta.AuthenticatedParticipantID,
		})
		nextEventNumber++

		// Reveal the value as a system-like placement.
		correct := true
		m.Values[cell] = digit
		m.Attribution[cell] = c.Meta.AuthenticatedParticipantID
		m.Cells[cell].Value = &digit
		m.Cells[cell].Attribution = c.Meta.AuthenticatedParticipantID
		m.Cells[cell].Correct = &correct
		placeMeta, err := m.newEventMeta(nextEventNumber, now)
		if err != nil {
			return nil, err
		}
		events = append(events, ValuePlacedEvent{
			Meta:          placeMeta,
			Cell:          cell,
			Digit:         digit,
			ParticipantID: c.Meta.AuthenticatedParticipantID,
			Correct:       &correct,
			Conflict:      false,
			ReplacesValue: false,
			IsHint:        true,
		})
		nextEventNumber++

		if m.Notes[cell] != 0 {
			removed := m.Notes[cell]
			m.Notes[cell] = 0
			m.Cells[cell].Notes = 0
			notesMeta, err := m.newEventMeta(nextEventNumber, now)
			if err != nil {
				return nil, err
			}
			events = append(events, NotesRemovedEvent{
				Meta:          notesMeta,
				Cell:          cell,
				Digits:        removed.Digits(),
				ParticipantID: c.Meta.AuthenticatedParticipantID,
			})
			nextEventNumber++
		}
		if m.Rules.AutoRemoveNotes {
			for _, peer := range cellPeers(cell) {
				if m.Puzzle.Clues[peer] != 0 {
					continue
				}
				if m.Notes[peer].Contains(digit) {
					m.Notes[peer] = m.Notes[peer].Remove(digit)
					notesMeta, err := m.newEventMeta(nextEventNumber, now)
					if err != nil {
						return nil, err
					}
					events = append(events, NotesAutoRemovedEvent{
						Meta:     notesMeta,
						Cell:     peer,
						Digits:   []shared.Digit{digit},
						CausedBy: cell,
					})
					nextEventNumber++
				}
			}
		}
		if m.checkCompleted(now) {
			result := m.buildResult(now)
			m.Result = &result
			m.State = shared.MatchCompleted
			completeMeta, err := m.newEventMeta(nextEventNumber, now)
			if err != nil {
				return nil, err
			}
			events = append(events, MatchCompletedEvent{Meta: completeMeta, Result: result})
		}
		return events, nil
	default:
		return nil, shared.DomainError{Code: shared.ErrHintLevelUnavailable}
	}
}

func (m *Match) ping(c PingCommand, nextEventNumber uint64, now time.Time) ([]Event, error) {
	if err := m.requireGameplay(c.Meta, now); err != nil {
		return nil, err
	}
	if int(c.Cell) >= 81 {
		return nil, shared.DomainError{Code: shared.ErrCellIndexInvalid}
	}
	if c.Intent == "" {
		return nil, shared.DomainError{Code: shared.ErrInvalidValue}
	}
	meta, err := m.newEventMeta(nextEventNumber, now)
	if err != nil {
		return nil, err
	}
	return []Event{PingEvent{
		Meta:          meta,
		Cell:          c.Cell,
		Intent:        c.Intent,
		ParticipantID: c.Meta.AuthenticatedParticipantID,
	}}, nil
}

func (m *Match) participantDisconnected(c ParticipantDisconnectedCommand, nextEventNumber uint64, now time.Time) ([]Event, error) {
	if !m.isParticipant(c.ParticipantID) {
		return nil, shared.DomainError{Code: shared.ErrParticipantNotFound}
	}
	meta, err := m.newEventMeta(nextEventNumber, now)
	if err != nil {
		return nil, err
	}
	return []Event{ParticipantDisconnectedEvent{Meta: meta, ParticipantID: c.ParticipantID}}, nil
}

func (m *Match) participantReconnected(c ParticipantReconnectedCommand, nextEventNumber uint64, now time.Time) ([]Event, error) {
	if !m.isParticipant(c.ParticipantID) {
		return nil, shared.DomainError{Code: shared.ErrParticipantNotFound}
	}
	meta, err := m.newEventMeta(nextEventNumber, now)
	if err != nil {
		return nil, err
	}
	return []Event{ParticipantReconnectedEvent{Meta: meta, ParticipantID: c.ParticipantID}}, nil
}

func (m *Match) enterRecovery(c EnterRecoveryCommand, nextEventNumber uint64, now time.Time) ([]Event, error) {
	if m.State != shared.MatchActive {
		return nil, shared.DomainError{Code: shared.ErrMatchStateInvalid}
	}
	if c.Generation == 0 || c.Generation <= m.RecoveryGeneration {
		return nil, shared.DomainError{Code: shared.ErrTimerTokenStale}
	}
	startedAt := mustTimestamp(now)
	previous := m.State
	m.State = shared.MatchRecoveryPending
	m.RecoveryPreviousState = previous
	m.RecoveryStartedAt = &startedAt
	m.RecoveryGeneration = c.Generation
	m.bumpVersion()
	meta, err := m.newEventMeta(nextEventNumber, now)
	if err != nil {
		return nil, err
	}
	return []Event{MatchEnteredRecoveryEvent{
		Meta:          meta,
		Generation:    c.Generation,
		PreviousState: previous,
		StartedAt:     startedAt,
	}}, nil
}

func (m *Match) recover(c RecoverMatchCommand, nextEventNumber uint64, now time.Time) ([]Event, error) {
	if m.State != shared.MatchRecoveryPending || m.RecoveryStartedAt == nil {
		return nil, shared.DomainError{Code: shared.ErrMatchStateInvalid}
	}
	if c.Generation != m.RecoveryGeneration {
		return nil, shared.DomainError{Code: shared.ErrTimerTokenStale}
	}
	paused := now.UnixMilli() - m.RecoveryStartedAt.Milliseconds()
	if paused < 0 {
		return nil, shared.DomainError{Code: shared.ErrRecoveryFailed}
	}
	m.PausedMilliseconds += uint64(paused)
	if m.RecoveryPreviousState != shared.MatchActive {
		return nil, shared.DomainError{Code: shared.ErrRecoveryFailed}
	}
	m.State = m.RecoveryPreviousState
	m.RecoveryStartedAt = nil
	m.RecoveryPreviousState = ""
	m.bumpVersion()
	recoveredAt := mustTimestamp(now)
	meta, err := m.newEventMeta(nextEventNumber, now)
	if err != nil {
		return nil, err
	}
	return []Event{MatchRecoveredEvent{
		Meta:             meta,
		Generation:       c.Generation,
		PausedIntervalMs: uint64(paused),
		RecoveredAt:      recoveredAt,
	}}, nil
}

func (m *Match) cancelRecovery(c CancelRecoveryCommand, nextEventNumber uint64, now time.Time) ([]Event, error) {
	if m.State != shared.MatchRecoveryPending {
		return nil, shared.DomainError{Code: shared.ErrMatchStateInvalid}
	}
	if c.Generation != m.RecoveryGeneration {
		return nil, shared.DomainError{Code: shared.ErrTimerTokenStale}
	}
	if m.RecoveryStartedAt == nil {
		return nil, shared.DomainError{Code: shared.ErrRecoveryFailed}
	}
	paused := now.UnixMilli() - m.RecoveryStartedAt.Milliseconds()
	if paused < 0 {
		return nil, shared.DomainError{Code: shared.ErrRecoveryFailed}
	}
	m.PausedMilliseconds += uint64(paused)
	m.State = shared.MatchCancelled
	completedAt := mustTimestamp(now)
	m.CompletedAt = &completedAt
	result := m.buildResult(now)
	result.Reason = "RecoveryFailure"
	m.Result = &result
	m.RecoveryStartedAt = nil
	m.RecoveryPreviousState = ""
	m.bumpVersion()
	meta, err := m.newEventMeta(nextEventNumber, now)
	if err != nil {
		return nil, err
	}
	return []Event{MatchCancelledEvent{
		Meta:       meta,
		Generation: c.Generation,
		Reason:     "RecoveryFailure",
	}}, nil
}

func (m *Match) requireGameplay(meta shared.CommandMetadata, now time.Time) error {
	if m.State != shared.MatchActive {
		return shared.DomainError{Code: shared.ErrMatchStateInvalid}
	}
	if !m.isParticipant(meta.AuthenticatedParticipantID) {
		return shared.DomainError{Code: shared.ErrParticipantNotFound}
	}
	return nil
}

func (m *Match) newEventMeta(eventNumber uint64, now time.Time) (shared.EventMetadata, error) {
	en, err := shared.NewEventNumber(eventNumber)
	if err != nil {
		return shared.EventMetadata{}, err
	}
	return shared.NewEventMetadata(1, en, uint64(m.Version), shared.NewMatchTarget(m.ID), mustTimestamp(now))
}

func (m *Match) recordMistake(p shared.ParticipantID) {
	m.Mistakes[p]++
}

func (m *Match) recordContribution(p shared.ParticipantID) {
	m.Contributions[p]++
}

func (m *Match) checkCompleted(now time.Time) bool {
	if len(m.Puzzle.Solution) != 81 || len(m.Puzzle.Clues) != 81 {
		return false
	}
	return boardComplete(m.Values, m.Puzzle.Clues) && boardCorrect(m.Values, m.Puzzle.Clues, m.Puzzle.Solution)
}

func (m *Match) buildResult(now time.Time) Result {
	var elapsed uint64
	if m.StartedAt != nil {
		wall := mustTimestamp(now).Milliseconds() - m.StartedAt.Milliseconds()
		if wall > int64(m.PausedMilliseconds) {
			elapsed = uint64(wall) - m.PausedMilliseconds
		}
	}
	mistakes := make(map[shared.ParticipantID]uint32)
	for p, n := range m.Mistakes {
		mistakes[p] = n
	}
	var contributions uint32
	for _, n := range m.Contributions {
		contributions += n
	}
	return Result{
		Reason:                "PuzzleCompleted",
		CompletedAt:           mustTimestamp(now),
		ElapsedMilliseconds:   elapsed,
		PenaltyMilliseconds:   m.PenaltiesMs,
		Assisted:              m.Assisted,
		MistakesByPlayer:      mistakes,
		ContributionsByPlayer: cloneCountMap(m.Contributions),
		HintCount:             m.HintsUsed,
		ContributionCount:     contributions,
	}
}

// ApplyEvent applies a single durable event to the Match state. It is used for
// replay and recovery reconstruction. It ignores request-idempotency checks.
func (m *Match) ApplyEvent(e Event) error {
	switch ev := e.(type) {
	case MatchPreparedEvent:
		m.State = shared.MatchPrepared
	case MatchCountdownStartedEvent:
		m.State = shared.MatchCountdown
	case MatchStartedEvent:
		m.State = shared.MatchActive
		startedAt := ev.Meta.OccurredAt
		m.StartedAt = &startedAt
	case ValuePlacedEvent:
		m.Values[ev.Cell] = ev.Digit
		m.Cells[ev.Cell].Value = &ev.Digit
		m.Cells[ev.Cell].Attribution = ev.ParticipantID
		m.Cells[ev.Cell].Correct = ev.Correct
		m.Attribution[ev.Cell] = ev.ParticipantID
		if ev.IsHint {
			m.Assisted = true
		} else if !ev.IsHint && ev.Correct != nil && *ev.Correct && !ev.Conflict {
			m.Contributions[ev.ParticipantID]++
		}
		if (ev.Correct != nil && !*ev.Correct) || ev.Conflict || (ev.Correct == nil && m.Puzzle.Solution[ev.Cell] != byte(ev.Digit)) {
			m.Mistakes[ev.ParticipantID]++
		}
	case ValueRejectedEvent:
		m.Mistakes[ev.ParticipantID]++
		m.PenaltiesMs += ev.PenaltyMs
	case ValueErasedEvent:
		delete(m.Values, ev.Cell)
		delete(m.Attribution, ev.Cell)
		m.Cells[ev.Cell].Value = nil
		m.Cells[ev.Cell].Attribution = ""
		m.Cells[ev.Cell].Correct = nil
	case NotesAddedEvent:
		for _, d := range ev.Digits {
			m.Notes[ev.Cell] = m.Notes[ev.Cell].Add(d)
		}
		m.Cells[ev.Cell].Notes = m.Notes[ev.Cell]
	case NotesRemovedEvent:
		for _, d := range ev.Digits {
			m.Notes[ev.Cell] = m.Notes[ev.Cell].Remove(d)
		}
		m.Cells[ev.Cell].Notes = m.Notes[ev.Cell]
	case NotesAutoRemovedEvent:
		for _, d := range ev.Digits {
			m.Notes[ev.Cell] = m.Notes[ev.Cell].Remove(d)
		}
		m.Cells[ev.Cell].Notes = m.Notes[ev.Cell]
	case HintUsedEvent:
		m.HintsUsed++
		m.Assisted = true
	case PingEvent:
	case ParticipantDisconnectedEvent:
	case ParticipantReconnectedEvent:
	case MatchEnteredRecoveryEvent:
		if m.State == shared.MatchCompleted {
			return shared.DomainError{Code: shared.ErrMatchStateInvalid}
		}
		m.State = shared.MatchRecoveryPending
		m.RecoveryGeneration = ev.Generation
		m.RecoveryPreviousState = ev.PreviousState
		startedAt := ev.StartedAt
		m.RecoveryStartedAt = &startedAt
	case MatchRecoveredEvent:
		if m.State != shared.MatchRecoveryPending || ev.Generation != m.RecoveryGeneration {
			return shared.DomainError{Code: shared.ErrRecoveryFailed}
		}
		m.PausedMilliseconds += ev.PausedIntervalMs
		m.State = m.RecoveryPreviousState
		m.RecoveryPreviousState = ""
		m.RecoveryStartedAt = nil
	case MatchCancelledEvent:
		if m.State == shared.MatchCompleted {
			return shared.DomainError{Code: shared.ErrMatchStateInvalid}
		}
		m.State = shared.MatchCancelled
		completedAt := ev.Meta.OccurredAt
		m.CompletedAt = &completedAt
		m.RecoveryStartedAt = nil
		m.RecoveryPreviousState = ""
	case MatchCompletedEvent:
		m.State = shared.MatchCompleted
		m.Result = &ev.Result
		m.CompletedAt = &ev.Result.CompletedAt
	default:
		return shared.DomainError{Code: shared.ErrMatchCommandInvalid}
	}
	return nil
}

// ReconstructMatch creates a Match from an immutable assignment and applies
// the supplied events in order. It returns an error if an event is unsupported.
func ReconstructMatch(puzzle shared.AssignedPuzzle, rules Rules, participants []shared.ParticipantID, events []Event) (*Match, error) {
	m := &Match{
		ID:                  "",
		RoomID:              "",
		Version:             1,
		State:               shared.MatchPrepared,
		Rules:               rules,
		Puzzle:              puzzle,
		Participants:        participants,
		Values:              make(map[shared.CellIndex]shared.Digit),
		Attribution:         make(map[shared.CellIndex]shared.ParticipantID),
		Mistakes:            make(map[shared.ParticipantID]uint32),
		Contributions:       make(map[shared.ParticipantID]uint32),
		processedRequestIDs: make(map[shared.RequestID]struct{}),
	}
	m.initCells()
	for _, e := range events {
		if err := m.ApplyEvent(e); err != nil {
			return nil, err
		}
		m.Version = shared.MatchVersion(e.Metadata().AggregateVersion)
	}
	return m, nil
}
