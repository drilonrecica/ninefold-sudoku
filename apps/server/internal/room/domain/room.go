package domain

import (
	"time"

	shared "github.com/drilonrecica/ninefold-sudoku/apps/server/internal/domain"
)

const countdownDuration = 3 * time.Second

// capacity holds mode-specific player and spectator limits.
type capacity struct {
	MaxPlayers    int
	MinPlayers    int
	MaxSpectators int
}

func modeCapacity(mode shared.Mode) capacity {
	switch mode {
	case shared.ModeCoop:
		return capacity{MaxPlayers: 6, MinPlayers: 1, MaxSpectators: 10}
	default:
		// Unknown or provisional modes default to Co-op limits for safety.
		return capacity{MaxPlayers: 6, MinPlayers: 1, MaxSpectators: 10}
	}
}

// Apply mutates the room according to the supplied command and returns the
// resulting durable events. The caller must work on a copy of committed state
// so that persistence failures do not corrupt the authoritative aggregate.
func (r *Room) Apply(cmd Command, now time.Time) ([]Event, error) {
	if r == nil {
		return nil, shared.DomainError{Code: shared.ErrRoomStateInvalid}
	}

	meta := cmd.Metadata()
	if meta.ExpectedVersion != uint64(r.Version) {
		return nil, shared.DomainError{Code: shared.ErrStaleVersion}
	}

	switch c := cmd.(type) {
	case RequestJoinCommand:
		return r.applyRequestJoin(c, now)
	case LeaveRoomCommand:
		return r.applyLeaveRoom(c, now)
	case ChangeParticipationRoleCommand:
		return r.applyChangeRole(c, now)
	case SetReadyCommand:
		return r.applySetReady(c, now)
	case ChangeRoomSettingsCommand:
		return r.applyChangeSettings(c, now)
	case LockRoomCommand:
		return r.applyLock(c, now)
	case UnlockRoomCommand:
		return r.applyUnlock(c, now)
	case RemoveParticipantCommand:
		return r.applyRemoveParticipant(c, now)
	case BlockParticipantCommand:
		return r.applyBlockParticipant(c, now)
	case TransferHostCommand:
		return r.applyTransferHost(c, now)
	case StartCountdownCommand:
		return r.applyStartCountdown(c, now)
	case CancelCountdownCommand:
		return r.applyCancelCountdown(c, now)
	case ActivateMatchCommand:
		return r.applyActivateMatch(c, now)
	case ExpireRoomCommand:
		return r.applyExpireRoom(c, now)
	default:
		return nil, shared.DomainError{Code: shared.ErrRoomStateInvalid}
	}
}

// NewRoom constructs a room in the Lobby state with a single host participant.
func NewRoom(id shared.RoomID, code shared.RoomCode, host Participant, rules MatchRules, now time.Time) (*Room, error) {
	if rules.Mode != shared.ModeCoop {
		return nil, shared.DomainError{Code: shared.ErrRoomStateInvalid}
	}
	ts, err := nowTimestamp(now)
	if err != nil {
		return nil, err
	}
	version, err := shared.NewRoomVersion(1)
	if err != nil {
		return nil, err
	}
	expires, err := shared.NewTimestamp(now.Add(2 * time.Hour))
	if err != nil {
		return nil, err
	}
	host.IsHost = true
	host.IsReady = false
	return &Room{
		ID:                id,
		Code:              code,
		Version:           version,
		State:             shared.RoomLobby,
		Participants:      []Participant{host},
		Rules:             rules,
		HostParticipantID: &host.ID,
		CreatedAt:         ts,
		LastActivityAt:    ts,
		ExpiresAt:         expires,
	}, nil
}

// CompleteMatch transitions an active Room to Results exactly once.
func (r *Room) CompleteMatch(matchID shared.MatchID, now time.Time) ([]Event, error) {
	if r.State != shared.RoomInMatch || r.CurrentMatchID == nil || *r.CurrentMatchID != matchID {
		return nil, shared.DomainError{Code: shared.ErrRoomStateInvalid}
	}
	r.State = shared.RoomResults
	r.bumpVersion()
	r.touch(now)
	meta, err := r.nextEventMeta(now)
	if err != nil {
		return nil, err
	}
	return []Event{RoomEnteredResultsEvent{Meta: meta, MatchID: matchID}}, nil
}

