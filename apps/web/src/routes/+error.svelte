<script lang="ts">
  import { page } from '$app/state';
</script>

<svelte:head>
  <title>Page unavailable — Ninefold Sudoku</title>
  <meta name="robots" content="noindex" />
</svelte:head>

<div class="shell error-page">
  <p class="eyebrow">Error {page.status}</p>
  <h1>
    {page.status === 404
      ? 'That page could not be found.'
      : 'This page is temporarily unavailable.'}
  </h1>
  <p>Private room links can expire. Check the room code or return to the play page.</p>
  <a class="button" href="/">Return to play</a>
</div>

<style>
  .error-page {
    max-width: 46rem;
    padding-block: clamp(4rem, 12vw, 8rem);
  }
  .eyebrow {
    color: var(--state-error);
    font-weight: 800;
  }
  h1 {
    font-size: clamp(2rem, 6vw, 3.5rem);
  }
  p {
    color: var(--text-secondary);
    line-height: 1.7;
  }
</style>
