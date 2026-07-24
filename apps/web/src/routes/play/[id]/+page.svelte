<script lang="ts">
  import { goto } from '$app/navigation';
  import { onMount } from 'svelte';
  import { SvelteMap } from 'svelte/reactivity';

  import type { Room } from '$lib/api/client';
  import NumberPad from '$lib/components/NumberPad.svelte';
  import SudokuBoard from '$lib/components/SudokuBoard.svelte';
  import {
    applyMatchMessage,
    beginPending,
    displayedCell,
    initialMatchState,
    markPendingUncertain,
  } from '$lib/game/match-reducer';
  import { loadGamePreferences } from '$lib/game/preferences';
  import type { Digit, MatchClientState, PendingCommand } from '$lib/game/types';
  import {
    applyRoomMessage,
    initialRoomState,
    type RoomClientState,
    type ServerMessage,
  } from '$lib/realtime/room-reducer';
  import {
    loadCheckpointByMatchId,
    saveCheckpoint,
    type ConnectionCheckpoint,
  } from '$lib/realtime/checkpoint';
  import { RoomSocket } from '$lib/realtime/socket';

  let { data } = $props();
  let matchState = $state<MatchClientState>({ ...initialMatchState, pending: new Map() });
  let roomState = $state<RoomClientState>({ ...initialRoomState });
  let socket = $state<RoomSocket | null>(null);
  let selectedIndex = $state(0);
  let notesMode = $state(false);
  let inputMode = $state<'cell-first' | 'number-first'>('cell-first');
  let selectedDigit = $state<Digit | null>(null);
  let now = $state(Date.now());
  const softLocks = new SvelteMap<number, string>();
  let reactions = $state<string[]>([]);
  let lockWarning = $state<{ cell: number; participantId: string } | null>(null);
  let connectionAnnouncement = $state('');
  let leaveRequestId = $state('');
  let pingedCell = $state<number | null>(null);
  let persistedCheckpoint = $state<ConnectionCheckpoint | null>(null);

  const match = $derived(matchState.confirmed);
  const room = $derived(roomState.room as Room | null);
  const cells = $derived(
    match ? match.cells.map((_, index) => displayedCell(matchState, index)!) : [],
  );
  const pendingCells = $derived(
    new Set(
      [...matchState.pending.values()]
        .map((command) => command.cell)
        .filter((cell): cell is number => cell !== undefined),
    ),
  );
  const participantViews = $derived.by(() => {
    const views = new SvelteMap<
      string,
      { id: string; name: string; marker: string; colorIndex: number }
    >();
    room?.participants.forEach((participant, index) => {
      views.set(participant.id, {
        id: participant.id,
        name: participant.name,
        marker: ['●', '◆', '▲', '■', '★', '⬟'][index % 6] ?? '●',
        colorIndex: (index % 6) + 1,
      });
    });
    return views;
  });
  const readOnly = $derived(
    !roomState.isController ||
      roomState.connection !== 'connected' ||
      match?.state !== 'Active' ||
      matchState.recoveryRequired,
  );
  const elapsedMs = $derived(
    match?.startedAt
      ? Math.max(0, now - match.startedAt - (match.pausedMs ?? 0)) + match.penaltiesMs
      : 0,
  );
  const selectedLocker = $derived(
    lockWarning ? participantViews.get(lockWarning.participantId) : undefined,
  );

  function formatElapsed(milliseconds: number): string {
    const seconds = Math.floor(milliseconds / 1_000);
    return `${Math.floor(seconds / 60)
      .toString()
      .padStart(2, '0')}:${(seconds % 60).toString().padStart(2, '0')}`;
  }

  function handleMessage(message: ServerMessage) {
    roomState = applyRoomMessage(roomState, message);
    matchState = applyMatchMessage(matchState, message);
    if (roomState.room) {
      persistedCheckpoint = {
        roomCode: roomState.room.code,
        roomVersion: roomState.room.version,
        matchId: matchState.confirmed?.id ?? data.matchId,
        matchVersion: matchState.confirmed?.version,
        matchEventNumber: matchState.lastEventNumber,
        pendingRequestIds: [...matchState.pending.keys()],
        updatedAt: Date.now(),
      };
      void saveCheckpoint(persistedCheckpoint);
    }

    if (message.type === 'ephemeral.focus') {
      const participantId = message.payload.participantId;
      const cell = message.payload.cell;
      if (
        participantId &&
        participantId !== roomState.selfParticipantId &&
        typeof cell === 'number'
      ) {
        if (message.payload.focused === false) softLocks.delete(cell);
        else {
          softLocks.set(cell, participantId);
          window.setTimeout(() => {
            if (softLocks.get(cell) === participantId) softLocks.delete(cell);
          }, 8_000);
        }
      }
    }
    if (message.type === 'ephemeral.reaction' && message.payload.reaction) {
      const participant = participantViews.get(message.payload.participantId ?? '');
      reactions = [
        ...reactions.slice(-2),
        `${participant?.name ?? 'A player'}: ${message.payload.reaction.replaceAll('_', ' ')}`,
      ];
    }
    if (message.type === 'match.event') {
      const event = message.payload.event as
        | { type?: string; payload?: { cell?: number; participantId?: string; intent?: string } }
        | undefined;
      if (event?.type === 'Ping' && typeof event.payload?.cell === 'number') {
        pingedCell = event.payload.cell;
        const participant = participantViews.get(event.payload.participantId ?? '');
        reactions = [
          ...reactions.slice(-2),
          `${participant?.name ?? 'A player'} pinged cell ${event.payload.cell + 1}: ${(event.payload.intent ?? 'look_here').replaceAll('_', ' ')}`,
        ];
        window.setTimeout(() => {
          if (pingedCell === event.payload!.cell) pingedCell = null;
        }, 2_000);
      }
    }
    if (message.type === 'connection.accepted' || message.type === 'connection.status') {
      connectionAnnouncement = roomState.isController
        ? 'Connected and synchronized.'
        : 'Connected in read-only mode.';
    }
    if (message.type === 'command.rejected') {
      connectionAnnouncement = `Command rejected: ${message.payload.code ?? 'unknown error'}.`;
      if (message.payload.code === 'STALE_VERSION') socket?.synchronize();
    }
    if (message.type === 'command.acknowledged' && message.payload.requestId === leaveRequestId) {
      void goto('/');
    }
  }

  function selectCell(index: number, override = false) {
    if (selectedIndex >= 0 && selectedIndex !== index) socket?.publishFocus(selectedIndex, false);
    if (index < 0) {
      selectedIndex = -1;
      lockWarning = null;
      return;
    }
    const lockedBy = softLocks.get(index);
    if (lockedBy && !override) {
      lockWarning = { cell: index, participantId: lockedBy };
      return;
    }
    selectedIndex = index;
    lockWarning = null;
    socket?.publishFocus(index, true);
    if (inputMode === 'number-first' && selectedDigit) sendDigit(selectedDigit);
  }

  function queueCommand(
    command: Omit<PendingCommand, 'requestId' | 'createdAt' | 'uncertain'>,
    requestId: string,
  ) {
    matchState = beginPending(matchState, {
      ...command,
      requestId,
      createdAt: performance.now(),
      uncertain: false,
    });
  }

  function sendDigit(digit: Digit) {
    if (!match || !socket || readOnly || selectedIndex < 0 || matchState.pending.size > 0) return;
    const cell = cells[selectedIndex];
    if (!cell || cell.isClue) return;
    if (inputMode === 'number-first') selectedDigit = digit;
    if (notesMode) {
      const removing = cell.notes.includes(digit);
      const requestId = socket.matchCommand(
        removing ? 'match.remove_note' : 'match.add_note',
        match.id,
        match.version,
        { cell: selectedIndex, digits: [digit] },
      );
      if (!requestId) return;
      queueCommand(
        { kind: removing ? 'remove_note' : 'add_note', cell: selectedIndex, digit },
        requestId,
      );
      return;
    }
    const requestId = socket.matchCommand('match.place_value', match.id, match.version, {
      cell: selectedIndex,
      value: digit,
    });
    if (!requestId) return;
    queueCommand({ kind: 'place', cell: selectedIndex, digit }, requestId);
    socket.publishFocus(selectedIndex, false);
  }

  function erase() {
    if (!match || !socket || readOnly || selectedIndex < 0 || matchState.pending.size > 0) return;
    const cell = cells[selectedIndex];
    if (!cell || cell.isClue || !cell.value) return;
    const requestId = socket.matchCommand('match.erase_value', match.id, match.version, {
      cell: selectedIndex,
    });
    if (!requestId) return;
    queueCommand({ kind: 'erase', cell: selectedIndex }, requestId);
  }

  function useHint(level: 'Nudge' | 'Reveal') {
    if (!match || !socket || readOnly || !match.rules.hintsEnabled) return;
    const requestId = socket.matchCommand('match.use_hint', match.id, match.version, {
      level,
      targetCell: level === 'Nudge' && selectedIndex >= 0 ? selectedIndex : undefined,
    });
    if (!requestId) return;
    queueCommand({ kind: 'hint', cell: selectedIndex >= 0 ? selectedIndex : undefined }, requestId);
  }

  function sendPing(intent: 'look_here' | 'unsure' | 'try_this_area') {
    if (!match || !socket || readOnly || selectedIndex < 0) return;
    socket.matchCommand('match.ping', match.id, match.version, { cell: selectedIndex, intent });
  }

  function toggleInputMode() {
    inputMode = inputMode === 'cell-first' ? 'number-first' : 'cell-first';
    localStorage.setItem('ninefold.inputMode', inputMode);
    selectedDigit = null;
  }

  function leaveMatch() {
    if (!room || !socket || leaveRequestId) return;
    const requestId = socket.roomCommand('room.leave', room.id, room.version, {
      intent: 'leave_match',
    });
    if (requestId) leaveRequestId = requestId;
  }

  onMount(() => {
    inputMode = loadGamePreferences(localStorage).inputMode;
    const clock = window.setInterval(() => (now = Date.now()), 1_000);
    let disposed = false;
    void loadCheckpointByMatchId(data.matchId)
      .catch(() => null)
      .then((checkpoint) => {
        if (disposed) return;
        persistedCheckpoint = checkpoint;
        socket = new RoomSocket(
          () => ({
            roomCode: room?.code ?? persistedCheckpoint?.roomCode ?? '',
            roomVersion: room?.version ?? persistedCheckpoint?.roomVersion,
            matchId: data.matchId,
            matchEventNumber:
              matchState.lastEventNumber || persistedCheckpoint?.matchEventNumber || 0,
          }),
          handleMessage,
          (connection) => {
            roomState = { ...roomState, connection };
            connectionAnnouncement =
              connection === 'reconnecting'
                ? 'Connection lost. Reconnecting.'
                : connection === 'offline'
                  ? 'You are offline. Gameplay is paused.'
                  : connection === 'synchronizing'
                    ? 'Connected. Syncing the latest moves.'
                    : connection === 'read_only'
                      ? 'This Room is active in another tab.'
                      : connection === 'maintenance'
                        ? 'Ninefold is restarting. Reconnecting.'
                        : connection === 'recovery_failed'
                          ? 'This Match could not be restored.'
                          : 'Connecting.';
          },
          (requestId) => {
            matchState = markPendingUncertain(matchState, requestId);
          },
        );
        socket.restoreUncertain(checkpoint?.pendingRequestIds ?? []);
        socket.connect();
      });
    return () => {
      disposed = true;
      clearInterval(clock);
      if (selectedIndex >= 0) socket?.publishFocus(selectedIndex, false);
      socket?.close();
    };
  });