func (r *Room) applyRequestJoin(cmd RequestJoinCommand, now time.Time) ([]Event, error) {
	if !r.isState(shared.RoomLobby, shared.RoomInMatch) {
		return nil, shared.DomainError{Code: shared.ErrRoomStateInvalid}
	}
	if r.isState(shared.RoomInMatch) && cmd.Role == shared.RolePlayer {
		return nil, shared.DomainError{Code: shared.ErrRoomStateInvalid}
	}
	if r.isLocked() {
		return nil, shared.DomainError{Code: shared.ErrRoomLocked}
	}

	// Display names must be unique among active participants.
	if existing := r.findActiveByComparisonKey(cmd.DisplayName.ComparisonKey()); existing != nil {
		_ = existing
		return nil, shared.DomainError{Code: shared.ErrNameAlreadyUsed}
	}

	cap := modeCapacity(r.Rules.Mode)
	players, spectators := r.counts()

	requestedRole := cmd.Role
	if requestedRole == "" {
		requestedRole = shared.RolePlayer
	}

	if requestedRole == shared.RolePlayer && players >= cap.MaxPlayers {
		details := map[string]any{"spectatorAvailable": r.Rules.SpectatorsAllowed && spectators < cap.MaxSpectators}
		return nil, shared.DomainError{Code: shared.ErrRoomFull, Details: details}
	}
	if requestedRole == shared.RoleSpectator {
		if !r.Rules.SpectatorsAllowed {
			return nil, shared.DomainError{Code: shared.ErrRoleChangeInvalid}
		}
		if spectators >= cap.MaxSpectators {
			return nil, shared.DomainError{Code: shared.ErrSpectatorCapacityReached}
		}
	}

	participant := Participant{
		ID:       cmd.ParticipantID,
		Name:     cmd.DisplayName,
		Role:     requestedRole,
		JoinedAt: mustTimestampFromTime(now),
	}
	r.Participants = append(r.Participants, participant)
	r.bumpVersion()
	r.touch(now)

	meta, err := r.nextEventMeta(now)
	if err != nil {
		return nil, err
	}
	return []Event{ParticipantJoinedEvent{Meta: meta, Participant: participant, Role: requestedRole}}, nil
}

func (r *Room) applyLeaveRoom(cmd LeaveRoomCommand, now time.Time) ([]Event, error) {
	if !r.isState(shared.RoomLobby, shared.RoomResults) {
		return nil, shared.DomainError{Code: shared.ErrRoomStateInvalid}
	}
	p := r.findActive(cmd.Meta.AuthenticatedParticipantID)
	if p == nil {
		return nil, shared.DomainError{Code: shared.ErrParticipantNotFound}
	}
	p.LeftAt = ptrTimestamp(mustTimestampFromTime(now))
	r.updateParticipant(*p)
	r.bumpVersion()
	r.touch(now)

	events := make([]Event, 0, 2)
	meta, err := r.nextEventMeta(now)
	if err != nil {
		return nil, err
	}
	events = append(events, ParticipantLeftEvent{Meta: meta, ParticipantID: p.ID, Intent: cmd.Intent})

	// If the host leaves, transfer authority to the longest-present active participant.
	if p.IsHost {
		if nextHost := r.longestActiveParticipant(nil); nextHost != nil {
			oldHost := r.HostParticipantID
			r.clearHost()
			nextHost.IsHost = true
			r.HostParticipantID = &nextHost.ID
			r.updateParticipant(*nextHost)
			meta2, err := r.nextEventMeta(now)
			if err != nil {
				return nil, err
			}
			events = append(events, HostTransferredEvent{Meta: meta2, From: oldHost, To: nextHost.ID})
		} else {
			r.clearHost()
		}
	}

	return events, nil
}

