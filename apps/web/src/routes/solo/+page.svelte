<script lang="ts">
  import { onMount } from 'svelte';

  import {
    ApiError,
    createSoloAssignment,
    requestSoloHint,
    validateSoloCompletion,
  } from '$lib/api/client';
  import NumberPad from '$lib/components/NumberPad.svelte';
  import SudokuBoard from '$lib/components/SudokuBoard.svelte';
  import type { Digit, MatchCell } from '$lib/game/types';
  import {
    cellsToValues,
    clearSoloData,
    completeAttempt,
    currentAttempt,
    recentPuzzleIds,
    saveAttempt,
    saveSoloPreference,
    soloPreference,
    soloResults,
    suspendedElapsed,
    type SoloAttempt,
    type SoloDifficulty,
    type SoloPlayStyle,
  } from '$lib/solo/storage';

  let attempt = $state<SoloAttempt | null>(null);
  let resumable = $state<SoloAttempt | null>(null);
  let cells = $state<MatchCell[]>([]);
  let selectedIndex = $state(-1);
  let notesMode = $state(false);
  let loading = $state(true);
  let busy = $state(false);
  let paused = $state(false);
  let message = $state('');
  let storageWarning = $state('');
  let result = $state<{
    elapsedMs: number;
    penaltyMs: number;
    hintsUsed: number;
    mistakes: number;
  } | null>(null);
  let displayElapsedMs = $state(0);
  let preferredStyle = $state<SoloPlayStyle>('Guided');
  let completedCount = $state(0);
  let bestByDifficulty = $state<Record<string, number>>({});
  let runningSince: number | null = null;
  let interval: ReturnType<typeof setInterval> | undefined;

  const emptyParticipants = new Map();

  onMount(() => {
    void loadResume();
    interval = setInterval(updateDisplayTime, 250);
    const visibility = () => {
      if (!attempt?.started || paused) return;
      if (document.hidden) {
        commitElapsed();
        void persist();
      } else {
        runningSince = Date.now();
      }
    };
    const unload = () => {
      commitElapsed();
      void persist();
    };
    document.addEventListener('visibilitychange', visibility);
    window.addEventListener('pagehide', unload);
    return () => {
      clearInterval(interval);
      document.removeEventListener('visibilitychange', visibility);
      window.removeEventListener('pagehide', unload);
    };
  });

  async function loadResume() {
    try {
      resumable = (await currentAttempt()) ?? null;
      preferredStyle =
        ((await soloPreference('soloPlayStyle')) as SoloPlayStyle | undefined) ?? 'Guided';
      const results = await soloResults();
      completedCount = results.length;
      bestByDifficulty = {};
      for (const entry of results) {
        const total = entry.elapsedMs + entry.penaltyMs;
        bestByDifficulty[entry.difficulty] = Math.min(
          bestByDifficulty[entry.difficulty] ?? Number.MAX_SAFE_INTEGER,
          total,
        );
      }
    } catch {
      storageWarning = 'Local storage is unavailable. Progress will last only for this tab.';
    } finally {
      loading = false;
    }
  }

  async function start(difficulty: SoloDifficulty | 'Random', playStyle: SoloPlayStyle) {
    busy = true;
    message = '';
    try {
      const recent = await recentPuzzleIds().catch(() => []);
      const assigned = await createSoloAssignment(difficulty, playStyle, recent);
      attempt = {
        id: assigned.attemptId,
        puzzleId: assigned.puzzleId,
        revision: assigned.revision,
        difficulty: assigned.difficulty,
        playStyle,
        clues: assigned.clues,
        assignmentProof: assigned.assignmentProof,
        values: [...assigned.clues].map(Number),
        notes: Array.from({ length: 81 }, () => []),
        elapsedMs: 0,
        penaltyMs: 0,
        started: false,
        paused: false,
        hintsUsed: 0,
        mistakes: 0,
        updatedAtMs: Date.now(),
        replay: [],
      };
      hydrate(attempt);
      await Promise.all([
        saveSoloPreference('soloPlayStyle', playStyle),
        saveSoloPreference('inputPreference', 'cell-first'),
      ]).catch(() => {});
      await persist();
    } catch (error) {
      message =
        error instanceof ApiError && error.code === 'PUZZLE_UNAVAILABLE'
          ? 'No puzzle is available for that choice. Try another difficulty.'
          : 'Solo play requires a connection. Check your connection and try again.';
    } finally {
      busy = false;
    }
  }

  function resume() {
    if (!resumable) return;
    attempt = $state.snapshot(resumable);
    attempt.paused = false;
    hydrate(attempt);
    if (attempt.started) runningSince = Date.now();
  }

  function hydrate(value: SoloAttempt) {
    cells = [...value.clues].map((clue, index) => ({
      index,
      isClue: clue !== '0',
      value: value.values[index] ? (value.values[index] as Digit) : undefined,
      notes: [...(value.notes[index] ?? [])],
      correct: null,
    }));
    paused = value.paused;
    displayElapsedMs = value.elapsedMs;
    selectedIndex = cells.findIndex((cell) => !cell.isClue);
  }

  function startTimerFor(index: number) {
    if (!attempt || cells[index]?.isClue || attempt.started) return;
    attempt.started = true;
    runningSince = Date.now();
  }

  function selectCell(index: number) {
    selectedIndex = index;
    startTimerFor(index);
  }

  async function enterValue(value: Digit) {
    const cell = cells[selectedIndex];
    if (!attempt || !cell || cell.isClue || paused || busy) return;
    startTimerFor(selectedIndex);
    if (notesMode) {
      cell.notes = cell.notes.includes(value)
        ? cell.notes.filter((digit) => digit !== value)
        : [...cell.notes, value].sort();
      record('note', selectedIndex, value);
    } else {
      cell.value = value;
      cell.notes = [];
      cell.correct = null;
      record('value', selectedIndex, value);
      await persist();
      if (attempt.playStyle === 'Guided') await check(false);
      if (cells.every((entry) => entry.value)) await check(true);
    }
    await persist();
  }

  async function erase() {
    const cell = cells[selectedIndex];
    if (!attempt || !cell || cell.isClue || paused) return;
    delete cell.value;
    cell.correct = null;
    record('erase', selectedIndex);
    await persist();
  }

  function toggleNotes() {
    notesMode = !notesMode;
  }

  async function check(completion: boolean) {
    if (!attempt || busy) return;
    busy = true;
    try {
      const response = await validateSoloCompletion(
        attempt.id,
        attempt.assignmentProof,
        cellsToValues(attempt.clues, cells),
      );
      if (attempt.playStyle === 'Guided') {
        const incorrect = new Set(response.incorrectCells ?? []);
        for (const cell of cells) {
          if (!cell.isClue && cell.value) {
            const wasIncorrect = cell.correct === false;
            cell.correct = !incorrect.has(cell.index);
            if (!wasIncorrect && cell.correct === false) attempt.mistakes++;
          }
        }
        message = incorrect.size ? 'Some values need another look.' : 'No incorrect values found.';
      } else if (completion && !response.complete) {
        message = 'The board is not solved yet. Direct conflicts remain marked on the board.';
      }
      if (response.complete) await finish();
    } catch {
      message = 'The board could not be checked. Check your connection and try again.';
    } finally {
      busy = false;
    }
  }

  async function hint(level: 'Nudge' | 'Reveal') {
    if (!attempt || busy || paused) return;
    busy = true;
    try {
      const response = await requestSoloHint(
        attempt.id,
        attempt.assignmentProof,
        cellsToValues(attempt.clues, cells),
        level,
      );
      attempt.hintsUsed++;
      attempt.penaltyMs += response.penaltyMs;
      record('hint', response.cell, response.value, level);
      if (level === 'Reveal' && response.cell !== undefined && response.value) {
        const cell = cells[response.cell];
        if (cell && !cell.isClue) {
          cell.value = response.value as Digit;
          cell.notes = [];
          cell.correct = true;
          selectedIndex = response.cell;
        }
        message = `Revealed row ${Math.floor(response.cell / 9) + 1}, column ${(response.cell % 9) + 1}. +20 seconds.`;
      } else {
        const cellsText = response.affectedCells?.map((index) => index + 1).join(', ');
        message = `${humanize(response.technique ?? 'next step')}${cellsText ? ` near cells ${cellsText}` : ''}. +20 seconds.`;
      }
      await persist();
      if (cells.every((entry) => entry.value)) await check(true);
    } catch {
      message = 'A hint is not available for the current board.';
    } finally {
      busy = false;
    }
  }

  async function togglePause() {
    if (!attempt || !attempt.started) return;
    if (paused) {
      paused = false;
      attempt.paused = false;
      runningSince = Date.now();
      record('resume');
    } else {
      commitElapsed();
      paused = true;
      attempt.paused = true;
      record('pause');
    }
    await persist();
  }

  async function finish() {
    if (!attempt) return;
    commitElapsed();
    record('complete');
    result = {
      elapsedMs: attempt.elapsedMs,
      penaltyMs: attempt.penaltyMs,
      hintsUsed: attempt.hintsUsed,
      mistakes: attempt.mistakes,
    };
    try {
      await completeAttempt($state.snapshot(attempt), {
        id: attempt.id,
        puzzleId: attempt.puzzleId,
        difficulty: attempt.difficulty,
        playStyle: attempt.playStyle,
        elapsedMs: attempt.elapsedMs,
        penaltyMs: attempt.penaltyMs,
        hintsUsed: attempt.hintsUsed,
        mistakes: attempt.mistakes,
        completedAtMs: Date.now(),
      });
    } catch {
      storageWarning = 'The result could not be saved on this device.';
    }
  }

  async function persist() {
    if (!attempt || result) return;
    attempt.values = cells.map((cell) => cell.value ?? 0);
    attempt.notes = cells.map((cell) => [...cell.notes]);
    attempt.updatedAtMs = Date.now();
    try {
      await saveAttempt($state.snapshot(attempt));
    } catch {
      storageWarning = 'Local storage is unavailable. Progress will last only for this tab.';
    }
  }

  function record(
    type: SoloAttempt['replay'][number]['type'],
    cell?: number,
    value?: number,
    detail?: string,
  ) {
    if (!attempt) return;
    attempt.replay.push({ atMs: currentElapsed(), type, cell, value, detail });
  }

  function commitElapsed() {
    if (!attempt || runningSince === null) return;
    attempt.elapsedMs = suspendedElapsed(attempt.elapsedMs, runningSince, Date.now());
    runningSince = null;
    displayElapsedMs = attempt.elapsedMs;
  }

  function currentElapsed() {
    return (attempt?.elapsedMs ?? 0) + (runningSince === null ? 0 : Date.now() - runningSince);
  }

  function updateDisplayTime() {
    displayElapsedMs = currentElapsed();
  }

  async function clearLocal() {
    if (
      !confirm(
        'Clear Solo attempts, history, statistics, replays, and Solo preferences on this device?',
      )
    ) {
      return;
    }
    await clearSoloData();
    resumable = null;
    attempt = null;
    result = null;
    message = 'Local Solo data cleared. Shared multiplayer replays were not deleted.';
  }

  function formatTime(value: number) {
    const seconds = Math.floor(value / 1000);
    return `${Math.floor(seconds / 60)}:${String(seconds % 60).padStart(2, '0')}`;
  }

  function humanize(value: string) {
    return value.replaceAll('_', ' ').replace(/^./, (character) => character.toUpperCase());
  }
