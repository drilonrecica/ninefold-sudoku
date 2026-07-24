<script lang="ts">
  import { replaceState } from '$app/navigation';
  import { onMount } from 'svelte';
  import { SvelteMap } from 'svelte/reactivity';

  import SudokuBoard from '$lib/components/SudokuBoard.svelte';
  import { replayStateAt, type ReplayDocument, type ReplayState } from '$lib/replay/reducer';

  let { data } = $props();
  let replay = $state<ReplayDocument | null>(null);
  let projection = $state<ReplayState | null>(null);
  let capability = $state('');
  let status = $state<'loading' | 'ready' | 'unavailable'>('loading');
  let playing = $state(false);
  let speed = $state<0.5 | 1 | 2 | 4>(1);
  let copied = $state(false);
  let timer: ReturnType<typeof setTimeout> | undefined;

  const participants = $derived.by(() => {
    const views = new SvelteMap<
      string,
      { id: string; name: string; marker: string; colorIndex: number }
    >();
    replay?.participants.forEach((participant, index) => {
      views.set(participant.id, {
        ...participant,
        marker: ['●', '◆', '▲', '■', '★', '⬟'][index % 6] ?? '●',
        colorIndex: (index % 6) + 1,
      });
    });
    return views;
  });

  function seek(boundary: number) {
    if (!replay) return;
    projection = replayStateAt(replay, boundary);
    if (boundary >= replay.events.length) playing = false;
  }

  function schedule() {
    clearTimeout(timer);
    if (!playing || !replay || !projection || projection.eventIndex >= replay.events.length) return;
    timer = setTimeout(() => {
      seek(projection!.eventIndex + 1);
      schedule();
    }, 600 / speed);
  }

  function togglePlayback() {
    playing = !playing;
    if (playing && replay && projection?.eventIndex === replay.events.length) seek(0);
    schedule();
  }

  function setSpeed(value: 0.5 | 1 | 2 | 4) {
    speed = value;
    schedule();
  }

  async function copyShareLink() {
    if (!capability) return;
    const shareURL = `${location.origin}/replay/${data.replayId}#cap=${capability}`;
    await navigator.clipboard.writeText(shareURL);
    copied = true;
    setTimeout(() => (copied = false), 2_000);
  }

  onMount(() => {
    const fragment = new URLSearchParams(location.hash.slice(1));
    capability = fragment.get('cap') ?? '';
    replaceState(`${location.pathname}${location.search}`, {});
    if (!capability) {
      status = 'unavailable';
      return;
    }
    void fetch(`/api/v1/replays/${encodeURIComponent(data.replayId)}`, {
      headers: { Authorization: `Bearer ${capability}`, Accept: 'application/json' },
      credentials: 'omit',
      referrerPolicy: 'no-referrer',
      cache: 'no-store',
    })
      .then(async (response) => {
        if (!response.ok) throw new Error('unavailable');
        replay = (await response.json()) as ReplayDocument;
        projection = replayStateAt(replay, 0);
        status = 'ready';
      })
      .catch(() => {
        capability = '';
        status = 'unavailable';
      });
    return () => clearTimeout(timer);
  });
</script>

<svelte:head>
  <title>Replay — Ninefold Sudoku</title>
  <meta name="robots" content="noindex, nofollow" />
  <meta name="referrer" content="no-referrer" />
</svelte:head>

