import type { ServerMessage } from '$lib/realtime/room-reducer';

import type {
  Digit,
  MatchCell,
  MatchClientState,
  MatchEventEnvelope,
  MatchSnapshot,
  PendingCommand,
} from './types';

export const initialMatchState: MatchClientState = {
  confirmed: null,
  lastEventNumber: 0,
  pending: new Map(),
  recoveryRequired: false,
  announcement: '',
};

function cloneSnapshot(snapshot: MatchSnapshot): MatchSnapshot {
  return {
    ...snapshot,
    pausedMs: snapshot.pausedMs ?? 0,
    cells: snapshot.cells.map((cell) => ({ ...cell, notes: [...(cell.notes ?? [])] })),
    mistakes: { ...snapshot.mistakes },
    contributions: { ...snapshot.contributions },
  };
}

function reconcilePending(
  pending: Map<string, PendingCommand>,
  kind: PendingCommand['kind'],
  cell?: number,
): Map<string, PendingCommand> {
  const next = new Map(pending);
  const command = [...next.values()].find(
    (candidate) => candidate.kind === kind && (cell === undefined || candidate.cell === cell),
  );
  if (command) next.delete(command.requestId);
  return next;
}

function eventAnnouncement(type: string, payload: Record<string, unknown>): string {
  const cell = typeof payload.cell === 'number' ? payload.cell + 1 : null;
  switch (type) {
    case 'ValuePlaced':
      return cell ? `Value accepted in cell ${cell}.` : 'Value accepted.';
    case 'ValueRejected':
      return cell ? `Value rejected in cell ${cell}.` : 'Value rejected.';
    case 'ValueErased':
      return cell ? `Cell ${cell} erased.` : 'Value erased.';
    case 'NotesAdded':
    case 'NotesRemoved':
    case 'NotesAutoRemoved':
      return cell ? `Notes updated in cell ${cell}.` : 'Notes updated.';
    case 'HintUsed':
      return 'Hint used.';
    case 'MatchCompleted':
      return 'Puzzle completed.';
    default:
      return '';
  }
}

function applyMatchEvent(
  snapshot: MatchSnapshot,
  envelope: MatchEventEnvelope,
): MatchSnapshot | null {
  const { type, payload } = envelope.event;
  if (payload.schemaVersion !== 1) return null;
  const next = cloneSnapshot(snapshot);
  next.version = envelope.aggregateVersion;
  const cellIndex = typeof payload.cell === 'number' ? payload.cell : -1;
  const cell = cellIndex >= 0 && cellIndex < 81 ? next.cells[cellIndex] : undefined;

  switch (type) {
    case 'MatchPrepared':
    case 'MatchCountdownStarted':
      return next;
    case 'MatchStarted':
      next.state = 'Active';
      return next;
    case 'ValuePlaced':
      if (!cell || typeof payload.value !== 'number') return null;
      cell.value = payload.value as Digit;
      cell.attribution =
        typeof payload.participantId === 'string' ? payload.participantId : undefined;
      cell.correct = typeof payload.correct === 'boolean' ? payload.correct : null;
      return next;
    case 'ValueRejected': {
      const participant = String(payload.participantId ?? '');
      if (participant) next.mistakes[participant] = (next.mistakes[participant] ?? 0) + 1;
      next.penaltiesMs += typeof payload.penaltyMs === 'number' ? payload.penaltyMs : 0;
      return next;
    }
    case 'ValueErased':
      if (!cell) return null;
      delete cell.value;
      delete cell.attribution;
      cell.correct = null;
      return next;
    case 'NotesAdded':
      if (!cell || !Array.isArray(payload.digits)) return null;
      cell.notes = [...new Set([...cell.notes, ...(payload.digits as Digit[])])].sort();
      return next;
    case 'NotesRemoved':
    case 'NotesAutoRemoved':
      if (!cell || !Array.isArray(payload.digits)) return null;
      cell.notes = cell.notes.filter((digit) => !(payload.digits as Digit[]).includes(digit));
      return next;
    case 'HintUsed':
      next.hintsUsed += 1;
      next.assisted = true;
      return next;
    case 'Ping':
    case 'ParticipantDisconnected':
    case 'ParticipantReconnected':
      return next;
    case 'MatchEnteredRecovery':
      next.state = 'RecoveryPending';
      next.recoveryGeneration =
        typeof payload.generation === 'number' ? payload.generation : undefined;
      next.recoveryStartedAt =
        typeof payload.startedAtMs === 'number' ? payload.startedAtMs : undefined;
      return next;
    case 'MatchRecovered':
      next.state = 'Active';
      next.pausedMs =
        (next.pausedMs ?? 0) +
        (typeof payload.pausedIntervalMs === 'number' ? payload.pausedIntervalMs : 0);
      delete next.recoveryStartedAt;
      return next;
    case 'MatchCancelled':
      next.state = 'Cancelled';
      delete next.recoveryStartedAt;
      return next;
    case 'MatchCompleted':
      next.state = 'Completed';
      next.completedAt = Date.now();
      next.assisted = Boolean(payload.assisted);
      return next;
    default:
      return null;
  }
}

