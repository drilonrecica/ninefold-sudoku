<script lang="ts">
  import { tick } from 'svelte';

  import {
    boxOf,
    columnOf,
    directConflictIndices,
    isRelated,
    moveIndex,
    rowOf,
  } from '$lib/game/board';
  import type { Digit, MatchCell } from '$lib/game/types';

  import CandidateGrid from './CandidateGrid.svelte';

  interface ParticipantView {
    id: string;
    name: string;
    marker: string;
    colorIndex: number;
  }

  let {
    cells,
    selectedIndex,
    pendingCells,
    softLocks,
    pingedCell = null,
    participants,
    readOnly = false,
    boardLabel = 'Shared Sudoku board',
    onselect,
    onvalue,
    onerase,
    ontogglenotes,
    onhint,
  }: {
    cells: MatchCell[];
    selectedIndex: number;
    pendingCells: Set<number>;
    softLocks: Map<number, string>;
    pingedCell?: number | null;
    participants: Map<string, ParticipantView>;
    readOnly?: boolean;
    boardLabel?: string;
    onselect: (index: number) => void;
    onvalue: (digit: Digit) => void;
    onerase: () => void;
    ontogglenotes: () => void;
    onhint: () => void;
  } = $props();

  const conflicts = $derived(directConflictIndices(cells));
  const selectedValue = $derived(cells[selectedIndex]?.value);

  function participantFor(cell: MatchCell): ParticipantView | undefined {
    return cell.attribution ? participants.get(cell.attribution) : undefined;
  }

  function labelFor(cell: MatchCell): string {
    const parts = [`Row ${rowOf(cell.index) + 1}, column ${columnOf(cell.index) + 1}`];
    parts.push(cell.isClue ? 'clue' : 'editable');
    if (cell.value) parts.push(`value ${cell.value}`);
    else if (cell.notes.length) parts.push(`shared notes ${cell.notes.join(', ')}`);
    else parts.push('empty');
    if (conflicts.has(cell.index)) parts.push('direct conflict');
    if (cell.correct === false) parts.push('incorrect value');
    if (pendingCells.has(cell.index)) parts.push('pending confirmation');
    if (pingedCell === cell.index) parts.push('pinged by a player');
    const contributor = participantFor(cell);
    if (contributor) parts.push(`entered by ${contributor.name}`);
    const locker = softLocks.get(cell.index);
    if (locker) parts.push(`${participants.get(locker)?.name ?? 'Another player'} is working here`);
    return parts.join(', ');
  }

  async function selectAndFocus(index: number) {
    onselect(index);
    await tick();
    document.querySelector<HTMLButtonElement>(`#sudoku-cell-${index}`)?.focus();
  }

  function handleKey(event: KeyboardEvent) {
    if (['ArrowLeft', 'ArrowRight', 'ArrowUp', 'ArrowDown', 'Home', 'End'].includes(event.key)) {
      event.preventDefault();
      void selectAndFocus(moveIndex(selectedIndex, event.key));
      return;
    }
    if (/^[1-9]$/.test(event.key) && !readOnly) {
      event.preventDefault();
      onvalue(Number(event.key) as Digit);
      return;
    }
    if ((event.key === 'Backspace' || event.key === 'Delete') && !readOnly) {
      event.preventDefault();
      onerase();
      return;
    }
    if ((event.key.toLowerCase() === 'n' || event.key === ' ') && !readOnly) {
      event.preventDefault();
      ontogglenotes();
      return;
    }
    if (event.key.toLowerCase() === 'h' && !readOnly) {
      event.preventDefault();
      onhint();
      return;
    }
    if (event.key === 'Escape') {
      event.preventDefault();
      onselect(-1);
    }
  }
</script>

<div
  class="sudoku-board"
  role="grid"
  aria-label={boardLabel}
  aria-rowcount="9"
  aria-colcount="9"
  tabindex="-1"
  onkeydown={handleKey}
