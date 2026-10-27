<script lang="ts">
  import { onMount } from 'svelte'
  import { cubicOut } from 'svelte/easing'
  import { fly } from 'svelte/transition'
  import ContextMenu from './lib/components/ContextMenu.svelte'
  import Icon from './lib/components/Icon.svelte'
  import Logo from './lib/components/Logo.svelte'
  import MiniPlayer from './lib/components/MiniPlayer.svelte'
  import Toasts from './lib/components/Toasts.svelte'
  import { LowGraphics } from '../wailsjs/go/main/App'
  import { engine, refreshEngine } from './lib/engine.svelte'
  import { initHistory } from './lib/history.svelte'
  import { player } from './lib/player.svelte'
  import { loadSettings } from './lib/settings.svelte'
  import { activeCount, initJobs } from './lib/jobs.svelte'
  import Download from './routes/Download.svelte'
  import Queue from './routes/Queue.svelte'
  import Reproductor from './routes/Reproductor.svelte'
  import History from './routes/History.svelte'
  import Settings from './routes/Settings.svelte'

  const nav = [
    { id: 'player', label: 'Reproducir', icon: 'playCircle' },
    { id: 'download', label: 'Descargar', icon: 'download' },
    { id: 'queue', label: 'Cola', icon: 'queue' },
    { id: 'history', label: 'Historial', icon: 'history' },
    { id: 'settings', label: 'Ajustes', icon: 'settings' },
  ]
  let view = $state('player') // the player is the home view

  // The mini player hides while the watching layout (with its own controls) is on screen.
  const watching = $derived(view === 'player' && player.stageOpen && player.queue.length > 0)
  const openPlayer = () => {
    view = 'player'
    player.stageOpen = true
  }

  const engineReady = $derived(!!engine.status?.installed && !!engine.status?.ffmpeg && !!engine.status?.jsRuntime)
  const engineTitle = $derived(
    !engine.status ? 'Comprobando el motor…' : engineReady ? `yt-dlp ${engine.status.version} · ffmpeg listo` : 'Falta yt-dlp, ffmpeg o el motor de JavaScript: revisa Ajustes',
  )

  onMount(() => {
    initJobs()
    refreshEngine()
    loadSettings()
    const offHistory = initHistory()
    LowGraphics().then((low) => low && document.documentElement.setAttribute('data-lowgfx', ''))

    // Pause decorative animations (CSS: html[data-paused]) whenever the window is not in front.
    const sync = () => document.documentElement.toggleAttribute('data-paused', document.hidden || !document.hasFocus())
    sync()
    window.addEventListener('focus', sync)
    window.addEventListener('blur', sync)
    document.addEventListener('visibilitychange', sync)
    return () => {
      window.removeEventListener('focus', sync)
      window.removeEventListener('blur', sync)
      document.removeEventListener('visibilitychange', sync)
      offHistory()
    }
  })
</script>

