<svelte:head>
  <title>Privacy — Ninefold Sudoku</title>
  <meta
    name="description"
    content="How Ninefold Sudoku handles essential cookies, local settings, rooms, and replays."
  />
</svelte:head>

<article class="shell content">
  <h1>Privacy</h1>
  <p class="lead">No accounts. No ads. No behavioral tracking.</p>
  <section>
    <h2>Essential storage</h2>
    <p>
      A secure HTTP-only cookie keeps you connected to one temporary room. Theme and accessibility
      preferences stay in your browser.
    </p>
  </section>
  <section>
    <h2>Temporary room data</h2>
    <p>
      Participant-linked multiplayer data is retained for no more than seven days. Short-lived
      operational logs may be used to keep the service secure.
    </p>
  </section>
  <section>
    <h2>Replays</h2>
    <p>
      Temporary replay data is access-controlled and expires. Replay event integrity is verified
      with a cryptographic hash chain.
    </p>
  </section>
  <section>
    <h2>Contact</h2>
    <p>
      Privacy questions can be sent to <a href="mailto:ninefold@recica.dev">ninefold@recica.dev</a>.
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