func (r *Room) applyChangeRole(cmd ChangeParticipationRoleCommand, now time.Time) ([]Event, error) {
	if !r.isState(shared.RoomLobby) {
		return nil, shared.DomainError{Code: shared.ErrRoomStateInvalid}
	}
	if cmd.Meta.AuthenticatedParticipantID != cmd.Participant {
		return nil, shared.DomainError{Code: shared.ErrActionNotAllowedForRole}
	}
	p := r.findActive(cmd.Participant)
	if p == nil {
		return nil, shared.DomainError{Code: shared.ErrParticipantNotFound}
	}
	from := p.Role
	if from == cmd.Role {
		return nil, nil
	}

	cap := modeCapacity(r.Rules.Mode)
	players, spectators := r.counts()
	if cmd.Role == shared.RolePlayer && players >= cap.MaxPlayers {
		return nil, shared.DomainError{Code: shared.ErrRoomFull}
	}
	if cmd.Role == shared.RoleSpectator && spectators >= cap.MaxSpectators {
		return nil, shared.DomainError{Code: shared.ErrSpectatorCapacityReached}
	}

	p.Role = cmd.Role
	p.IsReady = false
	r.updateParticipant(*p)
	r.bumpVersion()
	r.touch(now)

	meta, err := r.nextEventMeta(now)
	if err != nil {
		return nil, err
	}
	return []Event{ParticipantRoleChangedEvent{Meta: meta, ParticipantID: p.ID, From: from, To: cmd.Role}}, nil
}

func (r *Room) applySetReady(cmd SetReadyCommand, now time.Time) ([]Event, error) {
	if !r.isState(shared.RoomLobby) {
		return nil, shared.DomainError{Code: shared.ErrRoomStateInvalid}
	}
	p := r.findActive(cmd.Meta.AuthenticatedParticipantID)
	if p == nil {
		return nil, shared.DomainError{Code: shared.ErrParticipantNotFound}
	}
	if p.Role != shared.RolePlayer {
		return nil, shared.DomainError{Code: shared.ErrActionNotAllowedForRole}
	}
	if p.IsReady == cmd.Ready {
		return nil, nil
	}
	p.IsReady = cmd.Ready
	r.updateParticipant(*p)
	r.bumpVersion()
	r.touch(now)

	meta, err := r.nextEventMeta(now)
	if err != nil {
		return nil, err
	}
	return []Event{ParticipantReadyStateChangedEvent{Meta: meta, ParticipantID: p.ID, Ready: cmd.Ready}}, nil
}

func (r *Room) applyChangeSettings(cmd ChangeRoomSettingsCommand, now time.Time) ([]Event, error) {
	if !r.isState(shared.RoomLobby) {
		return nil, shared.DomainError{Code: shared.ErrRoomStateInvalid}
	}
	if err := r.requireHost(cmd.Meta.AuthenticatedParticipantID); err != nil {
		return nil, err
	}

	patched := r.Rules
	changed := false
	if cmd.Settings.Difficulty != nil {
		patched.Difficulty = *cmd.Settings.Difficulty
		changed = true
	}
	if cmd.Settings.ErrorPreset != nil {
		patched.ErrorPreset = *cmd.Settings.ErrorPreset
		changed = true
	}
	if cmd.Settings.HintsEnabled != nil {
		patched.HintsEnabled = *cmd.Settings.HintsEnabled
		changed = true
	}
	if cmd.Settings.SharedNotes != nil {
		patched.SharedNotes = *cmd.Settings.SharedNotes
		changed = true
	}
	if cmd.Settings.AutoRemoveNotes != nil {
		patched.AutoRemoveNotes = *cmd.Settings.AutoRemoveNotes
		changed = true
	}
	if cmd.Settings.SpectatorsAllowed != nil {
		patched.SpectatorsAllowed = *cmd.Settings.SpectatorsAllowed
		changed = true
	}
	if !changed {
		return nil, nil
	}

	// Any gameplay setting change resets every player's ready state.
	r.Rules = patched
	for i := range r.Participants {
		if r.Participants[i].IsActive() && r.Participants[i].Role == shared.RolePlayer {
			r.Participants[i].IsReady = false
		}
	}
	r.bumpVersion()
	r.touch(now)

	meta1, err := r.nextEventMeta(now)
	if err != nil {
		return nil, err
	}
	meta2, err := r.nextEventMeta(now)
	if err != nil {
		return nil, err
	}
	return []Event{
		RoomSettingsChangedEvent{Meta: meta1, Settings: patched},
		RoomReadyStatesResetEvent{Meta: meta2},
	}, nil
}

