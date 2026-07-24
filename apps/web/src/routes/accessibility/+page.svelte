<svelte:head>
  <title>Accessibility — Ninefold Sudoku</title>
  <meta
    name="description"
    content="Ninefold Sudoku accessibility support and WCAG 2.2 AA target."
  />
</svelte:head>

<article class="shell content">
  <h1>Accessibility</h1>
  <p class="lead">Ninefold targets WCAG 2.2 AA for its core play flows.</p>
  <section>
    <h2>Input</h2>
    <p>
      Create, join, lobby, and gameplay controls support keyboard and touch operation with visible
      focus.
    </p>
  </section>
  <section>
    <h2>Board semantics</h2>
    <p>
      Sudoku cells expose row, column, clue, value, notes, contributor, conflict, and pending state
      to assistive technology.
    </p>
  </section>
  <section>
    <h2>Display preferences</h2>
    <p>
      The interface supports light and dark themes, reduced motion, text expansion, and reflow at
      200% zoom.
    </p>
  </section>
  <section>
    <h2>Feedback</h2>
    <p>
      Connection, readiness, errors, and player identity use text and shapes rather than color
      alone.
    </p>
  </section>
  <section>
    <h2>Contact</h2>
    <p>
      Report an accessibility issue to <a href="mailto:ninefold@recica.dev">ninefold@recica.dev</a>.
    </p>
  </section>
</article>

<style>
  .content {
    max-width: 70ch;
    padding-block: var(--space-12);
    line-height: 1.75;
  }
  h1 {
    font-size: clamp(2.2rem, 6vw, 3.8rem);
  }
  .lead {
    font-size: 1.25rem;
    font-weight: 700;
  }
  section {
    margin-top: var(--space-8);
  }
</style>
