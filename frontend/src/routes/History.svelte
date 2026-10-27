<script lang="ts">
  import { cubicOut } from 'svelte/easing'
  import { fly } from 'svelte/transition'
  import { DeleteHistory, OpenFile, OpenFolder } from '../../wailsjs/go/main/App'
  import EmptyState from '../lib/components/EmptyState.svelte'
  import Icon from '../lib/components/Icon.svelte'
  import { copyLink, openMenu } from '../lib/contextmenu.svelte'
  import { dayLabel } from '../lib/format'
  import { history } from '../lib/history.svelte'
  import type { HistoryEntry } from '../lib/types'

  const items = $derived(history.items)
  let q = $state('')

  const shown = $derived(items.filter((e) => e.title.toLowerCase().includes(q.trim().toLowerCase())))

  // Newest first, split into "Hoy", "Ayer", … sections.
  const groups = $derived.by(() => {
    const out: { label: string; items: HistoryEntry[] }[] = []
    for (const e of shown) {
      const label = dayLabel(e.finishedAt)
      const last = out[out.length - 1]
      if (last && last.label === label) last.items.push(e)
      else out.push({ label, items: [e] })
    }
    return out
  })

  async function remove(id: string) {
    await DeleteHistory(id)
    history.items = history.items.filter((e) => e.id !== id)
  }

  const time = (iso: string) => new Date(iso).toLocaleTimeString('es', { hour: '2-digit', minute: '2-digit' })
</script>

<header class="head">
  <div>
    <h1>Historial</h1>
    {#if items.length}<p class="muted num">{items.length} {items.length === 1 ? 'descarga' : 'descargas'}</p>{/if}
  </div>
  {#if items.length}
    <label class="search">
      <Icon name="search" size={18} />
      <input type="text" bind:value={q} placeholder="Buscar por título" aria-label="Buscar en el historial" />
    </label>
  {/if}
</header>

{#if history.loaded && items.length === 0}
  <EmptyState title="Todavía no hay nada aquí" text="Las descargas completadas se guardan en este historial para que puedas abrirlas de nuevo o volver a su carpeta." />
{:else if shown.length === 0}
  <p class="muted none">Ninguna descarga coincide con «{q}».</p>
{/if}

{#each groups as g (g.label)}
  <section class="group">
    <h2>{g.label}</h2>
    <ul>
      {#each g.items as e, i (e.id)}
        <li
          in:fly={{ y: 8, opacity: 0.4, duration: i < 10 ? 200 : 0, delay: Math.min(i, 6) * 20, easing: cubicOut }}
          oncontextmenu={(ev) =>
            openMenu(ev, [
              { label: 'Abrir archivo', action: () => OpenFile(e.path) },
              { label: 'Mostrar carpeta', action: () => OpenFolder(e.path) },
              copyLink(e.url),
              { label: '', separator: true },
              { label: 'Quitar del historial', danger: true, action: () => remove(e.id) },
            ])}
        >
          <div class="thumb">
            {#if e.thumbnail}<img src={e.thumbnail} alt="" />{/if}
            <span class="fmt">{e.format.toUpperCase()}</span>
          </div>
          <div class="body">
            <div class="clamp2 title" title={e.title}>{e.title}</div>
            <div class="muted num">{e.channel} · {time(e.finishedAt)}</div>
          </div>
          <div class="actions">
            <button class="icon" title="Abrir archivo" aria-label="Abrir archivo" onclick={() => OpenFile(e.path)}><Icon name="play" size={20} /></button>
            <button class="icon" title="Mostrar carpeta" aria-label="Mostrar carpeta" onclick={() => OpenFolder(e.path)}><Icon name="folder" size={20} /></button>
            <button class="icon" title="Quitar del historial" aria-label="Quitar del historial" onclick={() => remove(e.id)}><Icon name="trash" size={20} /></button>
          </div>
        </li>
      {/each}
    </ul>
  </section>
{/each}

<style>
  .head {
    display: flex;
    justify-content: space-between;
    align-items: flex-end;
    gap: 16px;
    margin-bottom: 22px;
  }
  h1 {
    font-size: 26px;
  }
  .head p {
    margin-top: 4px;
  }
  .search {
    display: flex;
    align-items: center;
    gap: 8px;
    width: 280px;
    padding: 0 8px 0 14px;
    background: var(--surface-low);
    border: 1px solid var(--border-strong);
    border-radius: 999px;
    color: var(--text-muted);
    transition:
      border-color var(--dur) var(--ease),
      box-shadow var(--dur) var(--ease);
  }
  .search:focus-within {
    border-color: rgba(255, 255, 255, 0.55);
    box-shadow: 0 0 0 4px rgba(255, 255, 255, 0.07);
  }
  .search input {
    flex: 1;
    min-width: 0;
    border: 0;
    border-radius: 0;
    background: transparent;
    box-shadow: none;
    padding: 9px 4px;
    color: var(--text);
  }
  .search input:focus {
    border: 0;
    box-shadow: none;
  }
  .none {
    margin-top: 8px;
  }

  .group {
    margin-bottom: 26px;
  }
  .group h2 {
    margin-bottom: 10px;
    font-size: 13px;
    font-weight: 500;
    letter-spacing: 0.02em;
    color: var(--text-muted);
  }
  .group h2::first-letter {
    text-transform: uppercase;
  }
  ul {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  li {
    display: grid;
    grid-template-columns: 148px minmax(0, 1fr) auto;
    gap: 16px;
    align-items: center;
    padding: 10px 12px 10px 10px;
    background: var(--surface-low);
    border: 1px solid transparent;
    border-radius: var(--radius-panel);
    transition:
      background var(--dur) var(--ease),
      border-color var(--dur) var(--ease);
  }
  li:hover,
  li:focus-within {
    background: var(--surface);
    border-color: var(--border-strong);
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
  .fmt {
    position: absolute;
    right: 6px;
    bottom: 6px;
    padding: 1px 6px;
    border-radius: 5px;
    background: rgba(0, 0, 0, 0.8);
    font-size: 10.5px;
    font-weight: 700;
    letter-spacing: 0.06em;
  }
  .body {
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 3px;
  }
  .title {
    font-weight: 500;
    line-height: 1.3;
  }
  .body .muted {
    font-size: 12px;
  }
  /* Actions fade in on hover or keyboard focus so the list stays calm. */
  .actions {
    display: flex;
    gap: 2px;
    opacity: 0;
    transition: opacity var(--dur) var(--ease);
  }
  li:hover .actions,
  li:focus-within .actions {
    opacity: 1;
  }
</style>
