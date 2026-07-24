import { fireEvent, render, screen } from '@testing-library/svelte';
import { describe, expect, it, vi } from 'vitest';

import type { MatchCell } from '$lib/game/types';

import SudokuBoard from './SudokuBoard.svelte';

const cells: MatchCell[] = Array.from({ length: 81 }, (_, index) => ({
  index,
  isClue: index === 0,
  value: index === 0 ? 4 : undefined,
  notes: index === 1 ? [2, 7] : [],
}));

function renderBoard(selectedIndex = 0) {
  const onselect = vi.fn();
  const onvalue = vi.fn();
  const onerase = vi.fn();
  const ontogglenotes = vi.fn();
  const onhint = vi.fn();
  render(SudokuBoard, {
    cells,
    selectedIndex,
    pendingCells: new Set<number>(),
    softLocks: new Map<number, string>(),
    participants: new Map(),
    onselect,
    onvalue,
    onerase,
    ontogglenotes,
    onhint,
  });
  return { onselect, onvalue, onerase, ontogglenotes, onhint };
}

describe('SudokuBoard', () => {
  it('renders 81 semantic cells with useful screen-reader labels', () => {
    renderBoard();
    const gridCells = screen.getAllByRole('gridcell');
    expect(gridCells).toHaveLength(81);
    expect(gridCells[0]?.getAttribute('aria-label')).toContain('Row 1, column 1, clue, value 4');
    expect(gridCells[1]?.getAttribute('aria-label')).toContain('shared notes 2, 7');
    expect(gridCells.filter((cell) => cell.getAttribute('tabindex') === '0')).toHaveLength(1);
  });

  it('supports arrows, digits, erase, notes, hints, and escape', async () => {
    const handlers = renderBoard(10);
    const selected = screen.getAllByRole('gridcell')[10]!;
    selected.focus();

    await fireEvent.keyDown(selected, { key: 'ArrowRight' });
    expect(handlers.onselect).toHaveBeenCalledWith(11);
    await fireEvent.keyDown(selected, { key: '8' });
    expect(handlers.onvalue).toHaveBeenCalledWith(8);
    await fireEvent.keyDown(selected, { key: 'Delete' });
    expect(handlers.onerase).toHaveBeenCalled();
    await fireEvent.keyDown(selected, { key: 'n' });
    expect(handlers.ontogglenotes).toHaveBeenCalled();
    await fireEvent.keyDown(selected, { key: 'h' });
    expect(handlers.onhint).toHaveBeenCalled();
    await fireEvent.keyDown(selected, { key: 'Escape' });
    expect(handlers.onselect).toHaveBeenCalledWith(-1);
  });
});
