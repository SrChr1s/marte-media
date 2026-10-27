<script lang="ts">
  import { duration } from '../format'
  import Icon from './Icon.svelte'
  import type { Entry } from '../types'

  let {
    entry,
    selectable = false,
    selected = false,
    onToggle,
  }: { entry: Entry; selectable?: boolean; selected?: boolean; onToggle?: () => void } = $props()
</script>

{#snippet body()}
  <div class="thumb">
    {#if entry.thumbnail}<img src={entry.thumbnail} alt="" loading="lazy" />{/if}
    {#if entry.duration}<span class="dur num">{duration(entry.duration)}</span>{/if}
    {#if selectable}<span class="check">{#if selected}<Icon name="check" size={14} />{/if}</span>{/if}
  </div>
  <div class="meta">
    <div class="title clamp2" title={entry.title}>{entry.title}</div>
    {#if entry.channel}<div class="muted channel">{entry.channel}</div>{/if}
  </div>
{/snippet}

{#if selectable}
  <div
    class="card selectable"
    class:selected
    role="checkbox"
    aria-checked={selected}
    tabindex="0"
    onclick={onToggle}
    onkeydown={(e) => e.key === ' ' && (e.preventDefault(), onToggle?.())}
  >
    {@render body()}
  </div>
{:else}
  <div class="card">{@render body()}</div>
{/if}

<style>
  .card {
    display: flex;
    flex-direction: column;
    gap: 10px;
    transition: transform var(--dur) var(--ease);
  }
  .selectable {
    cursor: pointer;
  }
  .selectable:hover {
    transform: translateY(-2px);
  }
  .thumb {
    position: relative;
    aspect-ratio: 16 / 9;
    background: var(--surface-low);
    border-radius: var(--radius-card);
    overflow: hidden;
    transition:
      box-shadow var(--dur) var(--ease),
      opacity var(--dur) var(--ease);
  }
  .selectable:hover .thumb {
    box-shadow: var(--shadow-card);
  }
  .selectable:not(.selected) .thumb {
    opacity: 0.55;
  }
  .selectable:not(.selected):hover .thumb {
    opacity: 0.85;
  }
  .selected .thumb {
    box-shadow:
      0 0 0 2px var(--primary),
      var(--shadow-card);
  }
  img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
  }
  .dur {
    position: absolute;
    right: 6px;
    bottom: 6px;
    background: rgba(0, 0, 0, 0.82);
    border-radius: 4px;
    padding: 1px 5px;
    font-size: 12px;
    font-weight: 500;
  }
  .check {
    position: absolute;
    left: 8px;
    top: 8px;
    width: 22px;
    height: 22px;
    border-radius: 50%;
    background: rgba(0, 0, 0, 0.6);
    border: 2px solid rgba(255, 255, 255, 0.9);
    display: grid;
    place-items: center;
    transition:
      background var(--dur) var(--ease),
      border-color var(--dur) var(--ease);
  }
  .selected .check {
    background: var(--primary);
    border-color: var(--primary);
  }
  .title {
    font-weight: 500;
    line-height: 1.3;
  }
  .channel {
    margin-top: 2px;
    font-size: 12px;
  }
</style>
