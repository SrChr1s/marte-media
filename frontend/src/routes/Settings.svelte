<script lang="ts">
  import { onMount } from 'svelte'
  import { InstallYtdlp, PickFolder, SaveSettings, UpdateYtdlp } from '../../wailsjs/go/main/App'
  import Choice from '../lib/components/Choice.svelte'
  import Icon from '../lib/components/Icon.svelte'
  import { engine, refreshEngine } from '../lib/engine.svelte'
  import { persist, player, setAutoplay } from '../lib/player.svelte'
  import { settings } from '../lib/settings.svelte'
  import type { Settings } from '../lib/types'

  const cfg = $derived(settings.cfg)
  let msg = $state('')
  let busy = $state(false)

  onMount(refreshEngine)

  // Persist on every change; the file is tiny.
  $effect(() => {
    if (cfg) SaveSettings($state.snapshot(cfg) as Settings)
  })

  async function choose() {
    const d = await PickFolder(cfg!.downloadDir)
    if (d && settings.cfg) settings.cfg.downloadDir = d
  }

  async function run(fn: () => Promise<unknown>) {
    busy = true
    msg = ''
    try {
      const out = await fn()
      if (typeof out === 'string') msg = out
      await refreshEngine()
    } catch (e) {
      msg = String(e)
    } finally {
      busy = false
    }
  }

  // the player remembers its own choices; save them whenever one changes
  let auto = $state(player.autoplay ? 'on' : 'off')
  $effect(() => setAutoplay(auto === 'on'))
  $effect(() => {
    void player.height
    persist()
  })

  const heights = [2160, 1440, 1080, 720, 480, 360].map((h) => ({ value: h, label: h >= 2160 ? '4K' : `${h}p` }))
  const st = $derived(engine.status)
</script>

<h1>Ajustes</h1>

