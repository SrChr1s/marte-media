<script lang="ts">
  import { onDestroy, onMount } from 'svelte'
  import { WindowFullscreen, WindowUnfullscreen } from '../../../wailsjs/runtime/runtime'
  import { duration } from '../format'
  import { openMenu, type MenuItem } from '../contextmenu.svelte'
  import { current, next, player, seek, setMode, setQuality, setRate, setVolume, stage as attachVideo, toggle, toggleMute } from '../player.svelte'
  import Cover from './Cover.svelte'
  import Icon from './Icon.svelte'
  import OrbitLoader from './OrbitLoader.svelte'
  import Slider from './Slider.svelte'
  import Transport from './Transport.svelte'

  const t = $derived(current())
  let full = $state(false)
  let shown = $state(true) // controls visible
  let idle: ReturnType<typeof setTimeout> | undefined
  let box = $state<HTMLDivElement>()

  const showVideo = $derived(player.mode === 'video' && player.hasVideo)
  const quality = $derived(player.stream?.video?.height ?? 0)

  function wake() {
    shown = true
    clearTimeout(idle)
    if (player.status === 'playing') idle = setTimeout(() => (shown = false), 2600)
  }
  $effect(() => {
    // paused, loading or failed: the controls stay; playing: they hide after a while
    if (player.status !== 'playing') {
      shown = true
      clearTimeout(idle)
    } else wake()
  })

  async function setFull(v: boolean) {
    full = v
    try {
      if (v) WindowFullscreen()
      else WindowUnfullscreen()
    } catch {
      /* the page still fills the window */
    }
  }

  function key(e: KeyboardEvent) {
    const el = e.target as HTMLElement
    if (el.closest('input, textarea') || el.getAttribute('role') === 'slider') return
    if (e.key === 'Escape' && full) return void setFull(false)
    if (e.key === 'f' || e.key === 'F') return void setFull(!full)
    if (e.key === 'm' || e.key === 'M') return toggleMute()
    if (e.key === ' ' && !el.closest('button')) {
      e.preventDefault()
      toggle()
    } else if (e.key === 'ArrowRight' && !el.closest('button')) seek(player.time + 5)
    else if (e.key === 'ArrowLeft' && !el.closest('button')) seek(player.time - 5)
  }
  onMount(() => {
    window.addEventListener('keydown', key)
    return () => window.removeEventListener('keydown', key)
  })
  onDestroy(() => {
    clearTimeout(idle)
    if (full) setFull(false)
  })

  const label = (h: number) => (h >= 2160 ? '4K (2160p)' : `${h}p`)
  function qualityMenu(e: MouseEvent) {
    const items: MenuItem[] = player.heights.map((h) => ({ label: label(h), hint: h === quality ? '✓' : '', action: () => setQuality(h) }))
    if (items.length) openMenu(e, items)
  }
  function speedMenu(e: MouseEvent) {
    openMenu(
      e,
      [0.5, 0.75, 1, 1.25, 1.5, 2].map((r) => ({ label: r === 1 ? 'Normal' : `${r}×`, hint: r === player.rate ? '✓' : '', action: () => setRate(r) })),
    )
  }
</script>

<div
  bind:this={box}
  class="stage"
  class:full
  class:hide={!shown && player.status === 'playing'}
  role="group"
  aria-label="Reproductor"
  tabindex="-1"
  onmousemove={wake}
