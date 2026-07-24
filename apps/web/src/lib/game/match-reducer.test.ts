import { describe, expect, it } from 'vitest';

import type { ServerMessage } from '$lib/realtime/room-reducer';

import {
  applyMatchMessage,
  beginPending,
  displayedCell,
  initialMatchState,
  markPendingUncertain,
} from './match-reducer';
import type { MatchSnapshot, PendingCommand } from './types';

const snapshot: MatchSnapshot = {
  id: '01900000-0000-7000-8000-000000000010',
  state: 'Active',
  version: 2,
  startedAt: 1_000,
  penaltiesMs: 0,
  hintsUsed: 0,
  assisted: false,
  rules: {
    mode: 'Coop',
    difficulty: 'Easy',
    errorPreset: 'Casual',
    hintsEnabled: true,
    autoRemoveNotes: true,
    ruleVersion: 1,
  },
  cells: Array.from({ length: 81 }, (_, index) => ({
    index,
    isClue: index === 0,
    value: index === 0 ? 4 : undefined,
    notes: [],
  })),
  mistakes: {},
  contributions: {},
};

function message(
  type: ServerMessage['type'],
  eventNumber: number,
  aggregateVersion: number,
  payload: Record<string, unknown>,
): ServerMessage {
  return {
    schemaVersion: 1,
    eventNumber,
    aggregateVersion,
    serverTimestamp: 2_000,
    type,
    payload,
  } as ServerMessage;
}

function withSnapshot() {
  return applyMatchMessage(initialMatchState, message('match.snapshot', 2, 2, { match: snapshot }));
}

