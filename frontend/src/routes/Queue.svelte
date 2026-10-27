<script lang="ts">
  import { flip } from 'svelte/animate'
  import { cubicOut } from 'svelte/easing'
  import { fly } from 'svelte/transition'
  import { Cancel, ClearFinished, OpenFile, OpenFolder, Retry } from '../../wailsjs/go/main/App'
  import EmptyState from '../lib/components/EmptyState.svelte'
  import Icon from '../lib/components/Icon.svelte'
  import ProgressBar from '../lib/components/ProgressBar.svelte'
  import { copyLink, openMenu, type MenuItem } from '../lib/contextmenu.svelte'
  import { bytes, eta, modeLabel, pct } from '../lib/format'
  import { dropFinished, jobs } from '../lib/jobs.svelte'
  import type { Job } from '../lib/types'

  let { ongo }: { ongo: () => void } = $props()

  const labels: Record<Job['status'], string> = {
    queued: 'En espera',
    downloading: 'Descargando',
    processing: 'Procesando',
    done: 'Completada',
    error: 'Con error',
    canceled: 'Cancelada',
  }
  const running = (j: Job) => j.status === 'queued' || j.status === 'downloading' || j.status === 'processing'
  const activeN = $derived(jobs.list.filter(running).length)
  const finishedN = $derived(jobs.list.length - activeN)
  const newestFirst = $derived([...jobs.list].reverse())

  function menuFor(j: Job): MenuItem[] {
    const items: MenuItem[] = []
    if (running(j)) items.push({ label: 'Cancelar', danger: true, action: () => Cancel(j.id) })
    else if (j.status === 'done') items.push({ label: 'Abrir archivo', action: () => OpenFile(j.path) }, { label: 'Mostrar carpeta', action: () => OpenFolder(j.path) })
    else items.push({ label: 'Reintentar', action: () => Retry(j.id) })
    return [...items, copyLink(j.request.url)]
  }

  function clear() {
    ClearFinished()
    dropFinished()
  }
</script>

