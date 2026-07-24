<script lang="ts">
  import type { Digit } from '$lib/game/types';

  let { notes }: { notes: Digit[] } = $props();
</script>

<span class="candidates" aria-hidden="true">
  {#each [1, 2, 3, 4, 5, 6, 7, 8, 9] as digit (digit)}
    <span>{notes.includes(digit as Digit) ? digit : ''}</span>
  {/each}
</span>

<style>
  .candidates {
    display: grid;
    width: 100%;
    height: 100%;
    grid-template-columns: repeat(3, 1fr);
    grid-template-rows: repeat(3, 1fr);
    align-items: center;
    font-size: clamp(0.42rem, 1.4vw, 0.68rem);
    font-weight: 650;
    line-height: 1;
  }
  .candidates span {
    text-align: center;
  }
</style>
