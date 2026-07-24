import 'fake-indexeddb/auto';

import Dexie from 'dexie';
import { afterEach, describe, expect, it } from 'vitest';

import {
  clearSoloData,
  completeAttempt,
  currentAttempt,
  recentPuzzleIds,
  saveAttempt,
  SoloDatabase,
  suspendedElapsed,
  type SoloAttempt,
} from './storage';

const databases: SoloDatabase[] = [];

function db(): SoloDatabase {
  const value = new SoloDatabase(`ninefold-test-${crypto.randomUUID()}`);
  databases.push(value);
  return value;
}

function attempt(id: string = crypto.randomUUID()): SoloAttempt {
  return {
    id,
    puzzleId: crypto.randomUUID(),
    revision: 1,
    difficulty: 'Easy',
    playStyle: 'Guided',
    clues: '0'.repeat(81),
    assignmentProof: 'test-proof-value-long-enough',
    values: Array(81).fill(0),
    notes: Array.from({ length: 81 }, () => []),
    elapsedMs: 1000,
    penaltyMs: 0,
    started: true,
    paused: false,
    hintsUsed: 0,
    mistakes: 0,
    updatedAtMs: Date.now(),
    replay: [],
  };
}

afterEach(async () => {
  await Promise.all(databases.splice(0).map((value) => value.delete()));
});

describe('Solo IndexedDB storage', () => {
  it('keeps exactly one resumable attempt across database reopen', async () => {
    const database = db();
    const first = attempt('first');
    const second = { ...attempt('second'), updatedAtMs: first.updatedAtMs + 1 };
    await saveAttempt(first, database);
    await saveAttempt(second, database);
    database.close();
    const reopened = new SoloDatabase(database.name);
    databases.push(reopened);
    await expect(currentAttempt(reopened)).resolves.toMatchObject({
      id: 'second',
      elapsedMs: 1000,
    });
  });

  it('migrates a version-one database without losing an active attempt', async () => {
    const name = `ninefold-v1-${crypto.randomUUID()}`;
    const legacy = new Dexie(name);
    legacy.version(1).stores({
      soloAttempts: 'id, updatedAtMs',
      soloReplays: 'id',
      localStatistics: 'id, difficulty, completedAtMs',
      recentPuzzles: 'puzzleId, encounteredAtMs',
      preferences: 'key',
    });
    await legacy.table('soloAttempts').put(attempt('legacy'));
    legacy.close();
    const migrated = new SoloDatabase(name);
    databases.push(migrated);
    await expect(currentAttempt(migrated)).resolves.toMatchObject({ id: 'legacy' });
    expect(migrated.verno).toBe(2);
  });

  it('stores local result/replay and clears active progress', async () => {
    const database = db();
    const value = attempt('complete');
    value.replay.push({ atMs: 1000, type: 'value', cell: 0, value: 1 });
    await saveAttempt(value, database);
    await completeAttempt(
      value,
      {
        id: value.id,
        puzzleId: value.puzzleId,
        difficulty: value.difficulty,
        playStyle: value.playStyle,
        elapsedMs: 1000,
        penaltyMs: 0,
        hintsUsed: 0,
        mistakes: 0,
        completedAtMs: Date.now(),
      },
      database,
    );
    await expect(currentAttempt(database)).resolves.toBeUndefined();
    await expect(database.soloReplays.get(value.id)).resolves.toMatchObject({
      events: [{ type: 'value' }],
    });
  });

  it('bounds recent puzzles and clears all Solo data', async () => {
    const database = db();
    for (let index = 0; index < 55; index++) {
      const value = attempt(String(index));
      value.puzzleId = `puzzle-${index}`;
      value.updatedAtMs = index;
      await saveAttempt(value, database);
    }
    expect(await recentPuzzleIds(database)).toHaveLength(50);
    await clearSoloData(database);
    await expect(recentPuzzleIds(database)).resolves.toEqual([]);
    await expect(currentAttempt(database)).resolves.toBeUndefined();
  });

  it('discards corrupt attempts and surfaces unavailable storage', async () => {
    const database = db();
    await database.soloAttempts.put({ ...attempt('corrupt'), values: [12] });
    await expect(currentAttempt(database)).resolves.toBeUndefined();
    database.close({ disableAutoOpen: true });
    await expect(saveAttempt(attempt('unavailable'), database)).rejects.toThrow();
  });

  it('excludes paused and closed-tab intervals from elapsed time', () => {
    expect(suspendedElapsed(1000, 5000, 8000)).toBe(4000);
    expect(suspendedElapsed(4000, null, 20_000)).toBe(4000);
    expect(suspendedElapsed(4000, 30_000, 32_000)).toBe(6000);
  });
});
