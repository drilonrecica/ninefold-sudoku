<script lang="ts">
  import { browser } from '$app/environment';
  import { onMount } from 'svelte';

  import { t } from '$lib/i18n';

  type Theme = 'system' | 'light' | 'dark';
  let theme = $state<Theme>('system');

  function applyTheme(next: Theme) {
    theme = next;
    if (!browser) return;
    if (next === 'system') delete document.documentElement.dataset.theme;
    else document.documentElement.dataset.theme = next;
    localStorage.setItem('ninefold.theme', next);
  }

  onMount(() => {
    const saved = localStorage.getItem('ninefold.theme');
    applyTheme(saved === 'light' || saved === 'dark' ? saved : 'system');
  });
</script>

<label class="theme-selector">
  <span>{t('theme.label')}</span>
  <select value={theme} onchange={(event) => applyTheme(event.currentTarget.value as Theme)}>
    <option value="system">{t('theme.system')}</option>
    <option value="light">{t('theme.light')}</option>
    <option value="dark">{t('theme.dark')}</option>
  </select>
</label>

<style>
  .theme-selector {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    color: var(--text-secondary);
    font-size: 0.875rem;
    font-weight: 700;
  }
  select {
    min-height: var(--touch-target);
    border: 1px solid var(--border-default);
    border-radius: var(--radius-small);
    padding-inline: var(--space-2);
    background: var(--surface-primary);
    color: var(--text-primary);
  }
</style>
