<script lang="ts">
  import { goto } from '$app/navigation';
  import { onMount } from 'svelte';

  import { ApiError, createRoom, leaveRoom, type Difficulty } from '$lib/api/client';
  import { safeErrorKey } from '$lib/api/errors';
  import { t } from '$lib/i18n';

  const difficulties: Difficulty[] = ['Easy', 'Medium', 'Hard', 'Expert'];
  let displayName = $state('');
  let difficulty = $state<Difficulty>('Medium');
  let error = $state('');
  let activeRoomCode = $state('');
  let submitting = $state(false);
  let interactive = $state(false);

  onMount(() => {
    interactive = true;
  });

  async function submit() {
    error = '';
    activeRoomCode = '';
    const name = displayName.trim();
    if (name.length < 1 || name.length > 24) {
      error = t('error.NAME_INVALID');
      return;
    }
    submitting = true;
    try {
      const response = await createRoom(name, difficulty);
      await goto(`/room/${response.room.code}`);
    } catch (cause) {
      error = t(safeErrorKey(cause));
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
      error = t(safeErrorKey(cause));
      submitting = false;
    }
  }
</script>

<svelte:head>
  <title>Create Room — Ninefold Sudoku</title>
  <meta name="robots" content="noindex, nofollow" />
</svelte:head>

<div class="shell create-page">
  <a class="back-link" href="/">← Back to play</a>
  <section aria-labelledby="create-title">
    <p class="eyebrow">{t('create.mode')}</p>
    <h1 id="create-title">{t('create.title')}</h1>
    <form
      class="panel"
      onsubmit={(event) => {
        event.preventDefault();
        void submit();
      }}
    >
      <div class="field">
        <label for="display-name">{t('create.name')}</label>
        <input
          id="display-name"
          bind:value={displayName}
          autocomplete="nickname"
          minlength="1"
          maxlength="24"
          required
          aria-describedby="name-help"
        />
        <small id="name-help" class="muted">{t('create.nameHelp')}</small>
      </div>

      <fieldset>
        <legend>{t('create.difficulty')}</legend>
        <div class="difficulty-grid">
          {#each difficulties as option (option)}
            <label class:checked={difficulty === option}>
              <input type="radio" name="difficulty" value={option} bind:group={difficulty} />
              <span>{option}</span>
            </label>
          {/each}
        </div>
      </fieldset>

      {#if error}
        <div class="error-message" role="alert">
          <p>{error}</p>
          {#if activeRoomCode}
            <div class="replacement-actions">
              <a class="button secondary" href={`/room/${activeRoomCode}`}>Return to current room</a
              >
              <button class="button" type="button" disabled={submitting} onclick={replaceRoom}>
                Leave and create new room
              </button>
            </div>
          {/if}
        </div>
      {/if}

      <button class="button submit" type="submit" disabled={submitting || !interactive}>
        {submitting ? 'Creating room…' : t('create.submit')}
      </button>
    </form>
  </section>
</div>

<style>
  .create-page {
    max-width: 42rem;
    padding-block: var(--space-8) var(--space-12);
  }
  .back-link {
    display: inline-flex;
    min-height: var(--touch-target);
    align-items: center;
    margin-bottom: var(--space-4);
  }
  .eyebrow {
    margin-bottom: var(--space-2);
    color: var(--brand-primary);
    font-weight: 800;
  }
  h1 {
    margin-top: 0;
    font-size: clamp(2rem, 6vw, 3.25rem);
  }
  form {
    display: grid;
    gap: var(--space-6);
    padding: clamp(1.25rem, 5vw, 2rem);
  }
  fieldset {
    min-width: 0;
    margin: 0;
    border: 0;
    padding: 0;
  }
  legend {
    margin-bottom: var(--space-3);
    font-weight: 700;
  }
  .difficulty-grid {
    display: grid;
    grid-template-columns: repeat(4, minmax(0, 1fr));
    gap: var(--space-2);
  }
  .difficulty-grid label {
    display: grid;
    min-height: var(--touch-target);
    cursor: pointer;
    place-items: center;
    border: 1px solid var(--border-default);
    border-radius: var(--radius-control);
    background: var(--surface-interactive);
    font-weight: 700;
  }
  .difficulty-grid label.checked {
    border: 2px solid var(--brand-primary);
    background: var(--brand-subtle);
  }
  .difficulty-grid input {
    position: absolute;
    opacity: 0;
  }
  .difficulty-grid label:has(input:focus-visible) {
    outline: 3px solid var(--border-focus);
    outline-offset: 3px;
  }
  .error-message p {
    margin-top: 0;
  }
  .replacement-actions {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-3);
  }
  .submit {
    justify-self: start;
    min-width: 10rem;
  }
  @media (max-width: 520px) {
    .difficulty-grid {
      grid-template-columns: repeat(2, minmax(0, 1fr));
    }
    .submit {
      width: 100%;
    }
  }
</style>
