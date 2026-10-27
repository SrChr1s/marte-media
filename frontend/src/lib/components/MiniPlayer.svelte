<script lang="ts">
  import { duration } from '../format'
  import { openMenu, copyLink } from '../contextmenu.svelte'
  import { downloadTrack } from '../downloads'
  import { clear, current, player, seek, setVolume, toggleMute } from '../player.svelte'
  import Cover from './Cover.svelte'
  import Icon from './Icon.svelte'
  import QueueList from './QueueList.svelte'
  import Slider from './Slider.svelte'
  import Transport from './Transport.svelte'

  let { onopen }: { onopen: () => void } = $props()

  const t = $derived(current())
  let queueOpen = $state(false)
  let pop = $state<HTMLDivElement>()

  function outside(e: PointerEvent) {
    if (queueOpen && pop && !pop.contains(e.target as Node) && !(e.target as HTMLElement).closest('[data-queue-btn]')) queueOpen = false
  }
  $effect(() => {
    window.addEventListener('pointerdown', outside, true)
    return () => window.removeEventListener('pointerdown', outside, true)
  })

  function menu(e: MouseEvent) {
    if (!t) return
    openMenu(e, [
      { label: 'Descargar video', action: () => downloadTrack(t, 'video') },
      { label: 'Descargar audio', action: () => downloadTrack(t, 'audio') },
      copyLink(t.url),
    ])
  }
</script>

{#if player.queue.length > 0 && t}
  <section class="mini" aria-label="Reproductor">
    <div class="now">
      <button class="cover" onclick={onopen} title="Abrir el reproductor" aria-label="Abrir el reproductor">
        <Cover src={t.thumbnail} size={20} />
      </button>
      <div class="text">
        <span class="title clamp2" title={t.title}>{t.title}</span>
        {#if t.channel}<span class="muted ch">{t.channel}</span>{/if}
      </div>
    </div>

    <div class="center">
      <Transport />
      <div class="seek">
        <span class="num muted">{duration(player.time) || '0:00'}</span>
        <Slider label="Posición" value={player.time} max={player.duration || 1} buffered={player.buffered} onseek={seek} />
        <span class="num muted">{duration(player.duration) || '--:--'}</span>
      </div>
    </div>

    <div class="side">
      <div class="vol">
        <button class="icon" title={player.muted ? 'Activar el sonido' : 'Silenciar'} aria-label="Silenciar" onclick={toggleMute}>
          <Icon name={player.muted || player.volume === 0 ? 'volumeOff' : 'volume'} size={20} />
        </button>
        <div class="vslider"><Slider label="Volumen" value={player.muted ? 0 : player.volume} max={1} step={0.05} onseek={setVolume} /></div>
      </div>
      <button class="icon" title="Descargar" aria-label="Descargar" onclick={menu}><Icon name="download" size={20} /></button>
      <button class="icon" class:on={queueOpen} data-queue-btn title="Cola de reproducción" aria-label="Cola de reproducción" onclick={() => (queueOpen = !queueOpen)}>
        <Icon name="queueMusic" size={20} />
      </button>
      <button class="icon" title="Abrir en el reproductor" aria-label="Abrir en el reproductor" onclick={onopen}><Icon name="expand" size={18} /></button>
      <button class="icon" title="Cerrar el reproductor" aria-label="Cerrar el reproductor" onclick={clear}><Icon name="close" size={18} /></button>
    </div>

    {#if queueOpen}
      <div class="pop" bind:this={pop}>
        <h2>Cola <span class="muted num">{player.index + 1} de {player.queue.length}</span></h2>
        <div class="scroll"><QueueList /></div>
      </div>
    {/if}
  </section>
{/if}

<style>
  .mini {
    position: absolute;
    left: 16px;
    right: 16px;
    bottom: 14px;
    z-index: var(--z-player);
    display: grid;
    grid-template-columns: minmax(0, 1fr) minmax(260px, 1.5fr) minmax(0, 1fr);
    align-items: center;
    gap: 16px;
    padding: 10px 14px;
    background: var(--surface-low);
    border: 1px solid var(--border-strong);
    border-radius: var(--radius-panel);
    box-shadow: var(--shadow-menu);
  }
  .now {
    display: flex;
    align-items: center;
    gap: 12px;
    min-width: 0;
  }
  .cover {
    flex: none;
    width: 76px;
    aspect-ratio: 16 / 9;
    padding: 0;
    border-radius: 8px;
    overflow: hidden;
    background: var(--background);
  }
  .text {
    display: flex;
    flex-direction: column;
    min-width: 0;
  }
  .title {
    font-weight: 500;
    font-size: 13px;
    line-height: 1.25;
  }
  .ch {
    font-size: 12px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .center {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 2px;
    min-width: 0;
  }
  .seek {
    display: flex;
    align-items: center;
    gap: 10px;
    width: 100%;
  }
  .num {
    font-size: 11px;
    min-width: 34px;
    text-align: center;
  }
  .side {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    gap: 2px;
  }
  .icon {
    width: 34px;
    height: 34px;
  }
  .icon.on {
    color: var(--primary);
  }
  .vol {
    display: flex;
    align-items: center;
    gap: 2px;
    margin-right: 4px;
  }
  .vslider {
    width: 84px;
  }

  .pop {
    position: absolute;
    right: 0;
    bottom: calc(100% + 10px);
    width: min(420px, 100%);
    max-height: min(60vh, 460px);
    display: flex;
    flex-direction: column;
    padding: 12px;
    background: var(--surface-low);
    border: 1px solid var(--border-strong);
    border-radius: var(--radius-panel);
    box-shadow: var(--shadow-menu);
  }
  .pop h2 {
    font-size: 15px;
    padding: 2px 4px 8px;
  }
  .pop h2 span {
    font-family: var(--font);
    font-weight: 400;
    font-size: 12px;
    margin-left: 6px;
  }
  .scroll {
    overflow-y: auto;
  }

  @media (max-width: 900px) {
    .mini {
      grid-template-columns: minmax(0, 1fr) minmax(220px, 1.3fr) auto;
    }
    .vslider,
    .side .icon:nth-child(n + 4) {
      display: none;
    }
  }
</style>