</script>

<svelte:head>
  <title>Co-op Match — Ninefold Sudoku</title>
  <meta name="robots" content="noindex, nofollow" />
</svelte:head>

<div class="game-shell">
  <header class="match-header">
    <div>
      <p class="eyebrow">{match?.rules.difficulty ?? 'Co-op'} Sudoku</p>
      <h1>Shared board</h1>
    </div>
    <dl class="match-metrics">
      <div>
        <dt>Time</dt>
        <dd>{formatElapsed(elapsedMs)}</dd>
      </div>
      <div>
        <dt>Penalty</dt>
        <dd>+{Math.floor((match?.penaltiesMs ?? 0) / 1_000)}s</dd>
      </div>
      <div>
        <dt>Status</dt>
        <dd>{roomState.connection.replaceAll('_', ' ')}</dd>
      </div>
    </dl>
  </header>

  {#if matchState.recoveryRequired}
    <section class="recovery-banner" role="alert">
      <p>Updates arrived out of order. Gameplay is paused while the board resynchronizes.</p>
      <button class="button secondary" type="button" onclick={() => socket?.synchronize()}>
        Resynchronize board
      </button>
    </section>
  {/if}

  {#if roomState.connection === 'reconnecting' || roomState.connection === 'offline'}
    <section class="recovery-banner" role="status">
      <p>Connection lost. Reconnecting… Confirmed moves remain visible; gameplay is paused.</p>
    </section>
  {:else if roomState.connection === 'synchronizing'}
    <section class="recovery-banner" role="status">
      <p>Connected. Syncing the latest moves…</p>
    </section>
  {:else if roomState.connection === 'maintenance' || match?.state === 'RecoveryPending'}
    <section class="recovery-banner" role="status">
      <p>Ninefold restarted and is restoring this Match.</p>
    </section>
  {:else if roomState.connection === 'read_only' || (match && !roomState.isController)}
    <section class="recovery-banner" role="status">
      <p>This Room is active in another tab.</p>
      <button class="button secondary" type="button" onclick={() => socket?.requestControl()}>
        Control from this tab
      </button>
    </section>
  {:else if roomState.connection === 'recovery_failed' || match?.state === 'Cancelled'}
    <section class="recovery-banner" role="alert">
      <p>This Match could not be restored. Return home to create or join another Room.</p>
      <a class="button secondary" href="/">Return home</a>
    </section>
  {/if}

  {#if !match}
    <section class="panel loading" aria-busy="true">
      <h2>Synchronizing match</h2>
      <p>The authoritative board is loading.</p>
    </section>
  {:else}
    <div class="game-layout">
      <aside class="panel roster" aria-labelledby="players-title">
        <h2 id="players-title">Players</h2>
        <ul>
          {#each room?.participants ?? [] as participant, index (participant.id)}
            <li>
              <span class={`player-symbol player-${(index % 6) + 1}`} aria-hidden="true">
                {participantViews.get(participant.id)?.marker}
              </span>
              <span
                ><strong>{participant.name}</strong><small
                  >{participant.isHost ? 'Host' : 'Player'}</small
                ></span
              >
              <span>{match.contributions[participant.id] ?? 0} entries</span>
            </li>
          {/each}
        </ul>
        <div class="rules-summary">
          <strong>{match.rules.errorPreset}</strong>
          <span>{match.rules.hintsEnabled ? 'Nudge and Reveal enabled' : 'Hints disabled'}</span>
          <span>Shared notes · No undo</span>
        </div>
      </aside>

      <section class="board-column" aria-label="Sudoku controls">
        <SudokuBoard
          {cells}
          {selectedIndex}
          {pendingCells}
          {softLocks}
          {pingedCell}
          participants={participantViews}
          {readOnly}
          onselect={selectCell}
          onvalue={sendDigit}
          onerase={erase}
          ontogglenotes={() => (notesMode = !notesMode)}
          onhint={() => useHint('Nudge')}
        />

        {#if lockWarning && selectedLocker}
          <div class="lock-warning" role="status">
            <p>{selectedLocker.name} is working on this cell.</p>
            <button type="button" onclick={() => (lockWarning = null)}>Choose another</button>
            <button type="button" onclick={() => selectCell(lockWarning!.cell, true)}
              >Use anyway</button
            >
          </div>
        {/if}

        <NumberPad
          {notesMode}
          disabled={readOnly || matchState.pending.size > 0}
          hintsEnabled={match.rules.hintsEnabled}
          onvalue={sendDigit}
          onerase={erase}
          ontogglenotes={() => (notesMode = !notesMode)}
          onhint={useHint}
        />
      </section>

      <aside class="panel tools" aria-labelledby="tools-title">
        <h2 id="tools-title">Collaboration</h2>
        <button class="button secondary" type="button" onclick={toggleInputMode}>
          Input: {inputMode === 'cell-first' ? 'Cell first' : 'Number first'}
        </button>
        <fieldset disabled={readOnly || selectedIndex < 0}>
          <legend>Ping selected cell</legend>
          <button type="button" onclick={() => sendPing('look_here')}>Look here</button>
          <button type="button" onclick={() => sendPing('unsure')}>Unsure</button>
          <button type="button" onclick={() => sendPing('try_this_area')}>Try this area</button>
        </fieldset>
        <fieldset disabled={readOnly}>
          <legend>Reaction</legend>
          <button type="button" onclick={() => socket?.sendReaction('agree')}>Agree</button>
          <button type="button" onclick={() => socket?.sendReaction('nice_move')}>Nice move</button>
        </fieldset>
        <div aria-live="polite">
          {#each reactions as reaction, index (`${reaction}-${index}`)}<p>{reaction}</p>{/each}
        </div>
        <details>
          <summary>Keyboard shortcuts</summary>
          <p>Arrows move. 1–9 enters. N toggles notes. Delete erases. H requests a nudge.</p>
        </details>
        <button
          class="button secondary leave"
          type="button"
          disabled={Boolean(leaveRequestId)}
          onclick={leaveMatch}>Leave match</button
        >
      </aside>
    </div>
  {/if}

  <div class="sr-only" aria-live="polite">{matchState.announcement}</div>
  <div class="sr-only" aria-live="assertive">
    {connectionAnnouncement}
    {match?.state === 'Completed' ? 'Puzzle completed.' : ''}
  </div>
</div>

<style>
  .game-shell {
    width: min(100% - 1.5rem, 92rem);
    margin-inline: auto;
    padding-block: var(--space-5) var(--space-10);
  }
  .match-header {
    display: flex;
    align-items: end;
    justify-content: space-between;
    gap: var(--space-5);
    margin-bottom: var(--space-5);
  }
  .eyebrow {
    margin: 0;
    color: var(--brand-primary);
    font-weight: 800;
  }
  h1 {
    margin: var(--space-1) 0 0;
    font-size: clamp(1.7rem, 4vw, 2.6rem);
  }
  h2 {
    margin-top: 0;
    font-size: 1.05rem;
  }
  .match-metrics {
    display: flex;
    gap: var(--space-5);
    margin: 0;
  }
  .match-metrics div {
    display: grid;
  }
  .match-metrics dt {
    color: var(--text-muted);
    font-size: 0.75rem;
  }
  .match-metrics dd {
    margin: 0;
    font-weight: 800;
    text-transform: capitalize;
  }
  .game-layout {
    display: grid;
    grid-template-columns: minmax(12rem, 0.7fr) minmax(24rem, 42rem) minmax(13rem, 0.8fr);
    align-items: start;
    justify-content: center;
    gap: var(--space-5);
  }
  .roster,
  .tools {
    padding: var(--space-4);
  }
  .leave {
    margin-top: var(--space-4);
  }
  .roster ul {
    margin: 0;
    padding: 0;
    list-style: none;
  }
  .roster li {
    display: grid;
    grid-template-columns: auto 1fr auto;
    align-items: center;
    gap: var(--space-2);
    min-height: 3rem;
    border-bottom: 1px solid var(--border-default);
    font-size: 0.875rem;
  }
  .roster li > span:nth-child(2) {
    display: grid;
  }
  .roster small,
  .rules-summary {
    color: var(--text-muted);
  }
  .player-symbol {
    font-size: 1.1rem;
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
  .rules-summary {
    display: grid;
    gap: var(--space-1);
    margin-top: var(--space-4);
    font-size: 0.8rem;
  }
  .board-column {
    display: grid;
    min-width: 0;
    gap: var(--space-4);
  }
  .tools {
    display: grid;
    gap: var(--space-4);
  }
  .tools fieldset {
    display: grid;
    gap: var(--space-2);
    margin: 0;
    border: 0;
    padding: 0;
  }
  .tools legend {
    margin-bottom: var(--space-2);
    font-weight: 750;
  }
  .tools fieldset button,
  .lock-warning button {
    min-height: 44px;
    border: 1px solid var(--border-default);
    border-radius: var(--radius-control);
    background: var(--surface-interactive);
    color: var(--text-primary);
  }
  .lock-warning,
  .recovery-banner {
    border: 1px solid var(--state-warning);
    border-radius: var(--radius-control);
    padding: var(--space-3);
    background: var(--state-warning-subtle);
  }
  .lock-warning {
    display: flex;
    align-items: center;
    gap: var(--space-2);
  }
  .lock-warning p {
    flex: 1;
    margin: 0;
  }
  .recovery-banner {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-4);
    margin-bottom: var(--space-4);
  }
  .loading {
    padding: var(--space-8);
  }
  @media (max-width: 1040px) {
    .game-layout {
      grid-template-columns: minmax(24rem, 42rem) minmax(13rem, 18rem);
    }
    .roster {
      grid-column: 1 / -1;
      grid-row: 2;
    }
  }
  @media (max-width: 760px) {
    .match-header {
      align-items: flex-start;
      flex-direction: column;
    }
    .match-metrics {
      width: 100%;
      justify-content: space-between;
    }
    .game-layout {
      grid-template-columns: minmax(0, 1fr);
    }
    .board-column {
      grid-row: 1;
    }
    .roster {
      grid-column: auto;
      grid-row: auto;
    }
    .tools {
      grid-row: auto;
    }
  }
  @media (max-width: 430px) {
    .game-shell {
      width: min(100% - 0.5rem, 92rem);
    }
    .match-metrics {
      gap: var(--space-2);
      font-size: 0.8rem;
    }
    .lock-warning,
    .recovery-banner {
      align-items: stretch;
      flex-direction: column;
    }
  }
</style>
