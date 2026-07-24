<script lang="ts">
  import { goto } from '$app/navigation';
  import { onMount } from 'svelte';

  import type { Difficulty, Room } from '$lib/api/client';
  import { safeErrorKeyFromCode } from '$lib/api/errors';
  import { t } from '$lib/i18n';
  import {
    applyRoomMessage,
    initialRoomState,
    type ConnectionState,
    type RoomClientState,
    type ServerMessage,
  } from '$lib/realtime/room-reducer';
  import { RoomSocket } from '$lib/realtime/socket';

  type LobbyRoom = Room & {
    countdown?: { matchId: string; generation: number; deadlineAt: number };
  };

  const loadingCells = [0, 1, 2, 3, 4, 5, 6, 7, 8];
  let { data } = $props();
  let roomState = $state<RoomClientState>({ ...initialRoomState });
  let socket = $state<RoomSocket | null>(null);
  let now = $state(Date.now());
  let announcement = $state('');
  let commandError = $state('');
  let copied = $state(false);
  let settingsNotice = $state('');

  const room = $derived(roomState.room as LobbyRoom | null);
  const self = $derived(
    room?.participants.find((participant) => participant.id === roomState.selfParticipantId),
  );
  const allPlayersReady = $derived(
    Boolean(room) &&
      room!.participants.some((participant) => participant.role === 'Player') &&
      room!.participants
        .filter((participant) => participant.role === 'Player')
        .every((participant) => participant.isReady),
  );
  const isHost = $derived(Boolean(self?.isHost));
  const canStart = $derived(
    Boolean(room && isHost && allPlayersReady && room.state === 'Lobby' && roomState.isController),
  );
  const countdown = $derived(room?.countdown);
  const countdownValue = $derived(
    countdown ? Math.max(0, Math.ceil((countdown.deadlineAt - now) / 1_000)) : null,
  );

  function connectionLabel(connection: ConnectionState): string {
    const keys = {
      connecting: 'connection.connecting',
      connected: 'connection.connected',
      reconnecting: 'connection.reconnecting',
      read_only: 'connection.readOnly',
      disconnected: 'connection.disconnected',
    } as const;
    return t(keys[connection]);
  }

  function handleMessage(message: ServerMessage) {
    roomState = applyRoomMessage(roomState, message);
    if (message.type === 'command.rejected') {
      commandError = t(safeErrorKeyFromCode(message.payload.code));
    }
    if (message.type === 'command.acknowledged') commandError = '';
    if (message.type === 'room.snapshot') {
      const nextRoom = message.payload.room as LobbyRoom | undefined;
      if (nextRoom?.state === 'InMatch' && nextRoom.currentMatchId) {
        announcement = t('countdown.started');
        void goto(`/play/${nextRoom.currentMatchId}`);
      }
    }
  }

  function sendReady() {
    if (!room || !self || !socket) return;
    socket.roomCommand('room.set_ready', room.id, room.version, { ready: !self.isReady });
  }

  function startMatch() {
    if (!room || !socket || !canStart) return;
    socket.roomCommand('room.start_countdown', room.id, room.version, {});
  }

  function cancelCountdown() {
    if (!room || !socket || !isHost) return;
    socket.roomCommand('room.cancel_countdown', room.id, room.version, {});
    announcement = 'Countdown cancelled. Ready states were reset.';
  }

  function updateDifficulty(event: Event) {
    if (!room || !socket) return;
    const difficulty = (event.currentTarget as HTMLSelectElement).value as Difficulty;
    socket.roomCommand('room.change_settings', room.id, room.version, {
      settings: { difficulty },
    });
    settingsNotice = t('lobby.settingsReset');
  }

  async function copyInvitation() {
    await navigator.clipboard.writeText(`${window.location.origin}/join/${data.code}`);
    copied = true;
    window.setTimeout(() => (copied = false), 2_000);
  }

  onMount(() => {
    const clock = window.setInterval(() => (now = Date.now()), 250);
    socket = new RoomSocket(
      () => ({
        roomCode: data.code,
        roomVersion: roomState.room?.version,
        matchId: roomState.room?.currentMatchId ?? undefined,
      }),
      handleMessage,
      (connection) => {
        roomState = { ...roomState, connection };
      },
    );
    socket.connect();
    return () => {
      clearInterval(clock);
      socket?.close();
    };
  });
</script>

<svelte:head>
  <title>Room Lobby — Ninefold Sudoku</title>
  <meta name="robots" content="noindex, nofollow" />
</svelte:head>

