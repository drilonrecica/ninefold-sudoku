<script lang="ts">
  import { onMount } from 'svelte';

  import { defaultGamePreferences, loadGamePreferences } from '$lib/game/preferences';
  import { t } from '$lib/i18n';
  import { clearSoloData } from '$lib/solo/storage';

  type Theme = 'system' | 'light' | 'dark';

  let theme = $state<Theme>('system');
  let inputMode = $state(defaultGamePreferences.inputMode);
  let soundEnabled = $state(false);
  let hapticsEnabled = $state(false);
  let reducedMotion = $state(false);
  let highContrast = $state(false);
  let largerLabels = $state(false);
  let status = $state('');
  let interactive = $state(false);

  onMount(() => {
    const preferences = loadGamePreferences(localStorage);
    inputMode = preferences.inputMode;
    soundEnabled = preferences.soundEnabled;
    hapticsEnabled = preferences.hapticsEnabled;
    const savedTheme = localStorage.getItem('ninefold.theme');
    theme = savedTheme === 'light' || savedTheme === 'dark' ? savedTheme : 'system';
    reducedMotion = localStorage.getItem('ninefold.reducedMotion') === 'on';
    highContrast = localStorage.getItem('ninefold.highContrast') === 'on';
    largerLabels = localStorage.getItem('ninefold.largerLabels') === 'on';
    applyAccessibility();
    interactive = true;
  });

  function setTheme(value: Theme) {
    theme = value;
    localStorage.setItem('ninefold.theme', value);
    if (value === 'system') delete document.documentElement.dataset.theme;
    else document.documentElement.dataset.theme = value;
    status = 'Appearance saved.';
  }

  function saveGameplay() {
    localStorage.setItem('ninefold.inputMode', inputMode);
    localStorage.setItem('ninefold.sound', soundEnabled ? 'on' : 'off');
    localStorage.setItem('ninefold.haptics', hapticsEnabled ? 'on' : 'off');
    status = 'Gameplay preferences saved.';
  }

  function saveAccessibility() {
    localStorage.setItem('ninefold.reducedMotion', reducedMotion ? 'on' : 'off');
    localStorage.setItem('ninefold.highContrast', highContrast ? 'on' : 'off');
    localStorage.setItem('ninefold.largerLabels', largerLabels ? 'on' : 'off');
    applyAccessibility();
    status = 'Accessibility preferences saved.';
  }

  function applyAccessibility() {
    document.documentElement.dataset.reduceMotion = reducedMotion ? 'true' : 'false';
    document.documentElement.dataset.highContrast = highContrast ? 'true' : 'false';
    document.documentElement.dataset.largerLabels = largerLabels ? 'true' : 'false';
  }

  async function clearLocalData() {
    await clearSoloData();
    for (const key of [
      'ninefold.inputMode',
      'ninefold.sound',
      'ninefold.haptics',
      'ninefold.reducedMotion',
      'ninefold.highContrast',
      'ninefold.largerLabels',
      'ninefold.theme',
    ]) {
      localStorage.removeItem(key);
    }
    document.querySelector<HTMLDialogElement>('#clear-local-dialog')?.close();
    setTheme('system');
    status = 'Local preferences, Solo progress, statistics, recent puzzles, and replays cleared.';
  }

  function showConfirmation() {
    document.querySelector<HTMLDialogElement>('#clear-local-dialog')?.showModal();
  }

  function closeConfirmation() {
    document.querySelector<HTMLDialogElement>('#clear-local-dialog')?.close();
  }
</script>

<svelte:head>
  <title>Settings — Ninefold Sudoku</title>
  <meta name="robots" content="noindex, nofollow" />
</svelte:head>

<div class="shell settings">
  <h1>{t('settings.title')}</h1>
  <p>Settings are stored on this device. English is the only language in this release.</p>
  {#if status}<p class="status-message" role="status">{status}</p>{/if}

  <section class="panel">
    <h2>{t('settings.appearance')}</h2>
    <label>
      Theme
      <select value={theme} onchange={(event) => setTheme(event.currentTarget.value as Theme)}>
        <option value="system">System</option>
        <option value="light">Light</option>
        <option value="dark">Dark</option>
      </select>
    </label>
  </section>

  <section class="panel">
    <h2>{t('settings.gameplay')}</h2>
    <label>
      Input preference
      <select bind:value={inputMode} onchange={saveGameplay}>
        <option value="cell-first">Cell first</option>
        <option value="number-first">Number first</option>
      </select>
    </label>
    <label
      ><input type="checkbox" bind:checked={soundEnabled} onchange={saveGameplay} /> Sound</label
    >
    <label
      ><input type="checkbox" bind:checked={hapticsEnabled} onchange={saveGameplay} /> Haptics when supported</label
    >
    <p class="muted">Sound is off by default. Ninefold never auto-plays sound.</p>
  </section>

  <section class="panel">
    <h2>{t('settings.accessibility')}</h2>
    <label
      ><input type="checkbox" bind:checked={reducedMotion} onchange={saveAccessibility} /> Reduce motion</label
    >
    <label
      ><input type="checkbox" bind:checked={highContrast} onchange={saveAccessibility} /> Enhance contrast</label
    >
    <label
      ><input type="checkbox" bind:checked={largerLabels} onchange={saveAccessibility} /> Larger board
      labels</label
    >
    <a href="/how-to-play#keyboard">Review keyboard controls</a>
  </section>

  <section class="panel">
    <h2>{t('settings.privacy')}</h2>
    <p>
      Clearing local data removes preferences, Solo attempts, history, statistics, recent puzzle
      IDs, and local Solo replay from this device. It does not delete shared multiplayer replay or
      leave an active Room.
    </p>
    <button
      id="clear-local-trigger"
      class="button secondary"
      type="button"
      disabled={!interactive}
      onclick={showConfirmation}
    >
      {t('settings.clear')}
    </button>
    <dialog
      id="clear-local-dialog"
      class="confirmation"
      role="alertdialog"
      aria-labelledby="clear-local-title"
      onclose={() => document.querySelector<HTMLButtonElement>('#clear-local-trigger')?.focus()}
    >
      <p id="clear-local-title"><strong>Clear all local Ninefold data on this device?</strong></p>
      <div>
        <button
          id="keep-local-data"
          class="button secondary"
          type="button"
          onclick={closeConfirmation}
        >
          Keep Local Data
        </button>
        <button class="button danger" type="button" onclick={clearLocalData}>
          {t('settings.clear')}
        </button>
      </div>
    </dialog>
    <p><a href="/privacy">Read the privacy details</a>.</p>
  </section>
</div>

<style>
  .settings {
    max-width: 48rem;
    padding-block: var(--space-10);
  }
  .settings > section {
    display: grid;
    gap: var(--space-4);
    margin-top: var(--space-5);
    padding: var(--space-5);
  }
  label {
    display: flex;
    min-height: 44px;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-4);
  }
  select {
    min-height: 44px;
  }
  .confirmation {
    max-width: min(32rem, calc(100% - 2rem));
    border: 1px solid var(--state-error);
    border-radius: var(--radius-control);
    padding: var(--space-4);
    background: var(--state-error-subtle);
    color: var(--text-primary);
  }
  .confirmation::backdrop {
    background: rgb(0 0 0 / 55%);
  }
  .confirmation div {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-2);
  }
  .danger {
    background: var(--state-error);
  }
</style>
