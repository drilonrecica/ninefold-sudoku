<script lang="ts">
  import { page } from '$app/state';
  import { onMount } from 'svelte';

  import SudokuBoard from '$lib/components/SudokuBoard.svelte';
  import type { Digit, MatchCell } from '$lib/game/types';
  import { soloDatabase, type SoloReplay } from '$lib/solo/storage';

  let replay = $state<SoloReplay | null>(null);
  let cells = $state<MatchCell[]>([]);
  let position = $state(0);
  let status = $state<'loading' | 'ready' | 'missing'>('loading');

  onMount(async () => {
    try {
      replay = (await soloDatabase().soloReplays.get(page.params.id ?? '')) ?? null;
      if (!replay) {
        status = 'missing';
        return;
      }
      seek(0);
      status = 'ready';
    } catch {
      status = 'missing';
    }
  });

  function seek(next: number) {
    if (!replay) return;
    position = Math.max(0, Math.min(next, replay.events.length));
    cells = [...replay.clues].map((clue, index) => ({
      index,
      isClue: clue !== '0',
      value: clue === '0' ? undefined : (Number(clue) as Digit),
      notes: [],
    }));
    for (const event of replay.events.slice(0, position)) {
      const cell = event.cell === undefined ? undefined : cells[event.cell];
      if (!cell) continue;
      if (event.type === 'value' && event.value) {
        cell.value = event.value as Digit;
        cell.notes = [];
      } else if (event.type === 'erase') {
        delete cell.value;
      } else if (event.type === 'note' && event.value) {
        const digit = event.value as Digit;
        cell.notes = cell.notes.includes(digit)
          ? cell.notes.filter((value) => value !== digit)
          : [...cell.notes, digit].sort();
      }
    }
  }
</script>

<svelte:head>
  <title>Local Solo Replay — Ninefold Sudoku</title>
  <meta name="robots" content="noindex, nofollow" />
</svelte:head>

<div class="shell replay">
  {#if status === 'loading'}
    <section class="panel state" aria-busy="true"><h1>Loading local replay</h1></section>
  {:else if status === 'missing' || !replay}
    <section class="panel state">
      <h1>Local replay unavailable</h1>
      <p>It may have been cleared from this device.</p>
      <a class="button" href="/solo">Play Solo</a>
    </section>
  {:else}
    <header>
      <p class="eyebrow">Device-local · {replay.playStyle} · {replay.difficulty}</p>
      <h1>Solo replay</h1>
      <p>This replay never left this device.</p>
    </header>
    <div class="layout">
      <SudokuBoard
        {cells}
        selectedIndex={-1}
        pendingCells={new Set()}
        softLocks={new Map()}
        participants={new Map()}
        readOnly={true}
        boardLabel="Local Solo replay board"
        onselect={() => {}}
        onvalue={() => {}}
        onerase={() => {}}
        ontogglenotes={() => {}}
        onhint={() => {}}
      />
      <aside class="panel controls">
        <p>Event {position} of {replay.events.length}</p>
        <input
          aria-label="Replay position"
          type="range"
          min="0"
          max={replay.events.length}
          value={position}
          oninput={(event) => seek(Number(event.currentTarget.value))}
        />
        <div>
          <button type="button" onclick={() => seek(position - 1)} disabled={position === 0}>
            Previous
          </button>
          <button
            type="button"
            onclick={() => seek(position + 1)}
            disabled={position === replay.events.length}>Next</button
          >
        </div>
      </aside>
    </div>
  {/if}
</div>

<style>
  .replay {
    padding-block: var(--space-6) var(--space-10);
  }
  .eyebrow {
    color: var(--brand-primary);
    font-weight: 800;
  }
  .layout {
    display: grid;
    grid-template-columns: minmax(18rem, 42rem) minmax(14rem, 1fr);
    gap: var(--space-5);
    align-items: start;
  }
  .controls,
  .state {
    padding: var(--space-5);
  }
  input {
    width: 100%;
    min-height: 44px;
  }
  .controls div {
    display: flex;
    gap: var(--space-2);
  }
  .controls button {
    min-height: 44px;
  }
  @media (max-width: 48rem) {
    .layout {
      grid-template-columns: 1fr;
    }
  }
</style>