<header class="head">
  <div>
    <h1>Cola</h1>
    {#if jobs.list.length}
      <p class="muted num">{activeN} {activeN === 1 ? 'activa' : 'activas'} · {finishedN} {finishedN === 1 ? 'terminada' : 'terminadas'}</p>
    {/if}
  </div>
  {#if finishedN > 0}<button class="ghost" onclick={clear}>Limpiar terminadas</button>{/if}
</header>

{#if jobs.list.length === 0}
  <EmptyState title="Sin descargas en curso" text="Cuando empieces una descarga verás aquí el progreso, la velocidad y el tiempo restante.">
    <button class="primary" onclick={ongo}>Ir a Descargar</button>
  </EmptyState>
{/if}

<ul>
  {#each newestFirst as j, i (j.id)}
    <li
      class="s-{j.status}"
      animate:flip={{ duration: 240 }}
      in:fly={{ y: 8, opacity: 0.4, duration: i < 10 ? 200 : 0, delay: Math.min(i, 6) * 20, easing: cubicOut }}
      oncontextmenu={(ev) => openMenu(ev, menuFor(j))}
    >
      <div class="thumb">
        {#if j.request.thumbnail}<img src={j.request.thumbnail} alt="" />{/if}
        <span class="kind"><Icon name={j.request.mode === 'audio' ? 'music' : 'film'} size={13} /></span>
      </div>

      <div class="body">
        <div class="title clamp2" title={j.request.title}>{j.request.title || j.request.url}</div>
        <div class="meta">
          <span class="chip"><span class="dot"></span>{labels[j.status]}</span>
          <span class="muted">{modeLabel(j.request)}</span>
        </div>

        {#if j.status === 'downloading'}
          <div class="stats num">
            <b>{pct(j.percent)}</b>
            {#if j.total}<span>{bytes(j.downloaded)} de {bytes(j.total)}</span>{/if}
            {#if j.speed}<span>{bytes(j.speed)}/s</span>{/if}
            <span>quedan {eta(j.eta)}</span>
          </div>
        {:else if j.status === 'processing'}
          <div class="stats"><span>Uniendo pistas y aplicando metadatos…</span></div>
        {:else if j.status === 'error'}
          <div class="error detail">{j.error || 'La descarga falló.'}</div>
        {/if}

        {#if j.status === 'downloading' || j.status === 'processing' || j.status === 'done'}
          <ProgressBar percent={j.percent} state={j.status === 'done' ? 'done' : 'active'} />
        {/if}
      </div>

      <div class="actions">
        {#if running(j)}
          <button class="icon" title="Cancelar" aria-label="Cancelar" onclick={() => Cancel(j.id)}><Icon name="close" size={20} /></button>
        {:else if j.status === 'done'}
          <button class="icon" title="Abrir archivo" aria-label="Abrir archivo" onclick={() => OpenFile(j.path)}><Icon name="play" size={20} /></button>
          <button class="icon" title="Mostrar carpeta" aria-label="Mostrar carpeta" onclick={() => OpenFolder(j.path)}><Icon name="folder" size={20} /></button>
        {:else}
          <button class="icon" title="Reintentar" aria-label="Reintentar" onclick={() => Retry(j.id)}><Icon name="retry" size={20} /></button>
        {/if}
      </div>
    </li>
  {/each}
</ul>

<style>
  .head {
    display: flex;
    justify-content: space-between;
    align-items: flex-end;
    margin-bottom: 20px;
  }
  h1 {
    font-size: 26px;
  }
  .head p {
    margin-top: 4px;
  }

  ul {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 12px;
  }
  li {
    display: grid;
    grid-template-columns: 168px minmax(0, 1fr) auto;
    gap: 16px;
    align-items: center;
    padding: 12px;
    background: var(--surface-low);
    border: 1px solid var(--border);
    border-radius: var(--radius-panel);
    transition:
      border-color var(--dur) var(--ease),
      box-shadow var(--dur) var(--ease);
  }
  li:hover {
    border-color: var(--border-strong);
  }
  li.s-downloading,
  li.s-processing {
    border-color: rgba(255, 0, 0, 0.28);
    box-shadow: var(--shadow-card);
  }
  li.s-canceled {
    opacity: 0.7;
  }

  .thumb {
    position: relative;
    aspect-ratio: 16 / 9;
    background: var(--background);
    border-radius: var(--radius-card);
    overflow: hidden;
  }
  .thumb img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
  }
  .kind {
    position: absolute;
    left: 6px;
    top: 6px;
    display: grid;
    place-items: center;
    width: 22px;
    height: 22px;
    border-radius: 6px;
    background: rgba(0, 0, 0, 0.72);
  }

  .body {
    display: flex;
    flex-direction: column;
    gap: 8px;
    min-width: 0;
  }
  .title {
    font-weight: 500;
    line-height: 1.3;
  }
  .meta {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 10px;
    font-size: 12px;
  }
  .chip {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 2px 9px 2px 7px;
    border-radius: 999px;
    background: var(--surface);
    font-weight: 500;
  }
  .dot {
    width: 7px;
    height: 7px;
    border-radius: 50%;
    background: #7a7a7a;
  }
  .s-downloading .dot,
  .s-processing .dot {
    background: var(--primary);
    box-shadow: 0 0 8px var(--primary-glow);
    animation: pulse 1.4s ease-in-out infinite;
  }
  .s-done .dot {
    background: var(--ok);
  }
  .s-error .dot {
    background: var(--danger);
  }
  @keyframes pulse {
    50% {
      opacity: 0.35;
    }
  }

  .stats {
    display: flex;
    flex-wrap: wrap;
    gap: 4px 14px;
    font-size: 12px;
    color: var(--text-muted);
  }
  .stats b {
    color: var(--text);
    font-weight: 500;
  }
  .detail {
    font-size: 12px;
  }

  .actions {
    display: flex;
    gap: 2px;
  }

  @media (max-width: 820px) {
    li {
      grid-template-columns: 120px minmax(0, 1fr) auto;
    }
  }
</style>
