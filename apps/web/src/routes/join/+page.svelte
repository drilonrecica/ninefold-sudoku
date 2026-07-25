<script lang="ts">
  let { data } = $props();
</script>

<svelte:head>
  <title>Join Room — Ninefold Sudoku</title>
  <meta name="robots" content="noindex, nofollow" />
</svelte:head>

<div class="shell narrow">
  <h1>Join a room</h1>
  {#if data.invalidCode}
    <p class="error-message" role="alert">Enter a valid six-character room code.</p>
  {/if}
  <form class="panel" method="get" action="/join">
    <div class="field">
      <label for="room-code">Room code</label>
      <input id="room-code" name="code" required autocomplete="off" maxlength="9" />
    </div>
    <button class="button" type="submit">Find Room</button>
  </form>
</div>

<style>
  .narrow {
    max-width: 38rem;
    padding-block: var(--space-12);
  }
  form {
    display: grid;
    gap: var(--space-5);
    padding: var(--space-6);
  }
</style>
