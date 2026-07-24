<script lang="ts">
  import type { Digit } from '$lib/game/types';

  let {
    notesMode,
    disabled = false,
    hintsEnabled = false,
    onvalue,
    onerase,
    ontogglenotes,
    onhint,
  }: {
    notesMode: boolean;
    disabled?: boolean;
    hintsEnabled?: boolean;
    onvalue: (digit: Digit) => void;
    onerase: () => void;
    ontogglenotes: () => void;
    onhint: (level: 'Nudge' | 'Reveal') => void;
  } = $props();
</script>

<section class="number-controls" aria-label="Number input">
  <div class="digits">
    {#each [1, 2, 3, 4, 5, 6, 7, 8, 9] as digit (digit)}
      <button type="button" {disabled} onclick={() => onvalue(digit as Digit)}>
        {digit}
      </button>
    {/each}
  </div>
  <div class="actions">
    <button type="button" {disabled} onclick={onerase}>Erase</button>
    <button
      type="button"
      class:active={notesMode}
      aria-pressed={notesMode}
      {disabled}
      onclick={ontogglenotes}>Notes <span class="shortcut">N</span></button
    >
    {#if hintsEnabled}
      <button type="button" {disabled} onclick={() => onhint('Nudge')}>Nudge</button>
      <button type="button" {disabled} onclick={() => onhint('Reveal')}>Reveal</button>
    {/if}
  </div>
</section>

<style>
  .number-controls {
    display: grid;
    gap: var(--space-3);
  }
  .digits {
    display: grid;
    grid-template-columns: repeat(9, minmax(44px, 1fr));
    gap: var(--space-1);
  }
  button {
    min-width: 44px;
    min-height: 44px;
    border: 1px solid var(--border-strong);
    border-radius: var(--radius-control);
    background: var(--surface-interactive);
    color: var(--text-primary);
    font-weight: 800;
  }
  button:hover:not(:disabled) {
    background: var(--surface-secondary);
  }
  .actions {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-2);
  }
  .actions button {
    flex: 1 1 6rem;
  }
  button.active {
    border: 2px solid var(--brand-primary);
    background: var(--brand-subtle);
  }
  .shortcut {
    color: var(--text-muted);
    font-size: 0.75rem;
  }
  @media (max-width: 580px) {
    .digits {
      grid-template-columns: repeat(5, minmax(44px, 1fr));
    }
  }
</style>