describe('match reducer', () => {
  it('applies ordered events and ignores duplicate delivery', () => {
    const state = withSnapshot();
    const placed = message('match.event', 3, 3, {
      event: {
        type: 'ValuePlaced',
        payload: {
          schemaVersion: 1,
          cell: 1,
          value: 8,
          participantId: '01900000-0000-7000-8000-000000000011',
          correct: true,
        },
      },
    });
    const next = applyMatchMessage(state, placed);
    expect(next.confirmed?.cells[1]?.value).toBe(8);
    expect(applyMatchMessage(next, placed)).toBe(next);
  });

  it('requires recovery for gaps and incompatible events', () => {
    const state = withSnapshot();
    expect(
      applyMatchMessage(
        state,
        message('match.event', 4, 3, {
          event: { type: 'ValueErased', payload: { schemaVersion: 1, cell: 1 } },
        }),
      ).recoveryRequired,
    ).toBe(true);
    expect(
      applyMatchMessage(
        state,
        message('match.event', 3, 3, {
          event: { type: 'FutureMutation', payload: { schemaVersion: 2 } },
        }),
      ).recoveryRequired,
    ).toBe(true);
  });

  it('shows reversible pending feedback and reconciles event-before-ack', () => {
    const pending: PendingCommand = {
      requestId: '01900000-0000-7000-8000-000000000099',
      kind: 'place',
      cell: 1,
      digit: 7,
      createdAt: 1,
      uncertain: false,
    };
    const optimistic = beginPending(withSnapshot(), pending);
    expect(displayedCell(optimistic, 1)?.value).toBe(7);
    expect(
      markPendingUncertain(optimistic, pending.requestId).pending.get(pending.requestId)?.uncertain,
    ).toBe(true);
    const confirmed = applyMatchMessage(
      optimistic,
      message('match.event', 3, 3, {
        event: {
          type: 'ValuePlaced',
          payload: { schemaVersion: 1, cell: 1, value: 7, correct: true },
        },
      }),
    );
    expect(confirmed.pending.size).toBe(0);
    expect(
      applyMatchMessage(
        confirmed,
        message('command.acknowledged', 0, 0, { requestId: pending.requestId }),
      ).confirmed?.cells[1]?.value,
    ).toBe(7);
  });

  it('starts local pending feedback within the response budget', () => {
    const started = performance.now();
    const state = beginPending(withSnapshot(), {
      requestId: '01900000-0000-7000-8000-000000000098',
      kind: 'place',
      cell: 1,
      digit: 6,
      createdAt: started,
      uncertain: false,
    });
    expect(displayedCell(state, 1)?.value).toBe(6);
    expect(performance.now() - started).toBeLessThan(50);
  });

  it('handles acknowledgement-before-event and unknown command status', () => {
    const requestId = '01900000-0000-7000-8000-000000000097';
    const pending = beginPending(withSnapshot(), {
      requestId,
      kind: 'place',
      cell: 1,
      digit: 5,
      createdAt: 1,
      uncertain: false,
    });
    const acknowledged = applyMatchMessage(
      pending,
      message('command.acknowledged', 0, 0, { requestId }),
    );
    expect(acknowledged.pending.size).toBe(1);
    const event = applyMatchMessage(
      acknowledged,
      message('match.event', 3, 3, {
        event: {
          type: 'ValuePlaced',
          payload: { schemaVersion: 1, cell: 1, value: 5, correct: true },
        },
      }),
    );
    expect(event.confirmed?.cells[1]?.value).toBe(5);

    const uncertain = beginPending(event, {
      requestId,
      kind: 'erase',
      cell: 1,
      createdAt: 2,
      uncertain: true,
    });
    expect(
      applyMatchMessage(
        uncertain,
        message('command.status', 0, 0, {
          requestId,
          status: 'COMMAND_OUTCOME_UNKNOWN',
        }),
      ).recoveryRequired,
    ).toBe(true);
  });

  it.each([
    ['NotesAdded', { cell: 1, digits: [2, 3], participantId: 'p' }, [2, 3]],
    ['NotesRemoved', { cell: 1, digits: [2], participantId: 'p' }, []],
    ['NotesAutoRemoved', { cell: 1, digits: [2], causedBy: 2 }, []],
  ])('reduces %s', (eventType, payload, expected) => {
    const base = withSnapshot();
    if (eventType !== 'NotesAdded') base.confirmed!.cells[1]!.notes = [2];
    const next = applyMatchMessage(
      base,
      message('match.event', 3, 3, {
        event: { type: eventType, payload: { schemaVersion: 1, ...payload } },
      }),
    );
    expect(next.confirmed?.cells[1]?.notes).toEqual(expected);
  });

  it('tracks penalties, hints, erase, and completion', () => {
    let state = withSnapshot();
    const events = [
      ['ValueRejected', { cell: 1, participantId: 'p', penaltyMs: 5_000 }],
      ['HintUsed', { level: 'Nudge', targetCell: 1, participantId: 'p' }],
      ['ValueErased', { cell: 1, participantId: 'p' }],
      ['MatchCompleted', { assisted: true }],
    ] as const;
    events.forEach(([type, payload], offset) => {
      state = applyMatchMessage(
        state,
        message('match.event', 3 + offset, 3 + offset, {
          event: { type, payload: { schemaVersion: 1, ...payload } },
        }),
      );
    });
    expect(state.confirmed?.penaltiesMs).toBe(5_000);
    expect(state.confirmed?.hintsUsed).toBe(1);
    expect(state.confirmed?.state).toBe('Completed');
  });

  it.each(['MatchStarted', 'Ping', 'ParticipantDisconnected', 'ParticipantReconnected'])(
    'accepts known compatible %s events',
    (eventType) => {
      const next = applyMatchMessage(
        withSnapshot(),
        message('match.event', 3, 3, {
          event: {
            type: eventType,
            payload: {
              schemaVersion: 1,
              cell: 1,
              intent: 'look_here',
              participantId: '01900000-0000-7000-8000-000000000011',
            },
          },
        }),
      );
      expect(next.recoveryRequired).toBe(false);
      expect(next.lastEventNumber).toBe(3);
    },
  );
});
