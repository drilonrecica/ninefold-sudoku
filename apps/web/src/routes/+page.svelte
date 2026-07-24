<script lang="ts">
  import { t } from '$lib/i18n';

  const motifCells = [0, 1, 2, 3, 4, 5, 6, 7, 8];
</script>

<svelte:head>
  <link rel="canonical" href="https://ninefold.recica.dev/" />
  <meta property="og:title" content="Ninefold Sudoku — Multiplayer Sudoku" />
  <meta property="og:description" content="Create a private room and solve Sudoku together." />
</svelte:head>

<div class="shell home">
  <section class="hero" aria-labelledby="home-title">
    <p class="eyebrow">Private multiplayer Sudoku</p>
    <h1 id="home-title">{t('home.title')}</h1>
    <p class="supporting">{t('home.supporting')}</p>

    <div class="play-panel panel">
      <a class="button create" href="/create">{t('home.create')}</a>
      <span class="divider" aria-hidden="true">or</span>
      <form method="get" action="/join">
        <label for="room-code">{t('home.code')}</label>
        <div class="join-controls">
          <input
            id="room-code"
            name="code"
            inputmode="text"
            autocomplete="off"
            autocapitalize="characters"
            maxlength="9"
            pattern={'[A-Za-z2-9 ]{6,9}'}
            required
            aria-describedby="code-help"
          />
          <button class="button secondary" type="submit">{t('home.join')}</button>
        </div>
        <small id="code-help" class="muted">Six characters, without 0, 1, I, or O.</small>
      </form>
    </div>
  </section>

  <section class="coop panel" aria-labelledby="coop-title">
    <div class="board-motif" aria-hidden="true">
      {#each motifCells as index (index)}
        <span class:accent={index === 4 || index === 8}></span>
      {/each}
    </div>
    <div>
      <h2 id="coop-title">{t('home.coopTitle')}</h2>
      <p>{t('home.coopBody')}</p>
    </div>
  </section>

  <section aria-labelledby="future-title">
    <h2 id="future-title">{t('home.futureTitle')}</h2>
    <div class="future-grid">
      <article class="panel">
        <h3>{t('home.race')}</h3>
        <p class="muted">{t('home.futureStatus')}</p>
      </article>
      <article class="panel">
        <h3>{t('home.duel')}</h3>
        <p class="muted">{t('home.futureStatus')}</p>
      </article>
    </div>
  </section>

  <section class="commitments" aria-label="Product commitments">
    <p><strong>Privacy:</strong> {t('home.privacy')}</p>
    <p><strong>Accessibility:</strong> {t('home.accessibility')}</p>
  </section>
</div>

<style>
  .home {
    display: grid;
    gap: clamp(3rem, 7vw, 5rem);
    padding-block: clamp(3rem, 8vw, 6rem);
  }
  .hero {
    max-width: 58rem;
  }
  .eyebrow {
    margin: 0 0 var(--space-3);
    color: var(--brand-primary);
    font-weight: 800;
    letter-spacing: 0.04em;
    text-transform: uppercase;
  }
  h1 {
    max-width: 14ch;
    margin: 0;
    font-size: clamp(2.6rem, 8vw, 5.6rem);
    line-height: 0.98;
    letter-spacing: -0.055em;
  }
  .supporting {
    max-width: 58ch;
    margin: var(--space-6) 0 var(--space-8);
    color: var(--text-secondary);
    font-size: clamp(1.1rem, 2vw, 1.35rem);
    line-height: 1.6;
  }
  .play-panel {
    display: grid;
    grid-template-columns: auto auto minmax(16rem, 1fr);
    align-items: center;
    gap: var(--space-5);
    padding: var(--space-5);
  }
  .create {
    min-width: 11rem;
  }
  .divider {
    color: var(--text-muted);
  }
  form {
    display: grid;
    gap: var(--space-2);
  }
  label {
    font-weight: 700;
  }
  .join-controls {
    display: grid;
    grid-template-columns: 1fr auto;
    gap: var(--space-2);
  }
  input {
    min-height: var(--touch-target);
    min-width: 0;
    border: 1px solid var(--border-strong);
    border-radius: var(--radius-control);
    padding: 0.65rem 0.75rem;
    background: var(--surface-primary);
    color: var(--text-primary);
    font:
      700 1.1rem ui-monospace,
      monospace;
    letter-spacing: 0.12em;
    text-transform: uppercase;
  }
  .coop {
    display: grid;
    grid-template-columns: auto 1fr;
    align-items: center;
    gap: var(--space-8);
    padding: var(--space-8);
  }
  .coop h2,
  .coop p,
  article h3,
  article p {
    margin-block: 0 var(--space-2);
  }
  .board-motif {
    display: grid;
    grid-template-columns: repeat(3, 2rem);
    gap: 0.25rem;
  }
  .board-motif span {
    aspect-ratio: 1;
    border: 1px solid var(--border-strong);
    border-radius: 0.25rem;
    background: var(--surface-secondary);
  }
  .board-motif .accent {
    border-color: var(--brand-primary);
    background: var(--brand-subtle);
  }
  section > h2 {
    margin-top: 0;
    font-size: clamp(1.75rem, 4vw, 2.4rem);
  }
  .future-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: var(--space-4);
  }
  article {
    padding: var(--space-5);
  }
  .commitments {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: var(--space-5);
    border-top: 1px solid var(--border-default);
    padding-top: var(--space-8);
    color: var(--text-secondary);
  }
  @media (max-width: 700px) {
    .play-panel {
      grid-template-columns: 1fr;
    }
    .divider {
      display: none;
    }
    .create {
      width: 100%;
    }
    .coop {
      align-items: start;
      grid-template-columns: 1fr;
    }
    .future-grid,
    .commitments {
      grid-template-columns: 1fr;
    }
  }
  @media (max-width: 390px) {
    .join-controls {
      grid-template-columns: 1fr;
    }
  }
</style>
