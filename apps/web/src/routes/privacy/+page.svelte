<svelte:head>
  <title>Privacy — Ninefold Sudoku</title>
  <meta
    name="description"
    content="How Ninefold Sudoku handles essential cookies, local settings, rooms, and replays."
  />
  <link rel="canonical" href="https://ninefold.recica.dev/privacy" />
  <meta property="og:title" content="Privacy — Ninefold Sudoku" />
  <meta
    property="og:description"
    content="How Ninefold handles essential cookies, device-local Solo data, temporary rooms, replays, and logs."
  />
</svelte:head>

<article class="shell content">
  <h1>Privacy</h1>
  <p class="lead">No accounts. No ads. No behavioral tracking.</p>
  <section>
    <h2>Essential storage</h2>
    <p>
      A secure, HTTP-only, same-site cookie keeps you connected to one temporary room. It is
      essential for Room access, so no cookie banner is shown. Theme preferences use small local
      browser storage.
    </p>
  </section>
  <section>
    <h2>Solo data on this device</h2>
    <p>
      Solo progress, notes, timers, assignment proof, recent puzzle IDs, history, statistics,
      preferences, and Solo replay are stored in IndexedDB on this device. The server validates
      assignments, hints, and completion but does not store personal Solo progress or history.
      Clearing browser storage may remove this data permanently.
    </p>
  </section>
  <section>
    <h2>Temporary room data</h2>
    <p>
      Rooms use temporary display names and server-authoritative Match events. Participant-linked
      multiplayer data is retained for no more than seven days. It is then scrubbed, leaving only a
      non-identifying operational tombstone for up to 30 days.
    </p>
  </section>
  <section>
    <h2>Replays</h2>
    <p>
      Multiplayer replay access uses an unguessable link capability and expires after seven days. A
      participant with the still-valid originating Room session can delete it sooner for everyone.
      Replay event integrity is checked with a hash chain and server signature. Shared replay
      deletion is separate from clearing device-local Solo data.
    </p>
  </section>
  <section>
    <h2>Logs and administration</h2>
    <p>
      Operational logs are used for reliability and security and are retained according to the
      service policy. They do not intentionally contain cookies, session tokens, assignment proofs,
      puzzle solutions, private notes, or full request and WebSocket payloads. Private
      administration is restricted and destructive actions are audited.
    </p>
  </section>
  <section>
    <h2>No tracking</h2>
    <p>
      Ninefold has no advertising, product analytics, behavioral tracking, third-party tracking
      scripts, public player directory, or account profile. Network addresses may be processed
      transiently for abuse prevention without becoming a player profile.
    </p>
  </section>
  <section>
    <h2>Clear data</h2>
    <p>
      <a href="/settings">Settings</a> can clear preferences, Solo attempts, history, statistics, recent
      puzzles, and local replays from this device. It does not delete a shared multiplayer replay; that
      action requires the originating Room session.
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
