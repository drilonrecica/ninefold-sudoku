<svelte:head>
  <title>How to Play — Ninefold Sudoku</title>
  <meta
    name="description"
    content="Learn Sudoku basics and how private Ninefold co-op rooms work."
  />
</svelte:head>

<article class="shell content">
  <h1>How to play Ninefold Sudoku</h1>
  <section>
    <h2>Sudoku basics</h2>
    <p>Fill each row, column, and 3×3 box with the digits 1 through 9 without repeating a digit.</p>
  </section>
  <section>
    <h2>Co-op rooms</h2>
    <p>
      Create a private room, share its six-character code, ready up, and solve one board together.
      The server decides which simultaneous move is accepted.
    </p>
  </section>
  <section>
    <h2>Keyboard and notes</h2>
    <p>
      Use arrow keys to move, digits to enter values, Backspace or Delete to erase, and N or Space
      to toggle notes.
    </p>
  </section>
  <section>
    <h2>Reconnect</h2>
    <p>
      If your connection drops, confirmed state is preserved. Input pauses until the room is
      synchronized again.
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
    line-height: 1.05;
  }
  section {
    margin-top: var(--space-8);
  }
</style>
