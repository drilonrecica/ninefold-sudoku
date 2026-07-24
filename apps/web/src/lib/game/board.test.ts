import { describe, expect, it } from 'vitest';

import { boxOf, columnOf, directConflictIndices, isRelated, moveIndex, rowOf } from './board';
import type { MatchCell } from './types';

describe('board helpers', () => {
  it('calculates row, column, box, and related cells', () => {
    expect(rowOf(20)).toBe(2);
    expect(columnOf(20)).toBe(2);
    expect(boxOf(20)).toBe(0);
    expect(isRelated(0, 8)).toBe(true);
    expect(isRelated(0, 10)).toBe(true);
    expect(isRelated(0, 40)).toBe(false);
  });

  it('wraps roving navigation predictably', () => {
    expect(moveIndex(0, 'ArrowLeft')).toBe(8);
    expect(moveIndex(0, 'ArrowUp')).toBe(72);
    expect(moveIndex(13, 'Home')).toBe(9);
    expect(moveIndex(13, 'End')).toBe(17);
  });

  it('marks both sides of a direct conflict without solution knowledge', () => {
    const cells: MatchCell[] = Array.from({ length: 81 }, (_, index) => ({
      index,
      isClue: false,
      notes: [],
    }));
    cells[0]!.value = 4;
    cells[8]!.value = 4;
    expect([...directConflictIndices(cells)]).toEqual([0, 8]);
  });
});