<main class="replay-shell">
  {#if status === 'loading'}
    <section class="panel message" aria-busy="true">
      <h1>Loading replay</h1>
      <p>Reconstructing the accepted event history.</p>
    </section>
  {:else if status === 'unavailable' || !replay || !projection}
    <section class="panel message">
      <h1>Replay unavailable</h1>
      <p>This link is invalid, expired, revoked, or no longer available.</p>
      <a class="button secondary" href="/">Return home</a>
    </section>
  {:else}
    <header>
      <div>
        <p class="eyebrow">Co-op · {replay.rules.difficulty}</p>
        <h1>Match replay</h1>
        <p>Integrity signature pending. The accepted server event history is shown as recorded.</p>
      </div>
      <button class="button secondary" type="button" onclick={copyShareLink}>
        {copied ? 'Link copied' : 'Copy replay link'}
      </button>
    </header>

    <div class="replay-layout">
      <section class="board-column" aria-label="Replay board">
        <SudokuBoard
          cells={projection.cells}
          selectedIndex={-1}
          pendingCells={new Set()}
          softLocks={new Map()}
          pingedCell={projection.pingedCell}
          {participants}
          readOnly={true}
          onselect={() => {}}
          onvalue={() => {}}
          onerase={() => {}}
          ontogglenotes={() => {}}
          onhint={() => {}}
        />
      </section>

      <aside class="panel controls">
        <h2>Playback</h2>
        <div class="transport">
          <button
            type="button"
            onclick={() => seek(projection!.eventIndex - 1)}
            disabled={projection.eventIndex === 0}
          >
            Previous
          </button>
          <button class="button" type="button" onclick={togglePlayback}>
            {playing ? 'Pause' : 'Play'}
          </button>
          <button
            type="button"
            onclick={() => seek(projection!.eventIndex + 1)}
            disabled={projection.eventIndex === replay.events.length}>Next</button
          >
        </div>
        <label for="replay-position">Event {projection.eventIndex} of {replay.events.length}</label>
        <input
          id="replay-position"
          type="range"
          min="0"
          max={replay.events.length}
          value={projection.eventIndex}
          oninput={(event) => seek(Number(event.currentTarget.value))}
        />
        <fieldset>
          <legend>Speed</legend>
          {#each [0.5, 1, 2, 4] as value (value)}
            <button
              type="button"
              aria-pressed={speed === value}
              onclick={() => setSpeed(value as 0.5 | 1 | 2 | 4)}>{value}×</button
            >
          {/each}
        </fieldset>
        <h2>Players</h2>
        <ul class="players">
          {#each replay.participants as participant (participant.id)}
            <li>
              <span>{participants.get(participant.id)?.marker} {participant.name}</span>
              <span
                >{projection.connections[participant.id] === false
                  ? 'Disconnected'
                  : 'Connected'}</span
              >
            </li>
          {/each}
        </ul>
        <details>
          <summary>Event markers</summary>
          <ol class="events">
            {#each replay.events as event, index (event.eventNumber)}
              <li class:current={projection.eventIndex === index + 1}>
                <button type="button" onclick={() => seek(index + 1)}>
                  {event.eventNumber}. {event.type.replaceAll(/([A-Z])/g, ' $1').trim()}
                </button>
              </li>
            {/each}
          </ol>
        </details>
      </aside>
    </div>
    <div class="sr-only" aria-live="polite">
      Replay at event {projection.eventIndex}. {projection.completed ? 'Puzzle completed.' : ''}
    </div>
  {/if}
</main>

<style>
  .replay-shell {
    width: min(100% - 1.5rem, 72rem);
    margin-inline: auto;
    padding-block: var(--space-6) var(--space-10);
  }
  header {
    display: flex;
    align-items: start;
    justify-content: space-between;
    gap: var(--space-4);
    margin-bottom: var(--space-5);
  }
  h1 {
    margin: var(--space-1) 0;
  }
  .eyebrow {
    color: var(--brand-primary);
    font-weight: 800;
    margin: 0;
  }
  .replay-layout {
    display: grid;
    grid-template-columns: minmax(20rem, 42rem) minmax(15rem, 1fr);
    gap: var(--space-5);
    align-items: start;
  }
  .controls,
  .message {
    padding: var(--space-5);
  }
  .transport,
  fieldset {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-2);
    margin-block: var(--space-3);
  }
  input[type='range'] {
    width: 100%;
    min-height: 44px;
  }
  .players,
  .events {
    padding-left: var(--space-5);
  }
  .players li {
    display: flex;
    justify-content: space-between;
    gap: var(--space-2);
  }
  .events {
    max-height: 18rem;
    overflow: auto;
  }
  .events button {
    min-height: 2rem;
    text-align: left;
  }
  .events .current {
    font-weight: 800;
  }
  @media (max-width: 48rem) {
    header {
      display: grid;
    }
    .replay-layout {
      grid-template-columns: 1fr;
    }
  }
</style>
