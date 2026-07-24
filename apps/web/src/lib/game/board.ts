import type { MatchCell } from './types';

export function rowOf(index: number): number {
  return Math.floor(index / 9);
}

export function columnOf(index: number): number {
  return index % 9;
}

export function boxOf(index: number): number {
  return Math.floor(rowOf(index) / 3) * 3 + Math.floor(columnOf(index) / 3);
}

export function isRelated(first: number, second: number): boolean {
  return (
    first !== second &&
    (rowOf(first) === rowOf(second) ||
      columnOf(first) === columnOf(second) ||
      boxOf(first) === boxOf(second))
  );
}

export function directConflictIndices(cells: MatchCell[]): Set<number> {
  const conflicts = new Set<number>();
  cells.forEach((cell) => {
    if (!cell.value) return;
    cells.forEach((peer) => {
      if (peer.value === cell.value && isRelated(cell.index, peer.index)) {
        conflicts.add(cell.index);
        conflicts.add(peer.index);
      }
    });
  });
  return conflicts;
}

export function moveIndex(index: number, key: string): number {
  const row = rowOf(index);
  const column = columnOf(index);
  switch (key) {
    case 'ArrowLeft':
      return row * 9 + ((column + 8) % 9);
    case 'ArrowRight':
      return row * 9 + ((column + 1) % 9);
    case 'ArrowUp':
      return ((row + 8) % 9) * 9 + column;
    case 'ArrowDown':
      return ((row + 1) % 9) * 9 + column;
    case 'Home':
      return row * 9;
    case 'End':
      return row * 9 + 8;
    default:
      return index;
  }
}