export function beginPending(state: MatchClientState, command: PendingCommand): MatchClientState {
  const pending = new Map(state.pending);
  pending.set(command.requestId, command);
  return { ...state, pending };
}

export function markPendingUncertain(state: MatchClientState, requestId: string): MatchClientState {
  const command = state.pending.get(requestId);
  if (!command) return state;
  const pending = new Map(state.pending);
  pending.set(requestId, { ...command, uncertain: true });
  return { ...state, pending };
}

export function applyMatchMessage(
  state: MatchClientState,
  message: ServerMessage,
): MatchClientState {
  if (message.schemaVersion !== 1) return { ...state, recoveryRequired: true };
  if (message.type === 'match.snapshot') {
    const snapshot = message.payload.match as unknown as MatchSnapshot | undefined;
    if (!snapshot || snapshot.cells?.length !== 81) return { ...state, recoveryRequired: true };
    if (state.confirmed && snapshot.version < state.confirmed.version) return state;
    return {
      ...state,
      confirmed: cloneSnapshot(snapshot),
      lastEventNumber: message.eventNumber,
      pending: new Map(),
      recoveryRequired: false,
    };
  }
  if (message.type === 'match.event') {
    if (!state.confirmed) return { ...state, recoveryRequired: true };
    if (message.eventNumber <= state.lastEventNumber) return state;
    if (message.eventNumber !== state.lastEventNumber + 1) {
      return { ...state, recoveryRequired: true };
    }
    const event = message.payload.event as unknown as MatchEventEnvelope['event'];
    if (!event) return { ...state, recoveryRequired: true };
    const envelope: MatchEventEnvelope = {
      schemaVersion: message.schemaVersion,
      eventNumber: message.eventNumber,
      aggregateVersion: message.aggregateVersion,
      event,
    };
    const confirmed = applyMatchEvent(state.confirmed, envelope);
    if (!confirmed) return { ...state, recoveryRequired: true };
    const cell = typeof event.payload.cell === 'number' ? event.payload.cell : undefined;
    const pending =
      event.type === 'ValuePlaced'
        ? reconcilePending(state.pending, 'place', cell)
        : event.type === 'ValueRejected'
          ? reconcilePending(state.pending, 'place', cell)
          : event.type === 'ValueErased'
            ? reconcilePending(state.pending, 'erase', cell)
            : event.type === 'NotesAdded'
              ? reconcilePending(state.pending, 'add_note', cell)
              : event.type === 'NotesRemoved'
                ? reconcilePending(state.pending, 'remove_note', cell)
                : event.type === 'HintUsed'
                  ? reconcilePending(state.pending, 'hint')
                  : state.pending;
    return {
      ...state,
      confirmed,
      pending,
      lastEventNumber: message.eventNumber,
      announcement: eventAnnouncement(event.type, event.payload),
    };
  }
  if (message.type === 'command.acknowledged' || message.type === 'command.rejected') {
    const requestId = message.payload.requestId;
    if (!requestId || !state.pending.has(requestId)) return state;
    if (message.type === 'command.acknowledged') return state;
    const pending = new Map(state.pending);
    pending.delete(requestId);
    return {
      ...state,
      pending,
      announcement: `Move rejected: ${message.payload.code ?? 'UNKNOWN'}.`,
      recoveryRequired: message.payload.code === 'STALE_VERSION' ? true : state.recoveryRequired,
    };
  }
  if (message.type === 'command.status') {
    const requestId = message.payload.requestId;
    if (!requestId || !state.pending.has(requestId)) return state;
    if (message.payload.status === 'ok') {
      return { ...state, recoveryRequired: true };
    }
    if (message.payload.status === 'COMMAND_OUTCOME_UNKNOWN') {
      return { ...state, recoveryRequired: true };
    }
  }
  return state;
}

export function displayedCell(state: MatchClientState, index: number): MatchCell | null {
  const confirmed = state.confirmed?.cells[index];
  if (!confirmed) return null;
  const display = { ...confirmed, notes: [...confirmed.notes] };
  for (const pending of state.pending.values()) {
    if (pending.cell !== index) continue;
    if (pending.kind === 'place' && pending.digit) display.value = pending.digit;
    if (pending.kind === 'erase') delete display.value;
    if (pending.kind === 'add_note' && pending.digit && !display.notes.includes(pending.digit)) {
      display.notes.push(pending.digit);
    }
    if (pending.kind === 'remove_note' && pending.digit) {
      display.notes = display.notes.filter((digit) => digit !== pending.digit);
    }
  }
  return display;
}
