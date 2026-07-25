<script lang="ts">
  let roomCode = $state('');
  let room: Record<string, unknown> | null = $state(null);
  let status: Record<string, unknown> | null = $state(null);
  let message = $state('');
  let reason = $state('');
  let replayID = $state('');
  let puzzleID = $state('');
  let puzzleRevision = $state(1);

  async function loadHealth() {
    status = await request('/health/status');
  }

  async function lookupRoom() {
    room = await request(
      `/internal/admin/rooms/${encodeURIComponent(roomCode.trim().toUpperCase())}`,
    );
  }

  async function terminateRoom() {
    if (!confirm('Terminate this room? Players will be disconnected and cannot resume.')) return;
    await runMutation(
      `/internal/admin/rooms/${encodeURIComponent(roomCode.trim().toUpperCase())}/terminate`,
      'POST',
      'Room terminated.',
    );
    if (!message.startsWith('Operation failed')) await lookupRoom();
  }

  async function mutate(path: string, method: string) {
    const response = await fetch(path, {
      method,
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ confirm: true, reason, requestId: uuidv7() }),
    });
    if (!response.ok) throw new Error(`Operation failed (${response.status}).`);
  }

  async function deleteReplay() {
    if (!confirm('Delete this shared replay for everyone?')) return;
    await runMutation(
      `/internal/admin/replays/${encodeURIComponent(replayID.trim())}`,
      'DELETE',
      'Replay deleted.',
    );
  }

  async function retirePuzzle() {
    if (!confirm('Retire this puzzle revision from future assignments?')) return;
    await runMutation(
      `/internal/admin/puzzles/${encodeURIComponent(puzzleID.trim())}/${puzzleRevision}/retire`,
      'POST',
      'Puzzle retired.',
    );
  }

  async function runMutation(path: string, method: string, success: string) {
    message = '';
    try {
      await mutate(path, method);
      message = success;
    } catch (error) {
      message = error instanceof Error ? error.message : 'Operation failed.';
    }
  }

  async function request(path: string) {
    message = '';
    const response = await fetch(path, { headers: { Accept: 'application/json' } });
    if (!response.ok) {
      message = 'The protected operation failed. Verify private-network access and the target.';
      return null;
    }
    return response.json();
  }

  function uuidv7() {
    const bytes = crypto.getRandomValues(new Uint8Array(16));
    let timestamp = Date.now();
    for (let index = 5; index >= 0; index--) {
      bytes[index] = timestamp & 0xff;
      timestamp = Math.floor(timestamp / 256);
    }
    bytes[6] = ((bytes[6] ?? 0) & 0x0f) | 0x70;
    bytes[8] = ((bytes[8] ?? 0) & 0x3f) | 0x80;
    const hex = [...bytes].map((value) => value.toString(16).padStart(2, '0')).join('');
    return `${hex.slice(0, 8)}-${hex.slice(8, 12)}-${hex.slice(12, 16)}-${hex.slice(16, 20)}-${hex.slice(20)}`;
  }
</script>

<svelte:head>
  <title>Private Operations — Ninefold Sudoku</title>
  <meta name="robots" content="noindex, nofollow" />
</svelte:head>

<div class="shell admin">
  <h1>Private operations</h1>
  <p>This page is available only through the trusted private proxy.</p>
  {#if message}<p role="alert">{message}</p>{/if}

  <section class="panel">
    <h2>Health</h2>
    <button class="button secondary" type="button" onclick={loadHealth}>Load health</button>
    {#if status}<pre>{JSON.stringify(status, null, 2)}</pre>{/if}
  </section>

  <section class="panel">
    <h2>Room lookup</h2>
    <label>Room code <input bind:value={roomCode} maxlength="6" autocomplete="off" /></label>
    <button class="button secondary" type="button" onclick={lookupRoom}>Look up room</button>
    {#if room}
      <pre>{JSON.stringify(room, null, 2)}</pre>
      <label>Audit reason <input bind:value={reason} maxlength="200" /></label>
      <button class="button danger" type="button" onclick={terminateRoom}>Terminate room</button>
    {/if}
  </section>

  <section class="panel">
    <h2>Replay deletion</h2>
    <label>Replay ID <input bind:value={replayID} autocomplete="off" /></label>
    <label>Audit reason <input bind:value={reason} maxlength="200" /></label>
    <button class="button danger" type="button" onclick={deleteReplay}>Delete replay</button>
  </section>

  <section class="panel">
    <h2>Puzzle retirement</h2>
    <label>Puzzle ID <input bind:value={puzzleID} autocomplete="off" /></label>
    <label>Revision <input type="number" min="1" bind:value={puzzleRevision} /></label>
    <label>Audit reason <input bind:value={reason} maxlength="200" /></label>
    <button class="button danger" type="button" onclick={retirePuzzle}>Retire puzzle</button>
  </section>
</div>

<style>
  .admin {
    max-width: 52rem;
    padding-block: var(--space-10);
  }
  section {
    display: grid;
    gap: var(--space-4);
    margin-top: var(--space-6);
    padding: var(--space-5);
  }
  label {
    display: grid;
    gap: var(--space-2);
  }
  input {
    min-height: 44px;
    border: 1px solid var(--border-strong);
    padding: var(--space-2);
  }
  pre {
    overflow: auto;
    white-space: pre-wrap;
  }
  .danger {
    background: var(--state-error);
  }
</style>
