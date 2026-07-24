import { describe, expect, it } from 'vitest';
import type { ServerMessage } from '../../../../../contracts/generated/typescript/realtime/server';

import { applyRoomMessage, initialRoomState } from './room-reducer';

const snapshot = {
  id: '01900000-0000-7000-8000-000000000001',
  code: '7KMP4R',
  state: 'Lobby',
  version: 2,
  settings: {
    mode: 'Coop',
    difficulty: 'Easy',
    errorPreset: 'Casual',
    hintsEnabled: true,
    sharedNotes: true,
    autoRemoveNotes: true,
    spectatorsAllowed: true,
  },
  hostId: '01900000-0000-7000-8000-000000000002',
  currentMatchId: null,
  participants: [],
} as const;

describe('room reducer', () => {
  it('applies authoritative snapshots and ignores older delivery', () => {
    const message = {
      type: 'room.snapshot',
      payload: { room: snapshot },
    } as unknown as ServerMessage;
    const current = applyRoomMessage(initialRoomState, message);
    expect(current.room?.version).toBe(2);
    const older = {
      type: 'room.snapshot',
      payload: { room: { ...snapshot, version: 1 } },
    } as unknown as ServerMessage;
    expect(applyRoomMessage(current, older)).toBe(current);
  });

  it('treats revoked controllers as read only', () => {
    const next = applyRoomMessage(initialRoomState, {
      type: 'connection.controller_revoked',
      payload: {},
    } as ServerMessage);
    expect(next.connection).toBe('read_only');
    expect(next.isController).toBe(false);
  });
});
