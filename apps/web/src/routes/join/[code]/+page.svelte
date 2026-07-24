<script lang="ts">
  import { goto } from '$app/navigation';
  import { onMount } from 'svelte';

  import { ApiError, joinRoom, leaveRoom } from '$lib/api/client';
  import { safeErrorKey } from '$lib/api/errors';
  import { t } from '$lib/i18n';

  let { data } = $props();
  let displayName = $state('');
  let role = $state<'Player' | 'Spectator'>('Player');
  let message = $state('');
  let activeRoomCode = $state('');
  let submitting = $state(false);
  let interactive = $state(false);

  onMount(() => {
    interactive = true;
  });

  const playerSeatsFull = $derived(data.preview.playerSeatsAvailable === 0);
  const spectatorAvailable = $derived(data.preview.spectatorSeatsAvailable > 0);

  async function submit() {
    message = '';
    activeRoomCode = '';
    if (displayName.trim().length < 1 || displayName.trim().length > 24) {
      message = t('error.NAME_INVALID');
      return;
    }
    submitting = true;
    try {
      const response = await joinRoom(data.code, displayName.trim(), role);
      await goto(`/room/${response.room.code}`);
    } catch (cause) {
      message = t(safeErrorKey(cause));
      if (cause instanceof ApiError && typeof cause.details.roomCode === 'string') {
        activeRoomCode = cause.details.roomCode;
      }
    } finally {
      submitting = false;
    }
  }

  async function replaceRoom() {
    if (!activeRoomCode) return;
    submitting = true;
    try {
      await leaveRoom(activeRoomCode);
      await submit();
    } catch (cause) {
      message = t(safeErrorKey(cause));
      submitting = false;
    }
  }
</script>

<svelte:head>
  <title>Join Room — Ninefold Sudoku</title>
  <meta name="robots" content="noindex, nofollow" />
</svelte:head>

<div class="shell join-page">
  <a class="back-link" href="/">← Back to play</a>
  <section aria-labelledby="join-title">
    <h1 id="join-title">{t('join.title', { code: data.code })}</h1>
    <div class="preview panel" aria-label="Room preview">
      <div>
        <span class="label">Mode</span>
        <strong>Co-op</strong>
      </div>
      <div>
        <span class="label">Difficulty</span>
        <strong>{data.preview.difficulty}</strong>
      </div>
      <div>
        <span class="label">Player seats</span>
        <strong>{data.preview.playerSeatsAvailable} available</strong>
      </div>
      <div>
        <span class="label">Room state</span>
        <strong>{data.preview.locked ? 'Locked' : data.preview.state}</strong>
      </div>
    </div>

    {#if data.preview.locked}
      <p class="error-message" role="alert">{t('error.ROOM_LOCKED')}</p>
    {:else if playerSeatsFull && !spectatorAvailable}
      <p class="error-message" role="alert">{t('error.ROOM_FULL')}</p>
    {:else}
      <form
        class="panel"
        onsubmit={(event) => {
          event.preventDefault();
          void submit();
        }}
      >
        <div class="field">
          <label for="display-name">{t('join.name')}</label>
          <input
            id="display-name"
            bind:value={displayName}
            autocomplete="nickname"
            minlength="1"
            maxlength="24"
            required
          />
        </div>

        {#if playerSeatsFull && spectatorAvailable}
          <fieldset>
            <legend>Player seats are full</legend>
            <p class="muted">You can join this room as a spectator.</p>
            <label class="role-option">
              <input type="radio" bind:group={role} value="Spectator" />
              Join as Spectator
            </label>
          </fieldset>
        {/if}

        {#if message}
          <div class="error-message" role="alert">
            <p>{message}</p>
            {#if activeRoomCode}
              <div class="replacement-actions">
                <a class="button secondary" href={`/room/${activeRoomCode}`}
                  >Return to current room</a
                >
                <button class="button" type="button" disabled={submitting} onclick={replaceRoom}>
                  Leave and join this room
                </button>
              </div>
            {/if}
          </div>
        {/if}

        <button class="button" type="submit" disabled={submitting || !interactive}>
          {submitting ? 'Joining room…' : t('join.submit')}
        </button>
      </form>
    {/if}
  </section>
</div>

<style>
  .join-page {
    max-width: 46rem;
    padding-block: var(--space-8) var(--space-12);
  }
  .back-link {
    display: inline-flex;
    min-height: var(--touch-target);
    align-items: center;
  }
  h1 {
    font-size: clamp(2rem, 6vw, 3.25rem);
  }
  .preview {
    display: grid;
    grid-template-columns: repeat(4, minmax(0, 1fr));
    gap: var(--space-4);
    margin-bottom: var(--space-5);
    padding: var(--space-5);
  }
  .preview > div {
    display: grid;
    gap: var(--space-1);
  }
  .preview .label {
    color: var(--text-muted);
    font-size: 0.8rem;
  }
  form {
    display: grid;
    gap: var(--space-5);
    padding: clamp(1.25rem, 5vw, 2rem);
  }
  fieldset {
    margin: 0;
    border: 1px solid var(--border-default);
    border-radius: var(--radius-control);
    padding: var(--space-4);
  }
  .role-option {
    display: flex;
    min-height: var(--touch-target);
    align-items: center;
    gap: var(--space-3);
    font-weight: 700;
  }
  .role-option input {
    width: 1.25rem;
    height: 1.25rem;
  }
  .error-message p {
    margin-top: 0;
  }
  .replacement-actions {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-3);
  }
  @media (max-width: 640px) {
    .preview {
      grid-template-columns: repeat(2, minmax(0, 1fr));
    }
    form > .button {
      width: 100%;
    }
  }
</style>