{#if cfg}
  <section class="panel">
    <h2>Descargas</h2>
    <div class="row">
      <div class="what"><span>Carpeta por defecto</span><small class="muted">Donde se guardan los archivos.</small></div>
      <button class="path" onclick={choose} title={cfg.downloadDir}><Icon name="folder" size={18} /><span>{cfg.downloadDir}</span></button>
    </div>
    <div class="row">
      <div class="what"><span>Descargas simultáneas</span><small class="muted">Más de una a la vez consume más red.</small></div>
      <div class="stepper" role="group" aria-label="Descargas simultáneas">
        <button class="icon" aria-label="Menos" disabled={cfg.maxConcurrent <= 1} onclick={() => cfg && (cfg.maxConcurrent -= 1)}>−</button>
        <span class="num">{cfg.maxConcurrent}</span>
        <button class="icon" aria-label="Más" disabled={cfg.maxConcurrent >= 6} onclick={() => cfg && (cfg.maxConcurrent += 1)}>+</button>
      </div>
    </div>
  </section>

  <section class="panel">
    <h2>Valores por defecto</h2>
    <div class="row">
      <div class="what"><span>Tipo</span></div>
      <Choice label="Tipo por defecto" bind:value={cfg.mode} options={[{ value: 'video', label: 'Video' }, { value: 'audio', label: 'Solo audio' }]} />
    </div>
    <div class="row">
      <div class="what"><span>Calidad de video</span><small class="muted">Se usa la más alta disponible que no la supere.</small></div>
      <Choice label="Calidad de video" variant="chips" bind:value={cfg.videoHeight} options={heights} />
    </div>
    <div class="row">
      <div class="what"><span>Contenedor de video</span></div>
      <Choice label="Contenedor de video" bind:value={cfg.videoFormat} options={[{ value: 'mp4', label: 'MP4' }, { value: 'webm', label: 'WebM' }]} />
    </div>
    <div class="row">
      <div class="what"><span>Formato de audio</span></div>
      <Choice label="Formato de audio" bind:value={cfg.audioFormat} options={[{ value: 'mp3', label: 'MP3' }, { value: 'm4a', label: 'M4A' }, { value: 'opus', label: 'Opus' }]} />
    </div>
    <div class="row">
      <div class="what"><span>Bitrate de audio</span></div>
      <Choice label="Bitrate de audio" variant="chips" bind:value={cfg.audioBitrate} options={[96, 128, 192, 256, 320].map((b) => ({ value: b, label: `${b} kbps` }))} />
    </div>
  </section>
{/if}

<section class="panel">
  <h2>Reproductor</h2>
  <div class="row">
    <div class="what"><span>Calidad de video</span><small class="muted">VP9 por software es pesado: 720p es lo más fluido en equipos modestos.</small></div>
    <Choice label="Calidad del reproductor" variant="chips" bind:value={player.height} options={[360, 480, 720, 1080].map((h) => ({ value: h, label: `${h}p` }))} />
  </div>
  <div class="row">
    <div class="what"><span>Reproducir la siguiente automáticamente</span><small class="muted">Al terminar, pasa a la siguiente de la cola.</small></div>
    <Choice label="Reproducción automática" bind:value={auto} options={[{ value: 'on', label: 'Sí' }, { value: 'off', label: 'No' }]} />
  </div>
</section>

<section class="panel">
  <h2>Motor</h2>
  {#if st}
    <div class="row">
      <div class="what"><span>yt-dlp</span><small class="muted path-text">{st.installed ? st.path : 'No instalado'}</small></div>
      <span class="status" class:ok={st.installed} class:bad={!st.installed}><i></i>{st.installed ? st.version : 'Falta'}</span>
    </div>
    <div class="row">
      <div class="what"><span>JavaScript</span><small class="muted path-text">{st.jsRuntime ? `${st.jsRuntime} · ${st.jsPath}` : 'yt-dlp lo necesita para extraer de YouTube; sin él recibe menos formatos.'}</small></div>
      <span class="status" class:ok={!!st.jsRuntime} class:bad={!st.jsRuntime}><i></i>{st.jsRuntime ? st.jsRuntime : 'Falta'}</span>
    </div>
    <div class="row">
      <div class="what"><span>ffmpeg</span><small class="muted path-text">{st.ffmpeg || 'No se encontró en el sistema ni junto a la app.'}</small></div>
      <span class="status" class:ok={!!st.ffmpeg} class:bad={!st.ffmpeg}><i></i>{st.ffmpeg ? 'Listo' : 'Falta'}</span>
    </div>
    <div class="actions">
      {#if !st.installed}
        <button class="primary" disabled={busy} onclick={() => run(InstallYtdlp)}>{busy ? 'Instalando…' : 'Instalar yt-dlp'}</button>
      {:else}
        <button disabled={busy} onclick={() => run(UpdateYtdlp)}>{busy ? 'Actualizando…' : 'Actualizar yt-dlp'}</button>
      {/if}
    </div>
    {#if msg}<p class="msg muted" role="status">{msg}</p>{/if}
  {/if}
</section>

<style>
  h1 {
    font-size: 26px;
    margin-bottom: 22px;
  }
  .panel {
    max-width: 720px;
    margin-bottom: 20px;
    padding: 6px 22px 18px;
    background: var(--surface-low);
    border: 1px solid var(--border);
    border-radius: var(--radius-panel);
  }
  h2 {
    padding: 16px 0 6px;
    font-size: 13px;
    font-weight: 500;
    letter-spacing: 0.04em;
    color: var(--text-muted);
  }
  .row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 24px;
    padding: 13px 0;
    border-top: 1px solid var(--border);
  }
  h2 + .row {
    border-top: 0;
  }
  .what {
    display: flex;
    flex-direction: column;
    gap: 2px;
    min-width: 0;
  }
  .what small {
    font-size: 12px;
  }
  .path-text {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    max-width: 380px;
    -webkit-user-select: text;
    user-select: text;
  }
  .path {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    max-width: 320px;
    padding: 7px 14px 7px 10px;
    background: var(--surface);
  }
  .path span {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .stepper {
    display: flex;
    align-items: center;
    gap: 4px;
    padding: 3px;
    background: var(--background);
    border: 1px solid var(--border);
    border-radius: 999px;
  }
  .stepper .num {
    min-width: 28px;
    text-align: center;
    font-weight: 500;
  }
  .stepper button {
    font-size: 18px;
    line-height: 1;
  }
  .status {
    display: inline-flex;
    align-items: center;
    gap: 7px;
    padding: 4px 11px 4px 9px;
    border-radius: 999px;
    background: var(--surface);
    font-size: 12px;
    font-weight: 500;
    white-space: nowrap;
  }
  .status i {
    width: 7px;
    height: 7px;
    border-radius: 50%;
    background: #7a7a7a;
  }
  .status.ok i {
    background: var(--ok);
    box-shadow: 0 0 8px rgba(43, 166, 64, 0.7);
  }
  .status.bad i {
    background: var(--primary);
    box-shadow: 0 0 8px var(--primary-glow);
  }
  .actions {
    padding-top: 14px;
    border-top: 1px solid var(--border);
  }
  .msg {
    margin-top: 12px;
    white-space: pre-wrap;
    -webkit-user-select: text;
    user-select: text;
  }
</style>