<div class="app">
  <nav class="rail" aria-label="Principal">
    <div class="brand"><Logo size={42} /></div>
    <div class="items">
      {#each nav as item}
        <button class="item" class:active={view === item.id} aria-current={view === item.id ? 'page' : undefined} onclick={() => (view = item.id)} title={item.label}>
          <span class="ico">
            <Icon name={item.icon} />
            {#if item.id === 'queue' && activeCount() > 0}<span class="badge num">{activeCount()}</span>{/if}
            {#if item.id === 'player' && player.status === 'playing' && view !== 'player'}<span class="live" title="Reproduciendo"></span>{/if}
          </span>
          <span class="label">{item.label}</span>
        </button>
      {/each}
    </div>
    <div class="foot" title={engineTitle}>
      <span class="dot" class:ok={engineReady} class:bad={engine.status && !engineReady}></span>
      <span>motor</span>
    </div>
  </nav>

  <div class="stage">
    <div class="limb" aria-hidden="true"></div>
    <main class:padded={player.queue.length > 0 && !watching}>
      {#key view}
        <div class="view" in:fly={{ y: 8, opacity: 1, duration: 200, easing: cubicOut }}>
          {#if view === 'download'}<Download onqueued={() => (view = 'queue')} />
          {:else if view === 'player'}<Reproductor />
          {:else if view === 'queue'}<Queue ongo={() => (view = 'download')} />
          {:else if view === 'history'}<History />
          {:else}<Settings />{/if}
        </div>
      {/key}
    </main>
    {#if !watching}<MiniPlayer onopen={openPlayer} />{/if}
  </div>
</div>
<ContextMenu />
<Toasts />

<style>
  .app {
    display: flex;
    height: 100%;
  }

  .rail {
    width: var(--rail-width);
    flex: none;
    display: flex;
    flex-direction: column;
    align-items: stretch;
    padding: 14px 8px 12px;
    background: var(--surface-low);
    border-right: 1px solid var(--border);
    z-index: var(--z-content);
  }
  .brand {
    display: grid;
    place-items: center;
    margin-bottom: 18px;
  }
  .items {
    display: flex;
    flex-direction: column;
    gap: 4px;
    flex: 1;
  }
  .item {
    position: relative;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 4px;
    padding: 11px 0 8px;
    border-radius: var(--radius-card);
    background: transparent;
    color: var(--text-muted);
    font-size: 10.5px;
    font-weight: 500;
  }
  .item:hover:not(:disabled) {
    background: rgba(255, 255, 255, 0.06);
    color: var(--text);
  }
  .item.active {
    background: var(--surface);
    color: var(--text);
  }
  /* The active section is marked by a red notch on the rail's edge. */
  .item::before {
    content: '';
    position: absolute;
    left: -8px;
    top: 22%;
    bottom: 22%;
    width: 3px;
    border-radius: 0 3px 3px 0;
    background: var(--primary);
    box-shadow: 0 0 12px var(--primary-glow);
    transform: scaleY(0);
    transition: transform 220ms var(--ease);
  }
  .item.active::before {
    transform: scaleY(1);
  }
  .ico {
    position: relative;
    display: inline-flex;
  }
  .badge {
    position: absolute;
    top: -7px;
    right: -12px;
    min-width: 17px;
    padding: 0 5px;
    border-radius: 9px;
    background: var(--primary);
    font-size: 10px;
    font-weight: 700;
    line-height: 17px;
    text-align: center;
    box-shadow: 0 2px 10px var(--primary-glow);
  }

  .live {
    position: absolute;
    top: -3px;
    right: -9px;
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--primary);
    box-shadow: 0 0 8px var(--primary-glow);
    animation: pulse 1.4s ease-in-out infinite;
  }
  @keyframes pulse {
    50% {
      opacity: 0.4;
    }
  }

  .foot {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 5px;
    padding-top: 10px;
    font-size: 10px;
    color: var(--text-muted);
  }
  .dot {
    width: 7px;
    height: 7px;
    border-radius: 50%;
    background: #6a6a6a;
  }
  .dot.ok {
    background: var(--ok);
    box-shadow: 0 0 8px rgba(43, 166, 64, 0.7);
  }
  .dot.bad {
    background: var(--primary);
    box-shadow: 0 0 8px var(--primary-glow);
  }

  .stage {
    position: relative;
    flex: 1;
    min-width: 0;
    overflow: hidden;
    background:
      radial-gradient(900px 480px at 88% -12%, rgba(255, 0, 0, 0.07), transparent 62%),
      var(--background);
  }
  /* Stars, faint and fixed. */
  .stage::before {
    content: '';
    position: absolute;
    inset: 0;
    opacity: 0.5;
    background-image:
      radial-gradient(1px 1px at 12% 22%, #fff, transparent),
      radial-gradient(1px 1px at 31% 8%, rgba(255, 255, 255, 0.7), transparent),
      radial-gradient(1.2px 1.2px at 57% 16%, #fff, transparent),
      radial-gradient(1px 1px at 76% 34%, rgba(255, 255, 255, 0.6), transparent),
      radial-gradient(1px 1px at 92% 12%, #fff, transparent),
      radial-gradient(1px 1px at 22% 62%, rgba(255, 255, 255, 0.5), transparent),
      radial-gradient(1.2px 1.2px at 68% 72%, rgba(255, 255, 255, 0.6), transparent),
      radial-gradient(1px 1px at 88% 58%, rgba(255, 255, 255, 0.5), transparent);
  }
  /* The curve of a planet along the bottom edge, lit from above. */
  .limb {
    position: absolute;
    left: 50%;
    top: calc(100% - 56px);
    width: 190%;
    aspect-ratio: 1;
    transform: translateX(-50%);
    border-radius: 50%;
    background: radial-gradient(60% 18% at 50% 0%, rgba(255, 0, 0, 0.16), transparent 70%), #120707;
    border-top: 1px solid rgba(255, 90, 70, 0.4);
    box-shadow:
      0 -26px 90px -18px rgba(255, 0, 0, 0.3),
      inset 0 24px 70px -26px rgba(255, 0, 0, 0.25);
  }

  main {
    position: relative;
    z-index: var(--z-content);
    height: 100%;
    overflow-x: hidden;
    overflow-y: auto;
    padding: 30px 44px 44px;
  }
  main.padded {
    padding-bottom: 130px; /* room for the mini player */
  }
  .view {
    max-width: 1080px;
  }
</style>