<div class="shell lobby">
  <header class="lobby-header">
    <div>
      <p class="eyebrow">Private co-op room</p>
      <h1>{t('lobby.title')}</h1>
    </div>
    <div
      class:offline={roomState.connection !== 'connected'}
      class="connection-status"
      role="status"
    >
      <span aria-hidden="true">●</span>
      {connectionLabel(roomState.connection)}
    </div>
  </header>

  {#if commandError}
    <p class="error-message" role="alert">{commandError}</p>
  {/if}
  {#if roomState.connection === 'read_only'}
    <div class="status-message">
      <p>This tab is read-only because another tab controls this room.</p>
      <button class="button secondary" type="button" onclick={() => socket?.requestControl()}>
        Control from this tab
      </button>
    </div>
  {/if}

  {#if !room}
    <section class="panel loading" aria-busy="true" aria-label="Loading room">
      <div class="loading-grid" aria-hidden="true">
        {#each loadingCells as cell (cell)}<span></span>{/each}
      </div>
      <p>Synchronizing room…</p>
    </section>
  {:else}
    <section class="room-code panel" aria-labelledby="room-code-title">
      <div>
        <span id="room-code-title" class="label">Room code</span>
        <strong aria-label={`Room code ${room.code.split('').join(' ')}`}>{room.code}</strong>
      </div>
      <button class="button secondary" type="button" onclick={copyInvitation}>
        {t('lobby.share')}
      </button>
      <span class="copy-status" aria-live="polite">{copied ? t('lobby.copied') : ''}</span>
    </section>

    {#if countdown && countdownValue !== null}
      <section class="countdown panel" aria-labelledby="countdown-title">
        <p id="countdown-title" aria-live="assertive">
          {countdownValue > 0
            ? t('countdown.starts', { count: countdownValue })
            : t('countdown.started')}
        </p>
        <strong aria-hidden="true">{countdownValue}</strong>
        {#if isHost}
          <button class="button secondary" type="button" onclick={cancelCountdown}>
            {t('lobby.cancel')}
          </button>
        {/if}
      </section>
    {/if}

    <div class="lobby-grid">
      <section class="panel participants" aria-labelledby="participants-title">
        <h2 id="participants-title">{t('lobby.players')}</h2>
        <ul>
          {#each room.participants as participant, index (participant.id)}
            <li>
              <span class={`player-marker player-${(index % 6) + 1}`} aria-hidden="true">
                {index % 2 === 0 ? '●' : '◆'}
              </span>
              <span class="participant-name">
                <strong>{participant.name}</strong>
                <small>{participant.role}</small>
              </span>
              {#if participant.isHost}<span class="badge">{t('lobby.host')}</span>{/if}
              <span class:ready={participant.isReady} class="ready-state">
                <span aria-hidden="true">{participant.isReady ? '✓' : '○'}</span>
                {participant.isReady ? t('lobby.ready') : t('lobby.notReady')}
              </span>
            </li>
          {/each}
        </ul>
        {#if self?.role === 'Player'}
          <button
            class="button ready-button"
            class:secondary={self.isReady}
            type="button"
            disabled={!roomState.isController || room.state !== 'Lobby'}
            onclick={sendReady}
          >
            {self.isReady ? t('lobby.unreadyAction') : t('lobby.readyAction')}
          </button>
        {/if}
      </section>

      <aside class="panel settings" aria-labelledby="settings-title">
        <h2 id="settings-title">{t('lobby.settings')}</h2>
        <dl>
          <div>
            <dt>Mode</dt>
            <dd>Co-op</dd>
          </div>
          <div>
            <dt>Errors</dt>
            <dd>{room.settings.errorPreset}</dd>
          </div>
          <div>
            <dt>Hints</dt>
            <dd>{room.settings.hintsEnabled ? 'Enabled' : 'Disabled'}</dd>
          </div>
          <div>
            <dt>Shared notes</dt>
            <dd>{room.settings.sharedNotes ? 'Enabled' : 'Disabled'}</dd>
          </div>
        </dl>
        {#if isHost}
          <div class="field">
            <label for="difficulty">Difficulty</label>
            <select
              id="difficulty"
              value={room.settings.difficulty}
              disabled={!roomState.isController || room.state !== 'Lobby'}
              onchange={updateDifficulty}
            >
              {#each ['Easy', 'Medium', 'Hard', 'Expert'] as difficulty (difficulty)}
                <option value={difficulty}>{difficulty}</option>
              {/each}
            </select>
          </div>
        {/if}
        {#if settingsNotice}<p class="status-message" role="status">{settingsNotice}</p>{/if}
      </aside>

      <section class="panel activity" aria-labelledby="activity-title">
        <h2 id="activity-title">{t('lobby.activity')}</h2>
        <ol>
          {#each room.participants.slice(-5) as participant (participant.id)}
            <li>{participant.name} joined the room.</li>
          {/each}
        </ol>
      </section>
    </div>

    <section class="start-panel">
      {#if !isHost}
        <p>{t('lobby.hostOnly')}</p>
      {:else if !allPlayersReady}
        <p id="start-reason">{t('lobby.waitingReady')}</p>
      {/if}
      {#if isHost}
        <button
          class="button start-button"
          type="button"
          disabled={!canStart}
          aria-describedby={!canStart ? 'start-reason' : undefined}
          onclick={startMatch}
        >
          {t('lobby.start')}
        </button>
      {/if}
    </section>
  {/if}

  <div class="sr-only" aria-live="assertive">{announcement}</div>
</div>

<style>
  .lobby {
    padding-block: var(--space-8) var(--space-12);
  }
  .lobby-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-5);
  }
  .eyebrow {
    margin: 0;
    color: var(--brand-primary);
    font-weight: 800;
  }
  h1 {
    margin: var(--space-1) 0 var(--space-6);
    font-size: clamp(2rem, 5vw, 3.25rem);
  }
  h2 {
    margin-top: 0;
    font-size: 1.15rem;
  }
  .connection-status {
    display: inline-flex;
    min-height: var(--touch-target);
    align-items: center;
    gap: var(--space-2);
    border: 1px solid var(--state-success);
    border-radius: 999px;
    padding-inline: var(--space-4);
    color: var(--state-success);
    font-weight: 700;
  }
  .connection-status.offline {
    border-color: var(--state-warning);
    color: var(--state-warning);
  }
  .room-code {
    display: grid;
    grid-template-columns: 1fr auto;
    align-items: center;
    gap: var(--space-4);
    margin-bottom: var(--space-5);
    padding: var(--space-5);
  }
  .room-code > div {
    display: grid;
    gap: var(--space-1);
  }
  .room-code strong {
    font:
      800 clamp(1.8rem, 6vw, 3rem) ui-monospace,
      monospace;
    letter-spacing: 0.16em;
  }
  .copy-status {
    min-height: 1.3em;
    grid-column: 1 / -1;
    color: var(--state-success);
  }
  .lobby-grid {
    display: grid;
    grid-template-columns: minmax(0, 1.5fr) minmax(17rem, 1fr);
    gap: var(--space-5);
  }
  .participants,
  .settings,
  .activity {
    padding: var(--space-5);
  }
  .participants {
    grid-row: span 2;
  }
  ul,
  ol {
    margin: 0;
    padding: 0;
    list-style: none;
  }
  .participants li {
    display: grid;
    grid-template-columns: auto minmax(0, 1fr) auto auto;
    align-items: center;
    gap: var(--space-3);
    min-height: 3.5rem;
    border-bottom: 1px solid var(--border-default);
  }
  .player-marker {
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
  .participant-name {
    display: grid;
  }
  .participant-name small {
    color: var(--text-muted);
  }
  .badge {
    border: 1px solid var(--border-strong);
    border-radius: 999px;
    padding: 0.2rem 0.5rem;
    font-size: 0.75rem;
  }
  .ready-state {
    display: inline-flex;
    align-items: center;
    gap: var(--space-1);
    color: var(--text-muted);
    font-weight: 700;
  }
  .ready-state.ready {
    color: var(--state-success);
  }
  .ready-button {
    margin-top: var(--space-5);
  }
  dl {
    display: grid;
    gap: var(--space-2);
    margin-bottom: var(--space-5);
  }
  dl div {
    display: flex;
    justify-content: space-between;
    gap: var(--space-3);
  }
  dt {
    color: var(--text-muted);
  }
  dd {
    margin: 0;
    font-weight: 700;
  }
  .activity li {
    margin-bottom: var(--space-2);
    color: var(--text-secondary);
  }
  .start-panel {
    display: flex;
    min-height: 6rem;
    align-items: center;
    justify-content: flex-end;
    gap: var(--space-5);
  }
  .start-panel p {
    color: var(--text-muted);
  }
  .start-button {
    min-width: 11rem;
  }
  .countdown {
    display: grid;
    justify-items: center;
    gap: var(--space-3);
    margin-bottom: var(--space-5);
    padding: var(--space-6);
    text-align: center;
  }
  .countdown p {
    margin: 0;
    font-weight: 700;
  }
  .countdown strong {
    color: var(--brand-primary);
    font-size: clamp(3rem, 12vw, 6rem);
    line-height: 1;
  }
  .loading {
    display: grid;
    justify-items: center;
    padding: var(--space-12);
  }
  .loading-grid {
    display: grid;
    grid-template-columns: repeat(3, 1rem);
    gap: 0.25rem;
  }
  .loading-grid span {
    aspect-ratio: 1;
    background: var(--brand-primary);
    opacity: 0.45;
  }
  @media (max-width: 760px) {
    .lobby-header {
      align-items: flex-start;
      flex-direction: column;
    }
    .lobby-grid {
      grid-template-columns: 1fr;
    }
    .participants {
      grid-row: auto;
    }
    .start-panel {
      align-items: stretch;
      flex-direction: column;
      padding-top: var(--space-5);
    }
    .start-button {
      width: 100%;
    }
  }
  @media (max-width: 480px) {
    .room-code {
      grid-template-columns: 1fr;
    }
    .room-code .button {
      width: 100%;
    }
    .participants li {
      grid-template-columns: auto minmax(0, 1fr) auto;
    }
    .badge {
      grid-column: 2;
    }
  }
</style>