>
  {#if showVideo}
    <!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
    <div class="surface" use:attachVideo onclick={toggle} ondblclick={() => setFull(!full)}></div>
  {:else}
    <div class="cover">
      <Cover src={t?.thumbnail} icon="music" size={56} />
      <div class="shade"></div>
    </div>
  {/if}

  {#if player.status === 'loading'}
    <div class="center"><OrbitLoader size={46} /></div>
  {:else if player.status === 'error'}
    <div class="center err" role="alert">
      <Icon name="warning" size={28} />
      <p>{player.error || 'No se pudo reproducir.'}</p>
      <button class="primary" onclick={() => next()}>Siguiente</button>
    </div>
  {/if}

  <div class="controls">
    <div class="seekrow">
      <span class="num">{duration(player.time) || '0:00'}</span>
      <Slider label="Posición" thick value={player.time} max={player.duration || 1} buffered={player.buffered} onseek={seek} />
      <span class="num">{duration(player.duration) || '--:--'}</span>
    </div>
    <div class="row">
      <div class="left">
        <Transport />
        <button class="icon" title={player.muted ? 'Activar el sonido' : 'Silenciar'} aria-label="Silenciar" onclick={toggleMute}>
          <Icon name={player.muted || player.volume === 0 ? 'volumeOff' : 'volume'} size={20} />
        </button>
        <div class="vol"><Slider label="Volumen" value={player.muted ? 0 : player.volume} max={1} step={0.05} onseek={setVolume} /></div>
      </div>
      <div class="right">
        <button class="chip" title="Velocidad" onclick={speedMenu}>{player.rate === 1 ? '1×' : `${player.rate}×`}</button>
        {#if showVideo && player.heights.length > 1}
          <button class="chip" title="Calidad" onclick={qualityMenu}>{quality ? label(quality) : 'Auto'}</button>
        {/if}
        <button class="chip" title={player.mode === 'video' ? 'Reproducir solo el audio' : 'Ver el video'} onclick={() => setMode(player.mode === 'video' ? 'audio' : 'video')}>
          {player.mode === 'video' ? 'Solo audio' : 'Ver video'}
        </button>
        <button class="icon" title={full ? 'Salir de pantalla completa (F)' : 'Pantalla completa (F)'} aria-label="Pantalla completa" onclick={() => setFull(!full)}>
          <Icon name={full ? 'fullscreenExit' : 'fullscreen'} size={22} />
        </button>
      </div>
    </div>
  </div>
</div>

<style>
  .stage {
    position: relative;
    aspect-ratio: 16 / 9;
    background: #000;
    border-radius: var(--radius-panel);
    overflow: hidden;
    box-shadow: var(--shadow-card);
    outline: none;
  }
  .stage.full {
    position: fixed;
    inset: 0;
    z-index: var(--z-full);
    aspect-ratio: auto;
    border-radius: 0;
  }
  .stage.hide {
    cursor: none;
  }
  .surface {
    position: absolute;
    inset: 0;
  }
  .cover {
    position: absolute;
    inset: 0;
    display: grid;
    place-items: center;
    background: var(--surface-low);
  }
  .shade {
    position: absolute;
    inset: 0;
    background: linear-gradient(180deg, rgba(0, 0, 0, 0.15), rgba(0, 0, 0, 0.55));
  }
  .center {
    position: absolute;
    inset: 0;
    display: grid;
    place-content: center;
    justify-items: center;
    gap: 12px;
    pointer-events: none;
  }
  .center.err {
    pointer-events: auto;
    background: rgba(0, 0, 0, 0.7);
    color: var(--danger);
    text-align: center;
    padding: 24px;
  }
  .center.err p {
    max-width: 44ch;
    color: var(--text);
  }

  .controls {
    position: absolute;
    left: 0;
    right: 0;
    bottom: 0;
    padding: 36px 14px 10px;
    background: linear-gradient(180deg, transparent, rgba(0, 0, 0, 0.82));
    transition: opacity 200ms var(--ease);
  }
  .stage.hide .controls {
    opacity: 0;
    pointer-events: none;
  }
  .seekrow {
    display: flex;
    align-items: center;
    gap: 10px;
    font-size: 12px;
  }
  .num {
    min-width: 36px;
    text-align: center;
    font-variant-numeric: tabular-nums;
  }
  .row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 12px;
    margin-top: 2px;
  }
  .left,
  .right {
    display: flex;
    align-items: center;
    gap: 4px;
  }
  .vol {
    width: 80px;
    margin-left: 2px;
  }
  .icon {
    width: 36px;
    height: 36px;
  }
  .chip {
    padding: 5px 11px;
    font-size: 12px;
    background: rgba(255, 255, 255, 0.12);
  }
  .chip:hover:not(:disabled) {
    background: rgba(255, 255, 255, 0.22);
  }
</style>
