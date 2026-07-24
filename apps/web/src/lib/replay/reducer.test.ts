import { describe, expect, it } from 'vitest';

import { playbackOrder, replayStateAt, type ReplayDocument, validateReplay } from './reducer';

const replay: ReplayDocument = {
  schemaVersion: 1,
  replayId: '018f0000-0000-7000-8000-000000000001',
  matchId: '018f0000-0000-7000-8000-000000000002',
  expiresAt: 2_000_000_000_000,
  clues: '0'.repeat(81),
  rules: {
    mode: 'Coop',
    difficulty: 'Easy',
    errorPreset: 'Casual',
    hintsEnabled: true,
    autoRemoveNotes: true,
    ruleVersion: 1,
  },
  participants: [{ id: 'p1', name: 'Player' }],
  events: [
    {
      eventNumber: 1,
      aggregateVersion: 1,
      serverTimestamp: 1,
      type: 'MatchPrepared',
      payload: { schemaVersion: 1 },
    },
    {
      eventNumber: 2,
      aggregateVersion: 2,
      serverTimestamp: 2,
      type: 'ValuePlaced',
      payload: { schemaVersion: 1, cell: 0, value: 4, participantId: 'p1' },
    },
    {
      eventNumber: 3,
      aggregateVersion: 3,
      serverTimestamp: 3,
      type: 'NotesAdded',
      payload: { schemaVersion: 1, cell: 1, digits: [2, 3], participantId: 'p1' },
    },
  ],
};

describe('replay reducer', () => {
  it('reconstructs every boundary and seeks backward deterministically', () => {
    expect(replayStateAt(replay, 0).cells[0]?.value).toBeUndefined();
    expect(replayStateAt(replay, 2).cells[0]?.value).toBe(4);
    expect(replayStateAt(replay, 3).cells[1]?.notes).toEqual([2, 3]);
    expect(replayStateAt(replay, 1).cells[0]?.value).toBeUndefined();
    expect(replayStateAt(replay, 3)).toEqual(replayStateAt(replay, 3));
  });

  it('keeps event order independent of playback speed', () => {
    for (const speed of [0.5, 1, 2, 4] as const)
      expect(playbackOrder(replay, speed)).toEqual([1, 2, 3]);
  });

  it('rejects gaps and incompatible schemas or events', () => {
    expect(() =>
      validateReplay({ ...replay, events: [{ ...replay.events[0]!, eventNumber: 2 }] }),
    ).toThrow(/gap/);
    expect(() => validateReplay({ ...replay, schemaVersion: 2 })).toThrow(/schema/);
    expect(() =>
      validateReplay({
        ...replay,
        events: [{ ...replay.events[0]!, type: 'FutureEvent' }],
      }),
    ).toThrow(/Unsupported replay event/);
  });
});