func (r *Room) applyLock(cmd LockRoomCommand, now time.Time) ([]Event, error) {
	if !r.isState(shared.RoomLobby) {
		return nil, shared.DomainError{Code: shared.ErrRoomStateInvalid}
	}
	if err := r.requireHost(cmd.Meta.AuthenticatedParticipantID); err != nil {
		return nil, err
	}
	if r.isLocked() {
		return nil, nil
	}
	r.Rules.SpectatorsAllowed = false
	r.bumpVersion()
	r.touch(now)
	meta, err := r.nextEventMeta(now)
	if err != nil {
		return nil, err
	}
	return []Event{RoomLockedEvent{Meta: meta}}, nil
}

func (r *Room) applyUnlock(cmd UnlockRoomCommand, now time.Time) ([]Event, error) {
	if !r.isState(shared.RoomLobby) {
		return nil, shared.DomainError{Code: shared.ErrRoomStateInvalid}
	}
	if err := r.requireHost(cmd.Meta.AuthenticatedParticipantID); err != nil {
		return nil, err
	}
	if !r.isLocked() {
		return nil, nil
	}
	r.Rules.SpectatorsAllowed = true
	r.bumpVersion()
	r.touch(now)
	meta, err := r.nextEventMeta(now)
	if err != nil {
		return nil, err
	}
	return []Event{RoomUnlockedEvent{Meta: meta}}, nil
}

func (r *Room) applyRemoveParticipant(cmd RemoveParticipantCommand, now time.Time) ([]Event, error) {
	if !r.isState(shared.RoomLobby, shared.RoomCountdown) {
		return nil, shared.DomainError{Code: shared.ErrRoomStateInvalid}
	}
	if err := r.requireHost(cmd.Meta.AuthenticatedParticipantID); err != nil {
		return nil, err
	}
	if cmd.Meta.AuthenticatedParticipantID == cmd.Participant {
		return nil, shared.DomainError{Code: shared.ErrActionNotAllowedForRole}
	}
	p := r.findActive(cmd.Participant)
	if p == nil {
		return nil, shared.DomainError{Code: shared.ErrParticipantNotFound}
	}
	p.RemovedAt = ptrTimestamp(mustTimestampFromTime(now))
	p.RemovedReason = "removed_by_host"
	p.IsReady = false
	r.updateParticipant(*p)
	r.bumpVersion()
	r.touch(now)

	meta, err := r.nextEventMeta(now)
	if err != nil {
		return nil, err
	}
	return []Event{ParticipantRemovedEvent{Meta: meta, ParticipantID: p.ID, Reason: p.RemovedReason}}, nil
}

func (r *Room) applyBlockParticipant(cmd BlockParticipantCommand, now time.Time) ([]Event, error) {
	if !r.isState(shared.RoomLobby, shared.RoomResults) {
		return nil, shared.DomainError{Code: shared.ErrRoomStateInvalid}
	}
	if err := r.requireHost(cmd.Meta.AuthenticatedParticipantID); err != nil {
		return nil, err
	}
	p := r.findActive(cmd.Participant)
	if p == nil {
		return nil, shared.DomainError{Code: shared.ErrParticipantNotFound}
	}
	p.RemovedAt = ptrTimestamp(mustTimestampFromTime(now))
	p.RemovedReason = "blocked_by_host"
	p.IsReady = false
	r.updateParticipant(*p)
	r.bumpVersion()
	r.touch(now)

	meta, err := r.nextEventMeta(now)
	if err != nil {
		return nil, err
	}
	return []Event{ParticipantBlockedEvent{Meta: meta, ParticipantID: p.ID}}, nil
}

