<script lang="ts">
  // A thumbnail that never shows a broken-image glyph: if it fails to load, a quiet placeholder takes its place.
  import Icon from './Icon.svelte'

  let { src, icon = 'film', size = 24 }: { src?: string; icon?: string; size?: number } = $props()
  let failed = $state(false)
  $effect(() => {
    void src
    failed = false
  })
</script>

{#if src && !failed}
  <img {src} alt="" loading="lazy" onerror={() => (failed = true)} />
{:else}
  <span class="ph"><Icon name={icon} {size} /></span>
{/if}

<style>
  img,
  .ph {
    width: 100%;
    height: 100%;
    display: block;
  }
  img {
    object-fit: cover;
  }
  .ph {
    display: grid;
    place-items: center;
    background: var(--surface-low);
    color: var(--text-muted);
    opacity: 0.8;
  }
</style>