>
  {#each [0, 1, 2, 3, 4, 5, 6, 7, 8] as row (row)}
    <div class="grid-row" role="row" aria-rowindex={row + 1}>
      {#each cells.slice(row * 9, row * 9 + 9) as cell (cell.index)}
        {@const contributor = participantFor(cell)}
        {@const lockParticipant = softLocks.get(cell.index)}
        <button
          id={`sudoku-cell-${cell.index}`}
          class={`cell box-${boxOf(cell.index)} player-${contributor?.colorIndex ?? 0}`}
          class:clue={cell.isClue}
          class:selected={cell.index === selectedIndex}
          class:related={selectedIndex >= 0 && isRelated(cell.index, selectedIndex)}
          class:same-value={selectedValue &&
            cell.value === selectedValue &&
            cell.index !== selectedIndex}
          class:conflict={conflicts.has(cell.index)}
          class:incorrect={cell.correct === false}
          class:pending={pendingCells.has(cell.index)}
          class:locked={Boolean(lockParticipant)}
          class:pinged={pingedCell === cell.index}
          class:box-right={columnOf(cell.index) === 2 || columnOf(cell.index) === 5}
          class:box-bottom={rowOf(cell.index) === 2 || rowOf(cell.index) === 5}
          type="button"
          role="gridcell"
          aria-colindex={columnOf(cell.index) + 1}
          aria-selected={cell.index === selectedIndex}
          aria-readonly={cell.isClue || readOnly}
          aria-label={labelFor(cell)}
          tabindex={cell.index === selectedIndex || (selectedIndex < 0 && cell.index === 0)
            ? 0
            : -1}
          onclick={() => onselect(cell.index)}
        >
          {#if cell.value}
            <span class="value">{cell.value}</span>
          {:else if cell.notes.length}
            <CandidateGrid notes={cell.notes} />
          {/if}
          {#if contributor}
            <span class="contributor" aria-hidden="true">{contributor.marker}</span>
          {/if}
          {#if lockParticipant}
            <span class="lock-marker" aria-hidden="true">◈</span>
          {/if}
          {#if conflicts.has(cell.index) || cell.correct === false}
            <span class="problem-marker" aria-hidden="true">!</span>
          {/if}
          {#if pendingCells.has(cell.index)}
            <span class="sr-only">Pending confirmation</span>
          {/if}
        </button>
      {/each}
    </div>
  {/each}
</div>

<style>
  .sudoku-board {
    display: grid;
    width: min(100%, 42rem);
    aspect-ratio: 1;
    grid-template-columns: repeat(9, minmax(24px, 1fr));
    grid-template-rows: repeat(9, minmax(24px, 1fr));
    overflow: hidden;
    border: 3px solid var(--board-line-strong, var(--text-primary));
    border-radius: var(--radius-small);
    background: var(--border-strong);
  }
  .cell {
    position: relative;
    display: grid;
    min-width: 24px;
    min-height: 24px;
    cursor: pointer;
    place-items: center;
    border: 0;
    border-right: 1px solid var(--border-default);
    border-bottom: 1px solid var(--border-default);
    padding: 0;
    background: var(--surface-primary);
    color: var(--brand-primary);
    font-size: clamp(1rem, 5vw, 2.15rem);
    line-height: 1;
  }
  .grid-row {
    display: contents;
  }
  .cell.box-right {
    border-right: 3px solid var(--border-strong);
  }
  .cell.box-bottom {
    border-bottom: 3px solid var(--border-strong);
  }
  .cell.clue {
    cursor: default;
    color: var(--text-primary);
    font-weight: 850;
  }
  .cell.related {
    background: var(--surface-secondary);
  }
  .cell.same-value {
    background: var(--brand-subtle);
  }
  .cell.selected {
    z-index: 2;
    background: var(--surface-selected);
    box-shadow: inset 0 0 0 3px var(--brand-primary);
  }
  .cell:focus-visible {
    z-index: 4;
    outline: 3px solid var(--border-focus);
    outline-offset: -3px;
  }
  .cell.pending .value {
    opacity: 0.62;
    text-decoration: underline dotted 2px;
    text-underline-offset: 0.18em;
  }
  .cell.conflict,
  .cell.incorrect {
    box-shadow: inset 0 0 0 2px var(--state-error);
    text-decoration: underline wavy var(--state-error);
  }
  .cell.locked {
    outline: 2px dashed var(--state-warning);
    outline-offset: -3px;
  }
  .cell.pinged {
    animation: ping-pulse 900ms ease-out;
    box-shadow: inset 0 0 0 3px var(--state-info);
  }
  @keyframes ping-pulse {
    50% {
      background: var(--state-info-subtle);
    }
  }
  .contributor,
  .lock-marker,
  .problem-marker {
    position: absolute;
    font-size: clamp(0.48rem, 1.2vw, 0.72rem);
    line-height: 1;
  }
  .contributor {
    right: 0.16rem;
    bottom: 0.14rem;
    color: currentColor;
  }
  .lock-marker {
    top: 0.12rem;
    left: 0.14rem;
    color: var(--state-warning);
  }
  .problem-marker {
    top: 0.12rem;
    right: 0.14rem;
    color: var(--state-error);
    font-weight: 900;
  }
  .player-1 {
    color: var(--player-1);
  }
  .player-2 {
    color: var(--player-2);
  }
  .player-3 {
    color: var(--player-3);
  }
  .player-4 {
    color: var(--player-4);
  }
  .player-5 {
    color: var(--player-5);
  }
  .player-6 {
    color: var(--player-6);
  }
  @media (prefers-reduced-motion: reduce) {
    .cell,
    .cell.pinged {
      transition: none;
      animation: none;
    }
  }
</style>