func (r *Room) applyTransferHost(cmd TransferHostCommand, now time.Time) ([]Event, error) {
	if !r.isState(shared.RoomLobby, shared.RoomCountdown) {
		return nil, shared.DomainError{Code: shared.ErrRoomStateInvalid}
	}
	if err := r.requireHost(cmd.Meta.AuthenticatedParticipantID); err != nil {
		return nil, err
	}
	target := r.findActive(cmd.Participant)
	if target == nil {
		return nil, shared.DomainError{Code: shared.ErrParticipantNotFound}
	}
	if target.ID == *r.HostParticipantID {
		return nil, nil
	}
	oldHost := r.HostParticipantID
	if oldHost != nil {
		if h := r.findActive(*oldHost); h != nil {
			h.IsHost = false
			r.updateParticipant(*h)
		}
	}
	target.IsHost = true
	r.HostParticipantID = &target.ID
	r.updateParticipant(*target)
	r.bumpVersion()
	r.touch(now)

	meta, err := r.nextEventMeta(now)
	if err != nil {
		return nil, err
	}
	return []Event{HostTransferredEvent{Meta: meta, From: oldHost, To: target.ID}}, nil
}

func (r *Room) applyStartCountdown(cmd StartCountdownCommand, now time.Time) ([]Event, error) {
	if !r.isState(shared.RoomLobby) {
		return nil, shared.DomainError{Code: shared.ErrRoomStateInvalid}
	}
	if err := r.requireHost(cmd.Meta.AuthenticatedParticipantID); err != nil {
		return nil, err
	}
	if r.Countdown != nil {
		return nil, shared.DomainError{Code: shared.ErrCountdownAlreadyStarted}
	}
	if r.CurrentMatchID != nil {
		return nil, shared.DomainError{Code: shared.ErrMatchAlreadyExists}
	}

	cap := modeCapacity(r.Rules.Mode)
	players, ready := r.playerReadyCounts()
	if players < cap.MinPlayers {
		return nil, shared.DomainError{Code: shared.ErrInsufficientPlayers}
	}
	if ready != players {
		return nil, shared.DomainError{Code: shared.ErrPlayersNotReady}
	}

	deadline := mustTimestampFromTime(now.Add(countdownDuration))
	r.Countdown = &CountdownState{
		MatchID:    cmd.MatchID,
		Generation: 1,
		DeadlineAt: deadline,
		Rules:      r.Rules,
		Puzzle:     cmd.Puzzle,
	}
	r.State = shared.RoomCountdown
	r.bumpVersion()
	r.touch(now)

	meta, err := r.nextEventMeta(now)
	if err != nil {
		return nil, err
	}
	return []Event{CountdownStartedEvent{
		Meta:       meta,
		MatchID:    cmd.MatchID,
		Generation: 1,
		DeadlineAt: deadline,
		Rules:      r.Rules,
	}}, nil
}

func (r *Room) applyCancelCountdown(cmd CancelCountdownCommand, now time.Time) ([]Event, error) {
	if !r.isState(shared.RoomCountdown) {
		return nil, shared.DomainError{Code: shared.ErrCountdownNotActive}
	}
	if err := r.requireHost(cmd.Meta.AuthenticatedParticipantID); err != nil {
		return nil, err
	}
	r.Countdown = nil
	r.State = shared.RoomLobby
	for i := range r.Participants {
		if r.Participants[i].IsActive() && r.Participants[i].Role == shared.RolePlayer {
			r.Participants[i].IsReady = false
		}
	}
	r.bumpVersion()
	r.touch(now)

	meta1, err := r.nextEventMeta(now)
	if err != nil {
		return nil, err
	}
	meta2, err := r.nextEventMeta(now)
	if err != nil {
		return nil, err
	}
	return []Event{
		CountdownCancelledEvent{Meta: meta1},
		RoomReadyStatesResetEvent{Meta: meta2},
	}, nil
}

