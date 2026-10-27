<script lang="ts">
  let { percent, state = 'active' }: { percent: number; state?: 'active' | 'done' | 'error' } = $props()
  const p = $derived(Math.max(0, Math.min(100, percent)) / 100)
</script>

<div class="track" role="progressbar" aria-valuemin="0" aria-valuemax="100" aria-valuenow={Math.round(p * 100)}>
  <div class="fill {state}" style="transform: scaleX({p})"></div>
</div>

<style>
  .track {
    height: 6px;
    background: var(--surface-low);
    border: 1px solid var(--border);
    border-radius: 3px;
    overflow: hidden;
  }
  .fill {
    height: 100%;
    transform-origin: left;
    background: linear-gradient(90deg, #7a0000, #f00);
    box-shadow: 0 0 12px var(--primary-glow);
    transition: transform 0.25s linear;
  }
  .fill.done {
    background: var(--ok);
    box-shadow: none;
  }
  .fill.error {
    background: #5a2323;
    box-shadow: none;
  }
</style>
