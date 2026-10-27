<script lang="ts">
  import { onDestroy, onMount } from 'svelte'
  import { cubicOut } from 'svelte/easing'
  import { fly } from 'svelte/transition'
  import { CancelProbe, Enqueue, InstallYtdlp, OpenFile, PickFolder, Probe } from '../../wailsjs/go/main/App'
  import { ClipboardGetText } from '../../wailsjs/runtime/runtime'
  import Icon from '../lib/components/Icon.svelte'
  import OrbitLoader from '../lib/components/OrbitLoader.svelte'
  import Options from '../lib/components/Options.svelte'
  import Planet from '../lib/components/Planet.svelte'
  import VideoCard from '../lib/components/VideoCard.svelte'
  import { duration } from '../lib/format'
  import { engine, refreshEngine } from '../lib/engine.svelte'
  import { history } from '../lib/history.svelte'
  import { settings } from '../lib/settings.svelte'
  import type { Entry, MediaInfo, Request } from '../lib/types'

  let { onqueued }: { onqueued: () => void } = $props()

  let url = $state('')
  let info = $state<MediaInfo | null>(null)
  let loading = $state(false)
  let error = $state('')
  let installing = $state(false)

  let mode = $state('video')
  let height = $state(1080)
  let videoFormat = $state('mp4')
  let audioFormat = $state('mp3')
  let audioBitrate = $state(192)
  let dir = $state('')
  let selected = $state(new Set<string>())

  const heights = $derived(info?.kind === 'video' && info.heights?.length ? info.heights : [2160, 1440, 1080, 720, 480, 360])
  const entries = $derived(info?.entries ?? [])
  const count = $derived(info?.kind === 'playlist' ? selected.size : 1)
  const allSelected = $derived(entries.length > 0 && selected.size === entries.length)
  const ready = $derived(!!engine.status?.installed)
  const idle = $derived(!info && !loading)
  const recent = $derived(history.items.slice(0, 3))

  // Defaults come from the shared settings; applied once they are available (usually already on mount).
  let defaultsApplied = false
  $effect(() => {
    const cfg = settings.cfg
    if (!cfg || defaultsApplied) return
    defaultsApplied = true
    ;({ mode, videoFormat, audioFormat, audioBitrate } = cfg)
    height = cfg.videoHeight
    dir = cfg.downloadDir
  })

  onMount(() => {
    if (!engine.status) refreshEngine()
  })

  async function install() {
    installing = true
    error = ''
    try {
      await InstallYtdlp()
      await refreshEngine()
    } catch (e) {
      error = String(e)
    } finally {
      installing = false
    }
  }

  // Each analysis gets a number; a result that arrives after cancel (or after a newer analysis) is dropped.
  let run = 0

  async function analyze() {
    const u = url.trim()
    if (!u) return
    const mine = ++run
    loading = true
    error = ''
    info = null
    try {
      const result = (await Probe(u)) as MediaInfo
      if (mine !== run) return
      info = result
      if (result.kind === 'playlist') selected = new Set((result.entries ?? []).map((e) => e.id))
      else if (!result.heights?.includes(height) && result.heights?.length) {
        // pick the best available height not above the preferred one
        height = result.heights.find((h) => h <= height) ?? result.heights[result.heights.length - 1]
      }
    } catch (e) {
      if (mine === run) error = String(e)
    } finally {
      if (mine === run) loading = false
    }
  }

  function cancelAnalysis() {
    if (!loading) return
    run++ // invalidate the pending result
    loading = false
    CancelProbe()
  }

  onDestroy(cancelAnalysis)

  function toggle(id: string) {
    const next = new Set(selected)
    if (!next.delete(id)) next.add(id)
    selected = next
  }

  function toggleAll() {
    selected = allSelected ? new Set() : new Set(entries.map((e) => e.id))
  }

  async function choose() {
    const d = await PickFolder(dir)
    if (d) dir = d
  }

  function toRequest(e: Entry): Request {
    return { url: e.url, title: e.title, channel: e.channel, thumbnail: e.thumbnail, mode, height, videoFormat, audioFormat, audioBitrate, dir }
  }

  async function download() {
    if (!info) return
    const list = info.kind === 'playlist' ? entries.filter((e) => selected.has(e.id)) : [info as Entry]
    if (!list.length) return
    await Enqueue(list.map(toRequest))
    info = null
    url = ''
    onqueued()
  }

  async function paste() {
    try {
      const text = (await ClipboardGetText()).trim()
      if (!text) return
      url = text
      analyze()
    } catch (e) {
      error = 'No se pudo leer el portapapeles: ' + String(e)
    }
  }
