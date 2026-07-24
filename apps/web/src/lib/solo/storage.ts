import Dexie, { type EntityTable } from 'dexie';

import type { Digit } from '$lib/game/types';

export type SoloPlayStyle = 'Guided' | 'Classic';
export type SoloDifficulty = 'Easy' | 'Medium' | 'Hard' | 'Expert';

export interface SoloReplayEvent {
  atMs: number;
  type: 'value' | 'erase' | 'note' | 'hint' | 'pause' | 'resume' | 'complete';
  cell?: number;
  value?: number;
  detail?: string;
}

export interface SoloAttempt {
  id: string;
  puzzleId: string;
  revision: number;
  difficulty: SoloDifficulty;
  playStyle: SoloPlayStyle;
  clues: string;
  assignmentProof: string;
  values: number[];
  notes: Digit[][];
  elapsedMs: number;
  penaltyMs: number;
  started: boolean;
  paused: boolean;
  hintsUsed: number;
  mistakes: number;
  updatedAtMs: number;
  replay: SoloReplayEvent[];
}

export interface SoloResult {
  id: string;
  puzzleId: string;
  difficulty: SoloDifficulty;
  playStyle: SoloPlayStyle;
  elapsedMs: number;
  penaltyMs: number;
  hintsUsed: number;
  mistakes: number;
  completedAtMs: number;
}

export interface SoloReplay {
  id: string;
  clues: string;
  difficulty: SoloDifficulty;
  playStyle: SoloPlayStyle;
  events: SoloReplayEvent[];
}

export interface SoloPreference {
  key: string;
  value: string;
}

export interface RecentPuzzle {
  puzzleId: string;
  encounteredAtMs: number;
}

export class SoloDatabase extends Dexie {
  soloAttempts!: EntityTable<SoloAttempt, 'id'>;
  soloReplays!: EntityTable<SoloReplay, 'id'>;
  localStatistics!: EntityTable<SoloResult, 'id'>;
  recentPuzzles!: EntityTable<RecentPuzzle, 'puzzleId'>;
  preferences!: EntityTable<SoloPreference, 'key'>;

  constructor(name = 'ninefold') {
    super(name);
    this.version(1).stores({
      soloAttempts: 'id, updatedAtMs',
      soloReplays: 'id',
      localStatistics: 'id, difficulty, completedAtMs',
      recentPuzzles: 'puzzleId, encounteredAtMs',
      preferences: 'key',
    });
    this.version(2).stores({
      soloAttempts: 'id, updatedAtMs',
      soloReplays: 'id',
      localStatistics: 'id, difficulty, completedAtMs',
      recentPuzzles: 'puzzleId, encounteredAtMs',
      preferences: 'key',
    });
  }
}

let database: SoloDatabase | undefined;

export function soloDatabase(): SoloDatabase {
  database ??= new SoloDatabase();
  return database;
}

export async function currentAttempt(db = soloDatabase()): Promise<SoloAttempt | undefined> {
  const attempt = await db.soloAttempts.orderBy('updatedAtMs').last();
  if (!attempt) return undefined;
  if (!validAttempt(attempt)) {
    await db.soloAttempts.delete(attempt.id);
    return undefined;
  }
  return attempt;
}

export async function saveAttempt(attempt: SoloAttempt, db = soloDatabase()): Promise<void> {
  await db.transaction('rw', db.soloAttempts, db.recentPuzzles, async () => {
    await db.soloAttempts.clear();
    await db.soloAttempts.put(structuredClone(attempt));
    await db.recentPuzzles.put({
      puzzleId: attempt.puzzleId,
      encounteredAtMs: attempt.updatedAtMs,
    });
    const old = await db.recentPuzzles.orderBy('encounteredAtMs').reverse().offset(50).toArray();
    await db.recentPuzzles.bulkDelete(old.map(({ puzzleId }) => puzzleId));
  });
}

export async function completeAttempt(
  attempt: SoloAttempt,
  result: SoloResult,
  db = soloDatabase(),
): Promise<void> {
  await db.transaction('rw', db.soloAttempts, db.soloReplays, db.localStatistics, async () => {
    await db.soloReplays.put({
      id: result.id,
      clues: attempt.clues,
      difficulty: attempt.difficulty,
      playStyle: attempt.playStyle,
      events: structuredClone(attempt.replay),
    });
    await db.localStatistics.put(result);
    await db.soloAttempts.delete(attempt.id);
  });
}

export async function recentPuzzleIds(db = soloDatabase()): Promise<string[]> {
  return (await db.recentPuzzles.orderBy('encounteredAtMs').reverse().limit(50).toArray()).map(
    ({ puzzleId }) => puzzleId,
  );
}

export async function soloResults(db = soloDatabase()): Promise<SoloResult[]> {
  return db.localStatistics.orderBy('completedAtMs').reverse().toArray();
}

export async function soloPreference(
  key: 'soloPlayStyle' | 'inputPreference',
  db = soloDatabase(),
): Promise<string | undefined> {
  return (await db.preferences.get(key))?.value;
}

export async function saveSoloPreference(
  key: 'soloPlayStyle' | 'inputPreference',
  value: string,
  db = soloDatabase(),
): Promise<void> {
  await db.preferences.put({ key, value });
}

export async function clearSoloData(db = soloDatabase()): Promise<void> {
  await db.transaction(
    'rw',
    [db.soloAttempts, db.soloReplays, db.localStatistics, db.recentPuzzles, db.preferences],
    async () => {
      await Promise.all([
        db.soloAttempts.clear(),
        db.soloReplays.clear(),
        db.localStatistics.clear(),
        db.recentPuzzles.clear(),
        db.preferences.where('key').anyOf(['soloPlayStyle', 'inputPreference']).delete(),
      ]);
    },
  );
}

export function cellsToValues(clues: string, cells: { value?: number }[]): string {
  return [...clues]
    .map((clue, index) => (clue !== '0' ? clue : String(cells[index]?.value ?? 0)))
    .join('');
}

export function suspendedElapsed(
  elapsedMs: number,
  runningSince: number | null,
  now: number,
): number {
  return elapsedMs + (runningSince === null ? 0 : Math.max(0, now - runningSince));
}

function validAttempt(value: SoloAttempt): boolean {
  return (
    typeof value.id === 'string' &&
    typeof value.puzzleId === 'string' &&
    Number.isInteger(value.revision) &&
    value.revision > 0 &&
    /^[0-9]{81}$/.test(value.clues) &&
    typeof value.assignmentProof === 'string' &&
    value.assignmentProof.length >= 16 &&
    Array.isArray(value.values) &&
    value.values.length === 81 &&
    value.values.every((digit) => Number.isInteger(digit) && digit >= 0 && digit <= 9) &&
    Array.isArray(value.notes) &&
    value.notes.length === 81 &&
    Number.isFinite(value.elapsedMs) &&
    value.elapsedMs >= 0 &&
    Number.isFinite(value.penaltyMs) &&
    value.penaltyMs >= 0
  );
}
