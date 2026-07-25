<svelte:head>
  <title>How to Play — Ninefold Sudoku</title>
  <meta
    name="description"
    content="Learn Sudoku basics and how private Ninefold co-op rooms work."
  />
  <link rel="canonical" href="https://ninefold.recica.dev/how-to-play" />
  <meta property="og:title" content="How to Play — Ninefold Sudoku" />
  <meta
    property="og:description"
    content="Learn Sudoku basics, private co-op play, Solo styles, hints, reconnect, and replay."
  />
</svelte:head>

<article class="shell content">
  <h1>How to play Ninefold Sudoku</h1>
  <section>
    <h2>Sudoku basics</h2>
    <p>Fill each row, column, and 3×3 box with the digits 1 through 9 without repeating a digit.</p>
  </section>
  <section>
    <h2>Selecting cells and entering values</h2>
    <p>
      Select an editable cell, then choose a digit. Clues are fixed. Direct conflicts are marked
      with text and a symbol as well as color. Erase removes an entered value.
    </p>
  </section>
  <section>
    <h2>Notes</h2>
    <p>
      Turn on Notes before entering digits to record candidates. In Co-op, notes are shared. When
      automatic note removal is enabled, an accepted value removes that digit from related cells.
    </p>
  </section>
  <section>
    <h2>Guided and Classic Solo</h2>
    <p>
      Guided marks solution-wrong entries immediately and provides Check Board. Classic shows only
      direct Sudoku conflicts until the board is submitted for completion. Both styles support
      notes, pause, Nudge, and Reveal.
    </p>
  </section>
  <section>
    <h2>Co-op rooms</h2>
    <p>
      Create a private room, share its six-character code, ready up, and solve one board together.
      The server decides which simultaneous move is accepted.
    </p>
  </section>
  <section>
    <h2>Error presets</h2>
    <p>
      Casual accepts entries and may show correctness. Challenge rejects solution-wrong entries and
      adds a five-second result penalty. Blind hides solution correctness. Clean accepts only
      solution-correct entries. Direct row, column, and box conflicts remain visible.
    </p>
  </section>
  <section>
    <h2>Hints</h2>
    <p>
      Nudge points toward a deterministic next step without entering a value. Reveal enters one
      eligible value. Solo hints add 20 seconds to the result. Explain hints are not part of this
      release.
    </p>
  </section>
  <section>
    <h2>Keyboard controls</h2>
    <p>
      Use arrow keys to move, digits to enter values, Backspace or Delete to erase, and N or Space
      to toggle notes, H for a Nudge where hints are available, and Escape to clear selection.
    </p>
  </section>
  <section>
    <h2>Replay</h2>
    <p>
      Completed multiplayer matches can be played, paused, stepped, and scrubbed. Integrity is shown
      as verified only after the browser checks the event chain and server signature. Solo replays
      stay on the device.
    </p>
  </section>
  <section>
    <h2>Accessible play</h2>
    <p>
      Every cell announces its row, column, clue or editable state, value, notes, conflict, and
      multiplayer attribution when present. Status is never communicated by color alone.
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