func (r *Room) applyActivateMatch(cmd ActivateMatchCommand, now time.Time) ([]Event, error) {
	if !r.isState(shared.RoomCountdown) {
		return nil, shared.DomainError{Code: shared.ErrRoomStateInvalid}
	}
	if r.Countdown == nil || r.Countdown.Generation != cmd.Generation {
		return nil, shared.DomainError{Code: shared.ErrTimerTokenStale}
	}
	r.State = shared.RoomInMatch
	r.CurrentMatchID = &r.Countdown.MatchID
	r.Countdown = nil
	r.bumpVersion()
	r.touch(now)

	meta, err := r.nextEventMeta(now)
	if err != nil {
		return nil, err
	}
	return []Event{RoomEnteredMatchEvent{Meta: meta, MatchID: *r.CurrentMatchID}}, nil
}

func (r *Room) applyExpireRoom(cmd ExpireRoomCommand, now time.Time) ([]Event, error) {
	if r.isState(shared.RoomExpired, shared.RoomCancelled, shared.RoomTerminatedByAdmin) {
		return nil, nil
	}
	r.State = shared.RoomExpired
	r.bumpVersion()
	r.touch(now)
	meta, err := r.nextEventMeta(now)
	if err != nil {
		return nil, err
	}
	return []Event{RoomExpiredEvent{Meta: meta}}, nil
}

// --- helpers ---

func (r *Room) requireHost(id shared.ParticipantID) error {
	if r.HostParticipantID == nil || *r.HostParticipantID != id {
		return shared.DomainError{Code: shared.ErrNotRoomHost}
	}
	return nil
}

func (r *Room) isLocked() bool {
	return !r.Rules.SpectatorsAllowed
}

func (r *Room) findActive(id shared.ParticipantID) *Participant {
	for i := range r.Participants {
		if r.Participants[i].ID == id && r.Participants[i].IsActive() {
			return &r.Participants[i]
		}
	}
	return nil
}

func (r *Room) findActiveByComparisonKey(key string) *Participant {
	for i := range r.Participants {
		if r.Participants[i].IsActive() && r.Participants[i].Name.ComparisonKey() == key {
			return &r.Participants[i]
		}
	}
	return nil
}

func (r *Room) updateParticipant(updated Participant) {
	for i := range r.Participants {
		if r.Participants[i].ID == updated.ID {
			r.Participants[i] = updated
			return
		}
	}
}

func (r *Room) counts() (players, spectators int) {
	for _, p := range r.Participants {
		if !p.IsActive() {
			continue
		}
		if p.Role == shared.RolePlayer {
			players++
		} else {
			spectators++
		}
	}
	return
}

func (r *Room) playerReadyCounts() (players, ready int) {
	for _, p := range r.Participants {
		if !p.IsActive() || p.Role != shared.RolePlayer {
			continue
		}
		players++
		if p.IsReady {
			ready++
		}
	}
	return
}

func (r *Room) longestActiveParticipant(exclude *shared.ParticipantID) *Participant {
	var oldest *Participant
	for i := range r.Participants {
		p := &r.Participants[i]
		if !p.IsActive() {
			continue
		}
		if exclude != nil && p.ID == *exclude {
			continue
		}
		if oldest == nil || p.JoinedAt.Milliseconds() < oldest.JoinedAt.Milliseconds() {
			oldest = p
		}
	}
	return oldest
}

func (r *Room) clearHost() {
	r.HostParticipantID = nil
	for i := range r.Participants {
		r.Participants[i].IsHost = false
	}
}

func (r *Room) bumpVersion() {
	r.Version++
}

func (r *Room) touch(now time.Time) {
	if ts, err := shared.NewTimestamp(now); err == nil {
		r.LastActivityAt = ts
	}
}

func (r *Room) nextEventMeta(now time.Time) (shared.EventMetadata, error) {
	num := uint64(len(r.Participants)) // placeholder; actor assigns durable event numbers
	num = 1
	return shared.NewEventMetadata(1, shared.EventNumber(num), uint64(r.Version), shared.NewRoomTarget(r.ID), mustTimestampFromTime(now))
}

func mustTimestampFromTime(now time.Time) shared.Timestamp {
	ts, err := shared.NewTimestamp(now)
	if err != nil {
		panic(err)
	}
	return ts
}

func ptrTimestamp(ts shared.Timestamp) *shared.Timestamp {
	return &ts
}

// Counts returns the active player and spectator counts for capacity checks.
func (r *Room) Counts() (players, spectators int) {
	return r.counts()
}
