<script lang="ts">
  import type { Snippet } from 'svelte'
  import Planet from './Planet.svelte'

  let { title, text, children }: { title: string; text: string; children?: Snippet } = $props()
</script>

<div class="empty">
  <div class="art"><Planet size={150} /></div>
  <h2>{title}</h2>
  <p class="muted">{text}</p>
  {#if children}<div class="actions">{@render children()}</div>{/if}
</div>

<style>
  .empty {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 6px;
    max-width: 380px;
    padding: 28px 0 0 8px;
  }
  .art {
    margin: 0 0 6px -20px;
  }
  h2 {
    font-size: 22px;
  }
  .actions {
    margin-top: 14px;
  }
</style>
