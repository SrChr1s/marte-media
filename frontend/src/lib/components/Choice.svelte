<script lang="ts">
  // One-of-many picker. "segment" is a joined control for 2-3 options, "chips" a wrapping row for more.
  type Option = { value: string | number; label: string }
  let {
    options,
    value = $bindable(),
    variant = 'segment',
    label,
    onchange,
  }: { options: Option[]; value: string | number; variant?: 'segment' | 'chips'; label: string; onchange?: () => void } = $props()
</script>

<div class="choice {variant}" role="radiogroup" aria-label={label}>
  {#each options as o (o.value)}
    <button type="button" role="radio" aria-checked={value === o.value} class:on={value === o.value} onclick={() => {
        if (value === o.value) return
        value = o.value
        onchange?.()
      }}>
      {o.label}
    </button>
  {/each}
</div>

<style>
  .choice {
    display: flex;
    flex-wrap: wrap;
  }
  button {
    background: transparent;
    color: var(--text-muted);
    font-weight: 500;
    transition:
      background var(--dur) var(--ease),
      color var(--dur) var(--ease),
      box-shadow var(--dur) var(--ease),
      transform 120ms var(--ease);
  }
  button.on {
    color: #0f0f0f;
    background: var(--text);
  }
  button.on:hover:not(:disabled) {
    background: var(--text);
  }
  .segment {
    padding: 3px;
    gap: 2px;
    background: var(--surface-low);
    border: 1px solid var(--border);
    border-radius: var(--radius-button);
    width: fit-content;
  }
  .segment button {
    padding: 6px 16px;
  }
  .chips {
    gap: 8px;
  }
  .chips button {
    padding: 6px 14px;
    background: var(--surface-low);
    border: 1px solid var(--border);
    font-variant-numeric: tabular-nums;
  }
  .chips button.on {
    background: var(--text);
    border-color: var(--text);
    color: #0f0f0f;
  }
</style>