</script>

<div class="page" class:idle>
  {#if idle}
    <div class="planet" aria-hidden="true"><Planet size={430} float /></div>
    <header class="hero">
      <p class="eyebrow">marte-media</p>
      <h1>Descarga video y audio de YouTube</h1>
      <p class="lead">Pega el enlace de un video o de una playlist. Elige calidad, formato y carpeta; el resto lo hace la cola.</p>
    </header>
  {:else}
    <h1 class="slim">Descargar</h1>
  {/if}

  {#if engine.status && !engine.status.installed}
    <div class="notice" role="alert">
      <span class="ico"><Icon name="warning" size={18} /></span>
      <div class="msg">
        <strong>Falta yt-dlp</strong>
        <span class="muted">Hace falta para analizar y descargar. Se instala solo para esta app, sin permisos de administrador.</span>
      </div>
      <button class="primary" onclick={install} disabled={installing}>{installing ? 'Instalando…' : 'Instalar yt-dlp'}</button>
    </div>
  {/if}
  {#if engine.status && engine.status.installed && !engine.status.jsRuntime}
    <div class="notice" role="alert">
      <span class="ico"><Icon name="warning" size={18} /></span>
      <div class="msg">
        <strong>Falta un motor de JavaScript</strong>
        <span class="muted">yt-dlp lo usa para leer YouTube. Sin él puede recibir menos formatos o fallar. Instala deno, node o bun.</span>
      </div>
    </div>
  {/if}
  {#if engine.status && !engine.status.ffmpeg}
    <div class="notice" role="alert">
      <span class="ico"><Icon name="warning" size={18} /></span>
      <div class="msg">
        <strong>Falta ffmpeg</strong>
        <span class="muted">Sin él no se pueden unir audio y video ni convertir a MP3 u Opus.</span>
      </div>
    </div>
  {/if}

  <form class="bar" onsubmit={(e) => { e.preventDefault(); analyze() }}>
    <span class="link-ico"><Icon name="link" size={20} /></span>
    <input
      type="text"
      bind:value={url}
      placeholder="https://www.youtube.com/watch?v=…"
      aria-label="Enlace de YouTube"
      onkeydown={(e) => e.key === 'Escape' && cancelAnalysis()}
    />
    <button type="button" class="ghost paste" onclick={paste}><Icon name="paste" size={18} />Pegar</button>
    {#if loading}
      <button type="button" class="cancel" onclick={cancelAnalysis}>Cancelar</button>
    {:else}
      <button class="primary" type="submit" disabled={!url.trim() || !ready}>Analizar</button>
    {/if}
  </form>

  {#if error}
    <p class="error problem" role="alert"><Icon name="warning" size={16} />{error}</p>
  {/if}

  {#if idle}
    <p class="formats muted">
      <span>MP4</span><i></i><span>WebM</span><i></i><span>MP3</span><i></i><span>M4A</span><i></i><span>Opus</span><b>hasta 4K</b>
    </p>
  {/if}

  {#if idle && recent.length}
    <section class="recent" aria-label="Descargas recientes">
      <h2>Recientes</h2>
      <ul>
        {#each recent as r (r.id)}
          <li>
            <button class="item" title="Abrir {r.title}" onclick={() => OpenFile(r.path)}>
              <span class="rthumb">
                {#if r.thumbnail}<img src={r.thumbnail} alt="" />{/if}
                <span class="rfmt">{r.format.toUpperCase()}</span>
              </span>
              <span class="rtitle clamp2">{r.title}</span>
            </button>
          </li>
        {/each}
      </ul>
    </section>
  {/if}

  {#if loading}
    <p class="working muted" role="status"><OrbitLoader size={22} />Analizando el enlace… <span>Esc para cancelar</span></p>
    <div class="result single" aria-busy="true">
      <div class="media">
        <div class="skeleton thumb"></div>
        <div class="skeleton line" style="width: 78%"></div>
        <div class="skeleton line" style="width: 42%"></div>
      </div>
      <div class="panel">
        <div class="skeleton line" style="width: 36%"></div>
        <div class="skeleton block"></div>
        <div class="skeleton line" style="width: 30%"></div>
        <div class="skeleton block"></div>
        <div class="skeleton btn"></div>
      </div>
    </div>
  {/if}

  {#if info}
    <section class="result" class:single={info.kind === 'video'} in:fly={{ y: 12, duration: 260, easing: cubicOut }}>
      {#if info.kind === 'video'}
        <article class="media">
          <div class="thumb">
            {#if info.thumbnail}<img src={info.thumbnail} alt="" />{/if}
            {#if info.duration}<span class="dur num">{duration(info.duration)}</span>{/if}
          </div>
          <h2 class="clamp2" title={info.title}>{info.title}</h2>
          <p class="muted">{info.channel}</p>
        </article>
        <aside class="panel">
          <Options
            bind:mode
            bind:height
            bind:videoFormat
            bind:audioFormat
            bind:audioBitrate
            {heights}
            {dir}
            {count}
            onchoose={choose}
            ondownload={download}
          />
        </aside>
      {:else}
        <header class="plhead">
          <div>
            <h2 class="clamp2">{info.title}</h2>
            <p class="muted num">{info.playlistCount} videos · {selected.size} seleccionados</p>
          </div>
          <button class="ghost" onclick={toggleAll}>{allSelected ? 'Quitar todo' : 'Seleccionar todo'}</button>
        </header>
        <div class="grid">
          {#each entries as e, i (e.id)}
            <div in:fly={{ y: 8, opacity: 0.4, duration: i < 12 ? 200 : 0, delay: Math.min(i, 12) * 20, easing: cubicOut }}>
              <VideoCard entry={e} selectable selected={selected.has(e.id)} onToggle={() => toggle(e.id)} />
            </div>
          {/each}
        </div>
        <div class="dock">
          <Options
            compact
            bind:mode
            bind:height
            bind:videoFormat
            bind:audioFormat
            bind:audioBitrate
            {heights}
            {dir}
            {count}
            onchoose={choose}
            ondownload={download}
          />
        </div>
      {/if}
    </section>
  {/if}
</div>

<style>
  .page {
    position: relative;
  }
  /* Hero: the planet bleeds off the right edge behind the copy. */
  .planet {
    position: absolute;
    right: -130px;
    top: -22px;
    z-index: 0;
    pointer-events: none;
  }
  .hero {
    position: relative;
    max-width: 560px;
    margin: 26px 0 30px;
    z-index: 1;
  }
  .eyebrow {
    font-family: var(--font-display);
    font-size: 13px;
    font-weight: 500;
    letter-spacing: 0.18em;
    text-transform: lowercase;
    color: var(--primary);
    margin-bottom: 14px;
  }
  .hero h1 {
    font-size: 44px;
    line-height: 1.04;
    font-weight: 700;
    letter-spacing: -0.035em;
  }
  .lead {
    margin-top: 16px;
    max-width: 46ch;
    font-size: 15px;
    color: var(--text-muted);
  }
  .slim {
    font-size: 22px;
    margin-bottom: 18px;
  }

  .notice {
    display: flex;
    align-items: center;
    gap: 14px;
    max-width: 700px;
    margin-bottom: 12px;
    padding: 12px 14px;
    background: var(--surface-low);
    border: 1px solid var(--border-strong);
    border-left: 3px solid var(--primary);
    border-radius: var(--radius-card);
  }
  .notice .ico {
    color: var(--primary);
    display: grid;
  }
  .msg {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }

  .page.idle .bar,
  .page.idle .formats,
  .page.idle .recent {
    max-width: 560px;
  }
  .bar {
    position: relative;
    display: flex;
    align-items: center;
    gap: 6px;
    max-width: 700px;
    padding: 7px 7px 7px 16px;
    background: var(--surface-low);
    border: 1px solid var(--border-strong);
    border-radius: 999px;
    box-shadow: var(--shadow-card);
    transition:
      border-color var(--dur) var(--ease),
      box-shadow var(--dur) var(--ease);
    z-index: 1;
  }
  .bar:focus-within {
    border-color: rgba(255, 255, 255, 0.55);
    box-shadow:
      0 0 0 4px rgba(255, 255, 255, 0.07),
      var(--shadow-card);
  }
  .link-ico {
    display: grid;
    color: var(--text-muted);
  }
  .bar input {
    flex: 1;
    min-width: 0;
    padding: 9px 6px;
    border: 0;
    border-radius: 0;
    background: transparent;
    box-shadow: none;
    font-size: 15px;
  }
  .bar input:focus {
    box-shadow: none;
    border: 0;
  }
  .paste {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 8px 12px;
    border-radius: 999px;
  }
  .bar :global(button.primary),
  .cancel {
    padding: 10px 22px;
    border-radius: 999px;
  }

  .problem {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-top: 14px;
    max-width: 700px;
  }

  .formats {
    display: flex;
    align-items: center;
    gap: 10px;
    margin-top: 16px;
    font-size: 12px;
    font-weight: 500;
    letter-spacing: 0.04em;
  }
  .formats i {
    width: 3px;
    height: 3px;
    border-radius: 50%;
    background: var(--text-muted);
    opacity: 0.6;
  }
  .formats b {
    margin-left: 10px;
    padding: 2px 8px;
    border: 1px solid var(--border-strong);
    border-radius: 6px;
    font-weight: 500;
  }

  .recent {
    position: relative;
    margin-top: 34px;
  }
  .recent h2 {
    margin-bottom: 12px;
    font-size: 13px;
    font-weight: 500;
    letter-spacing: 0.04em;
    color: var(--text-muted);
  }
  .recent ul {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 14px;
  }
  .recent .item {
    display: flex;
    flex-direction: column;
    align-items: stretch;
    gap: 8px;
    width: 100%;
    padding: 0;
    background: transparent;
    text-align: left;
    font-weight: 500;
    line-height: 1.3;
  }
  .recent .item:hover:not(:disabled) {
    background: transparent;
  }
  .rthumb {
    position: relative;
    aspect-ratio: 16 / 9;
    background: var(--surface-low);
    border-radius: var(--radius-card);
    overflow: hidden;
    transition:
      transform var(--dur) var(--ease),
      box-shadow var(--dur) var(--ease);
  }
  .item:hover .rthumb {
    transform: translateY(-2px);
    box-shadow: var(--shadow-card);
  }
  .rthumb img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
  }
  .rfmt {
    position: absolute;
    right: 5px;
    bottom: 5px;
    padding: 1px 5px;
    border-radius: 5px;
    background: rgba(0, 0, 0, 0.8);
    font-size: 10px;
    font-weight: 700;
    letter-spacing: 0.06em;
  }
  .rtitle {
    font-size: 12.5px;
  }

  .working {
    display: flex;
    align-items: center;
    gap: 10px;
    margin: 18px 0 4px;
  }
  .working span {
    opacity: 0.7;
    font-size: 12px;
  }

  .result {
    margin-top: 28px;
  }
  .result.single {
    display: grid;
    grid-template-columns: minmax(0, 1.25fr) minmax(300px, 1fr);
    gap: 32px;
    align-items: start;
  }
  .media {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }
  .thumb {
    position: relative;
    aspect-ratio: 16 / 9;
    background: var(--surface-low);
    border-radius: var(--radius-panel);
    overflow: hidden;
    box-shadow: var(--shadow-card);
  }
  .thumb img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
  }
  .dur {
    position: absolute;
    right: 10px;
    bottom: 10px;
    padding: 2px 7px;
    border-radius: 6px;
    background: rgba(0, 0, 0, 0.82);
    font-size: 12px;
    font-weight: 500;
  }
  .media h2 {
    font-size: 21px;
    line-height: 1.2;
    margin-top: 6px;
  }

  .panel {
    padding: 22px;
    background: var(--surface-low);
    border: 1px solid var(--border);
    border-radius: var(--radius-panel);
    display: flex;
    flex-direction: column;
    gap: 16px;
  }

  .skeleton.thumb {
    border-radius: var(--radius-panel);
  }
  .skeleton.line {
    height: 14px;
    border-radius: 7px;
  }
  .skeleton.block {
    height: 36px;
    border-radius: var(--radius-button);
  }
  .skeleton.btn {
    height: 46px;
    border-radius: var(--radius-button);
    margin-top: 8px;
    background: var(--surface);
  }

  .plhead {
    display: flex;
    justify-content: space-between;
    align-items: flex-end;
    gap: 16px;
    margin-bottom: 18px;
  }
  .plhead h2 {
    font-size: 22px;
  }
  .grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(210px, 1fr));
    gap: 22px 16px;
    padding: 4px 4px 8px;
  }
  .dock {
    position: sticky;
    bottom: 4px;
    z-index: 2;
    margin-top: 22px;
    padding: 16px 20px;
    background: rgb(23, 23, 23);
    border: 1px solid var(--border-strong);
    border-radius: var(--radius-panel);
    box-shadow: var(--shadow-menu);
  }

  @media (max-width: 900px) {
    .result.single {
      grid-template-columns: 1fr;
    }
    .planet {
      opacity: 0.45;
    }
  }
</style>
