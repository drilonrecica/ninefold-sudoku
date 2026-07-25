<script lang="ts">
  import { onMount } from 'svelte';

  import ThemeSelector from '$lib/components/ThemeSelector.svelte';
  import { pseudoLocalize, t } from '$lib/i18n';

  import '../app.css';

  let { children } = $props();

  onMount(() => {
    document.documentElement.dataset.reduceMotion =
      localStorage.getItem('ninefold.reducedMotion') === 'on' ? 'true' : 'false';
    document.documentElement.dataset.highContrast =
      localStorage.getItem('ninefold.highContrast') === 'on' ? 'true' : 'false';
    document.documentElement.dataset.largerLabels =
      localStorage.getItem('ninefold.largerLabels') === 'on' ? 'true' : 'false';
    if (!import.meta.env.DEV || new URLSearchParams(location.search).get('locale') !== 'pseudo') {
      return;
    }
    document.documentElement.dataset.locale = 'pseudo';
    const walker = document.createTreeWalker(document.body, NodeFilter.SHOW_TEXT);
    const nodes: Text[] = [];
    while (walker.nextNode()) {
      const node = walker.currentNode as Text;
      if (node.textContent?.trim() && !node.parentElement?.closest('script, style'))
        nodes.push(node);
    }
    for (const node of nodes) node.textContent = pseudoLocalize(node.textContent ?? '');
  });
</script>

<svelte:head>
  <title>Ninefold Sudoku — Multiplayer Sudoku</title>
  <meta
    name="description"
    content="Play Sudoku together in private co-op rooms or solve a personal puzzle without creating an account."
  />
  <meta name="theme-color" content="#5145cd" />
</svelte:head>

<a class="skip-link" href="#main-content">Skip to main content</a>
<header class="site-header">
  <div class="shell header-inner">
    <a class="brand" href="/" aria-label="Ninefold Sudoku home">
      <svg aria-hidden="true" viewBox="0 0 30 30" width="30" height="30">
        <path
          d="M2 2h8v8H2zM11 2h8v8h-8zM20 2h8v8h-8zM2 11h8v8H2zM11 11h8v8h-8zM20 11h8v8h-8zM2 20h8v8H2zM11 20h8v8h-8z"
        />
        <path class="brand-cell" d="M20 20h8v8h-8z" />
      </svg>
      <span>{t('brand.name')}</span>
    </a>
    <nav aria-label="Primary navigation">
      <a href="/">{t('nav.play')}</a>
      <a href="/how-to-play">{t('nav.howToPlay')}</a>
      <a href="/privacy">{t('nav.privacy')}</a>
      <a href="/accessibility">{t('nav.accessibility')}</a>
      <a href="/settings">{t('nav.settings')}</a>
    </nav>
    <ThemeSelector />
  </div>
</header>

<div id="connection-status" class="sr-only" aria-live="polite">Application ready</div>
<main id="main-content" tabindex="-1">
  {@render children()}
</main>

<footer>
  <div class="shell footer-inner">
    <span>© {new Date().getFullYear()} Ninefold Sudoku</span>
    <span>No accounts. No ads. No tracking.</span>
  </div>
</footer>

<style>
  .site-header {
    border-bottom: 1px solid var(--border-default);
    background: color-mix(in srgb, var(--surface-canvas) 92%, transparent);
  }
  .header-inner {
    display: grid;
    grid-template-columns: auto 1fr auto;
    align-items: center;
    gap: var(--space-6);
    min-height: 4.5rem;
  }
  .brand {
    display: inline-flex;
    min-height: var(--touch-target);
    align-items: center;
    gap: var(--space-2);
    color: var(--text-primary);
    font-weight: 800;
    text-decoration: none;
  }
  svg {
    fill: var(--text-primary);
  }
  .brand-cell {
    fill: var(--brand-primary);
  }
  nav {
    display: flex;
    justify-content: center;
    gap: var(--space-5);
  }
  nav a {
    display: inline-flex;
    min-height: var(--touch-target);
    align-items: center;
    color: var(--text-secondary);
    font-weight: 650;
    text-decoration: none;
  }
  nav a:hover {
    color: var(--text-primary);
  }
  footer {
    margin-top: var(--space-12);
    border-top: 1px solid var(--border-default);
    color: var(--text-muted);
  }
  .footer-inner {
    display: flex;
    min-height: 5rem;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-4);
    font-size: 0.875rem;
  }
  @media (max-width: 800px) {
    .header-inner {
      grid-template-columns: 1fr auto;
      padding-block: var(--space-2);
    }
    nav {
      grid-column: 1 / -1;
      grid-row: 2;
      justify-content: flex-start;
      gap: var(--space-4);
      overflow-x: auto;
    }
  }
  @media (max-width: 430px) {
    nav a {
      font-size: 0.875rem;
    }
    .footer-inner {
      align-items: flex-start;
      flex-direction: column;
      justify-content: center;
    }
  }
</style>
