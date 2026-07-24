import type { Digit, MatchCell } from '$lib/game/types';

export interface ReplayEvent {
  proofVersion: 1;
  eventNumber: number;
  aggregateVersion: number;
  publicEventType: string;
  publicActorId: string;
  occurredAtMs: number;
  publicPayload: Record<string, unknown>;
  privatePayloadDigest: string;
  previousEventHash: string;
  eventHash: string;
}

export interface ReplayDocument {
  schemaVersion: number;
  replayId: string;
  matchId: string;
  expiresAt: number;
  clues: string;
  rules: {
    mode: 'Coop';
    difficulty: 'Easy' | 'Medium' | 'Hard' | 'Expert';
    errorPreset: 'Casual' | 'Challenge' | 'Blind' | 'Clean';
    hintsEnabled: boolean;
    autoRemoveNotes: boolean;
    ruleVersion: number;
  };
  participants: { id: string; name: string }[];
  events: ReplayEvent[];
  proof: {
    proofVersion: 1;
    matchId: string;
    finalEventNumber: number;
    finalEventHash: string;
    terminalAtMs: number;
    keyId: string;
    signature: string;
  };
}

export interface ReplayState {
  eventIndex: number;
  cells: MatchCell[];
  completed: boolean;
  pingedCell: number | null;
  connections: Record<string, boolean>;
}

const supportedEvents = new Set([
  'MatchPrepared',
  'MatchCountdownStarted',
  'MatchStarted',
  'ValuePlaced',
  'ValueRejected',
  'ValueErased',
  'NotesAdded',
  'NotesRemoved',
  'NotesAutoRemoved',
  'HintUsed',
  'Ping',
  'ParticipantDisconnected',
  'ParticipantReconnected',
  'MatchEnteredRecovery',
  'MatchRecovered',
  'MatchCancelled',
  'MatchCompleted',
]);

export function validateReplay(document: ReplayDocument): void {
  if (document.schemaVersion !== 1) throw new Error('Unsupported replay schema');
  if (!/^[0-9]{81}$/.test(document.clues)) throw new Error('Invalid replay clues');
  document.events.forEach((event, index) => {
    if (event.eventNumber !== index + 1) throw new Error(`Replay event gap at ${index + 1}`);
    if (event.publicPayload.schemaVersion !== 1) throw new Error('Unsupported event schema');
    if (!supportedEvents.has(event.publicEventType))
      throw new Error(`Unsupported replay event: ${event.publicEventType}`);
  });
}

export function replayStateAt(document: ReplayDocument, eventIndex: number): ReplayState {
  validateReplay(document);
  const boundary = Math.max(0, Math.min(eventIndex, document.events.length));
  const connections = Object.fromEntries(document.participants.map(({ id }) => [id, true]));
  const state: ReplayState = {
    eventIndex: 0,
    cells: [...document.clues].map((encoded, index) => {
      const value = Number(encoded);
      return {
        index,
        isClue: value !== 0,
        value: value === 0 ? undefined : (value as Digit),
        notes: [],
      };
    }),
    completed: false,
    pingedCell: null,
    connections,
  };
  for (let index = 0; index < boundary; index++) applyEvent(state, document.events[index]!);
  state.eventIndex = boundary;
  return state;
}

function applyEvent(state: ReplayState, event: ReplayEvent): void {
  const payload = event.publicPayload;
  const cellIndex = typeof payload.cell === 'number' ? payload.cell : -1;
  const cell = state.cells[cellIndex];
  switch (event.publicEventType) {
    case 'ValuePlaced':
      if (!cell || typeof payload.value !== 'number') throw new Error('Invalid value event');
      cell.value = payload.value as Digit;
      cell.attribution =
        typeof payload.participantId === 'string' ? payload.participantId : undefined;
      cell.correct = typeof payload.correct === 'boolean' ? payload.correct : null;
      break;
    case 'ValueErased':
      if (!cell) throw new Error('Invalid erase event');
      delete cell.value;
      delete cell.attribution;
      cell.correct = null;
      break;
    case 'NotesAdded':
      updateNotes(cell, payload.digits, true);
      break;
    case 'NotesRemoved':
    case 'NotesAutoRemoved':
      updateNotes(cell, payload.digits, false);
      break;
    case 'Ping':
      state.pingedCell = cell ? cellIndex : null;
      break;
    case 'ParticipantDisconnected':
      if (typeof payload.participantId === 'string')
        state.connections[payload.participantId] = false;
      break;
    case 'ParticipantReconnected':
      if (typeof payload.participantId === 'string')
        state.connections[payload.participantId] = true;
      break;
    case 'MatchCompleted':
      state.completed = true;
      break;
  }
}

function updateNotes(cell: MatchCell | undefined, value: unknown, add: boolean): void {
  if (!cell || !Array.isArray(value)) throw new Error('Invalid notes event');
  const digits = value as Digit[];
  cell.notes = add
    ? [...new Set([...cell.notes, ...digits])].sort()
    : cell.notes.filter((digit) => !digits.includes(digit));
}

export function playbackOrder(document: ReplayDocument, speed: 0.5 | 1 | 2 | 4): number[] {
  if (![0.5, 1, 2, 4].includes(speed)) throw new Error('Unsupported playback speed');
  validateReplay(document);
  return document.events.map((event) => event.eventNumber);
}
