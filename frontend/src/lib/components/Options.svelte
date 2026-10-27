<script lang="ts">
  import Choice from './Choice.svelte'
  import Icon from './Icon.svelte'

  let {
    mode = $bindable(),
    height = $bindable(),
    videoFormat = $bindable(),
    audioFormat = $bindable(),
    audioBitrate = $bindable(),
    heights,
    dir,
    count,
    compact = false,
    onchoose,
    ondownload,
  }: {
    mode: string
    height: number
    videoFormat: string
    audioFormat: string
    audioBitrate: number
    heights: number[]
    dir: string
    count: number
    compact?: boolean
    onchoose: () => void
    ondownload: () => void
  } = $props()

  const heightOptions = $derived(heights.map((h) => ({ value: h, label: h >= 2160 ? '4K' : `${h}p` })))
  const folder = $derived(dir.split('/').filter(Boolean).pop() || dir)
</script>

<div class="options" class:compact>
  <div class="field">
    <span class="label">Tipo</span>
    <Choice label="Tipo de descarga" bind:value={mode} options={[{ value: 'video', label: 'Video' }, { value: 'audio', label: 'Solo audio' }]} />
  </div>

  {#if mode === 'video'}
    <div class="field grow">
      <span class="label">Calidad</span>
      <Choice label="Calidad de video" variant="chips" bind:value={height} options={heightOptions} />
    </div>
    <div class="field">
      <span class="label">Contenedor</span>
      <Choice label="Contenedor de video" bind:value={videoFormat} options={[{ value: 'mp4', label: 'MP4' }, { value: 'webm', label: 'WebM' }]} />
    </div>
  {:else}
    <div class="field">
      <span class="label">Formato</span>
      <Choice
        label="Formato de audio"
        bind:value={audioFormat}
        options={[{ value: 'mp3', label: 'MP3' }, { value: 'm4a', label: 'M4A' }, { value: 'opus', label: 'Opus' }]}
      />
    </div>
    <div class="field grow">
      <span class="label">Bitrate</span>
      <Choice
        label="Bitrate de audio"
        variant="chips"
        bind:value={audioBitrate}
        options={[96, 128, 192, 256, 320].map((b) => ({ value: b, label: `${b} kbps` }))}
      />
    </div>
  {/if}

  <div class="field dest">
    <span class="label">Guardar en</span>
    <button type="button" class="ghost folder" onclick={onchoose} title={dir}>
      <Icon name="folder" size={18} /><span>{folder}</span>
    </button>
  </div>

  <button type="button" class="primary go" onclick={ondownload} disabled={count === 0}>
    <Icon name="download" size={20} />
    Descargar{count > 1 ? ` ${count} videos` : ''}
  </button>
</div>

<style>
  .options {
    display: flex;
    flex-direction: column;
    gap: 20px;
  }
  .field {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  .label {
    font-size: 12px;
    font-weight: 500;
    letter-spacing: 0.04em;
    color: var(--text-muted);
  }
  .folder {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    width: fit-content;
    max-width: 100%;
    padding: 6px 14px 6px 10px;
    background: var(--surface-low);
    border: 1px solid var(--border);
    color: var(--text);
  }
  .folder span {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .go {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 10px;
    padding: 13px 26px;
    font-size: 15px;
  }

  /* Playlist dock: everything on one wrapping row, button pushed to the end. */
  .compact {
    flex-direction: row;
    flex-wrap: wrap;
    align-items: flex-end;
    gap: 16px 28px;
  }
  .compact .go {
    margin-left: auto;
  }
</style>