</script>

<svelte:head>
  <title>Play Solo — Ninefold Sudoku</title>
  <meta name="robots" content="noindex, nofollow" />
</svelte:head>

<main class="solo-shell">
  {#if loading}
    <section class="panel message" aria-busy="true"><h1>Loading Solo</h1></section>
  {:else if result}
    <section class="panel result">
      <p class="eyebrow">Solo · {attempt?.difficulty}</p>
      <h1>Puzzle solved</h1>
      <dl>
        <div>
          <dt>Active time</dt>
          <dd>{formatTime(result.elapsedMs)}</dd>
        </div>
        <div>
          <dt>Hint penalty</dt>
          <dd>+{formatTime(result.penaltyMs)}</dd>
        </div>
        <div>
          <dt>Hints</dt>
          <dd>{result.hintsUsed}</dd>
        </div>
        <div>
          <dt>Mistakes</dt>
          <dd>{result.mistakes}</dd>
        </div>
      </dl>
      <div class="actions">
        <button type="button" class="button" onclick={() => location.reload()}>
          Play another puzzle
        </button>
        {#if attempt}
          <a class="button secondary" href={`/solo/replay/${attempt.id}`}>View local replay</a>
        {/if}
        <a class="button secondary" href="/">Return home</a>
      </div>
    </section>
  {:else if !attempt}
    <section class="entry">
      <p class="eyebrow">Account-free practice</p>
      <h1>Play Solo</h1>
      <p>Progress, history, statistics, and replay stay on this device.</p>
      {#if completedCount > 0}
        <p class="muted">
          {completedCount} completed
          {#each Object.entries(bestByDifficulty) as [difficulty, best] (difficulty)}
            · {difficulty} best {formatTime(best)}
          {/each}
        </p>
      {/if}
      {#if resumable}
        <button class="button continue" type="button" onclick={resume}>
          Continue {resumable.difficulty} puzzle
        </button>
      {/if}
      <div class="styles">
        {#each ['Guided', 'Classic'] as style (style)}
          <section class="panel">
            <h2>{style}</h2>
            {#if preferredStyle === style}<p class="muted">Last used</p>{/if}
            <p>
              {style === 'Guided'
                ? 'Mistakes are highlighted immediately, and Check Board is available.'
                : 'Only direct Sudoku conflicts are shown until you finish.'}
            </p>
            <div class="difficulties">
              {#each ['Easy', 'Medium', 'Hard', 'Expert', 'Random'] as difficulty (difficulty)}
                <button
                  type="button"
                  disabled={busy}
                  onclick={() =>
                    start(difficulty as SoloDifficulty | 'Random', style as SoloPlayStyle)}
                  >{difficulty}</button
                >
              {/each}
            </div>
          </section>
        {/each}
      </div>
      <button class="link-danger" type="button" onclick={clearLocal}>Clear local Solo data</button>
    </section>
  {:else}
    <header>
      <div>
        <p class="eyebrow">{attempt.playStyle} · {attempt.difficulty}</p>
        <h1>Solo Sudoku</h1>
      </div>
      <div class="timer" aria-label="Solo timer">
        <strong>{formatTime(displayElapsedMs)}</strong>
        <span>+{formatTime(attempt.penaltyMs)} hints</span>
      </div>
    </header>

    {#if storageWarning}<p class="notice" role="status">{storageWarning}</p>{/if}
    {#if message}<p class="notice" role="status" aria-live="polite">{message}</p>{/if}

    <div class="game-layout">
      <section class="board-wrap">
        <SudokuBoard
          {cells}
          {selectedIndex}
          pendingCells={new Set()}
          softLocks={new Map()}
          participants={emptyParticipants}
          readOnly={paused || busy}
          boardLabel="Solo Sudoku board"
          onselect={selectCell}
          onvalue={enterValue}
          onerase={erase}
          ontogglenotes={toggleNotes}
          onhint={() => hint('Nudge')}
        />
        {#if paused}
          <div class="pause-cover" role="status">Paused. Resume to show the board.</div>
        {/if}
      </section>
      <aside class="panel controls">
        <NumberPad
          {notesMode}
          disabled={paused || busy}
          hintsEnabled={true}
          onvalue={enterValue}
          onerase={erase}
          ontogglenotes={toggleNotes}
          onhint={hint}
        />
        <div class="secondary-actions">
          <button type="button" onclick={togglePause} disabled={!attempt.started}>
            {paused ? 'Resume puzzle' : 'Pause puzzle'}
          </button>
          {#if attempt.playStyle === 'Guided'}
            <button type="button" onclick={() => check(false)} disabled={paused || busy}>
              Check Board
            </button>
          {/if}
        </div>
        <p class="muted">{attempt.hintsUsed} hints · {attempt.mistakes} mistakes</p>
      </aside>
    </div>
  {/if}
</main>

<style>
  .solo-shell {
    width: min(100% - 1.5rem, 72rem);
    margin-inline: auto;
    padding-block: var(--space-6) var(--space-10);
  }
  header,
  .game-layout,
  .actions,
  .secondary-actions {
    display: flex;
    gap: var(--space-4);
  }
  header {
    align-items: start;
    justify-content: space-between;
    margin-bottom: var(--space-4);
  }
  h1,
  .eyebrow {
    margin-block: 0 var(--space-2);
  }
  .eyebrow {
    color: var(--brand-primary);
    font-weight: 800;
  }
  .timer {
    display: grid;
    text-align: right;
  }
  .timer strong {
    font-size: 1.5rem;
  }
  .game-layout {
    align-items: start;
  }
  .board-wrap {
    position: relative;
    width: min(100%, 42rem);
  }
  .controls,
  .result,
  .message {
    padding: var(--space-5);
  }
  .controls {
    flex: 1;
    min-width: 16rem;
  }
  .secondary-actions {
    flex-wrap: wrap;
    margin-top: var(--space-4);
  }
  .secondary-actions button {
    min-height: 44px;
    flex: 1;
  }
  .pause-cover {
    position: absolute;
    inset: 0;
    display: grid;
    place-items: center;
    border-radius: var(--radius-panel);
    background: var(--surface-primary);
    font-weight: 800;
  }
  .entry {
    max-width: 58rem;
    margin-inline: auto;
  }
  .continue {
    margin-block: var(--space-4);
  }
  .styles {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: var(--space-4);
    margin-block: var(--space-5);
  }
  .styles section {
    padding: var(--space-5);
  }
  .difficulties {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-2);
  }
  .difficulties button,
  .link-danger {
    min-height: 44px;
  }
  .notice {
    border: 1px solid var(--border-default);
    border-radius: var(--radius-control);
    padding: var(--space-3);
    background: var(--surface-secondary);
  }
  dl {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: var(--space-3);
  }
  dl div {
    padding: var(--space-3);
    background: var(--surface-secondary);
  }
  dd {
    margin: 0;
    font-weight: 800;
  }
  @media (max-width: 52rem) {
    .game-layout {
      display: grid;
    }
    .controls {
      min-width: 0;
    }
  }
  @media (max-width: 42rem) {
    .styles {
      grid-template-columns: 1fr;
    }
  }
</style>
