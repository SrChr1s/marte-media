<script lang="ts">
  import { cycleRepeat, next, player, prev, toggle, toggleShuffle } from '../player.svelte'
  import Icon from './Icon.svelte'
  import OrbitLoader from './OrbitLoader.svelte'

  let { big = false }: { big?: boolean } = $props()
  const playing = $derived(player.status === 'playing')
</script>

<div class="transport" class:big>
  <button class="icon" class:on={player.shuffle} title="Orden aleatorio" aria-label="Orden aleatorio" aria-pressed={player.shuffle} onclick={toggleShuffle}>
    <Icon name="shuffle" size={20} />
  </button>
  <button class="icon" title="Anterior" aria-label="Anterior" onclick={prev}><Icon name="prev" size={24} /></button>
  <button class="play" title={playing ? 'Pausar' : 'Reproducir'} aria-label={playing ? 'Pausar' : 'Reproducir'} onclick={toggle}>
    {#if player.status === 'loading'}
      <OrbitLoader size={22} />
    {:else}
      <Icon name={playing ? 'pause' : 'play'} size={big ? 30 : 24} />
    {/if}
  </button>
  <button class="icon" title="Siguiente" aria-label="Siguiente" onclick={() => next()}><Icon name="next" size={24} /></button>
  <button
    class="icon"
    class:on={player.repeat !== 'off'}
    title={player.repeat === 'one' ? 'Repetir esta' : player.repeat === 'all' ? 'Repetir todo' : 'No repetir'}
    aria-label="Repetir"
    onclick={cycleRepeat}
  >
    <Icon name={player.repeat === 'one' ? 'repeatOne' : 'repeat'} size={20} />
  </button>
</div>

<style>
  .transport {
    display: flex;
    align-items: center;
    gap: 6px;
  }
  .icon {
    width: 36px;
    height: 36px;
  }
  .icon.on {
    color: var(--primary);
  }
  .play {
    display: grid;
    place-items: center;
    width: 44px;
    height: 44px;
    padding: 0;
    border-radius: 50%;
    background: #fff;
    color: #0f0f0f;
  }
  .play:hover:not(:disabled) {
    background: #fff;
    transform: scale(1.06);
  }
  .big .play {
    width: 56px;
    height: 56px;
  }
</style>
