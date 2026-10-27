<script lang="ts">
  import { cubicOut } from 'svelte/easing'
  import { fly } from 'svelte/transition'
  import { ClipboardGetText } from '../../wailsjs/runtime/runtime'
  import Choice from '../lib/components/Choice.svelte'
  import Cover from '../lib/components/Cover.svelte'
  import EmptyState from '../lib/components/EmptyState.svelte'
  import Icon from '../lib/components/Icon.svelte'
  import OrbitLoader from '../lib/components/OrbitLoader.svelte'
  import QueueList from '../lib/components/QueueList.svelte'
  import Stage from '../lib/components/Stage.svelte'
  import { copyLink, openMenu } from '../lib/contextmenu.svelte'
  import { downloadTrack } from '../lib/downloads'
  import { duration, views } from '../lib/format'
  import { current, enqueue, isCurrent, playNext, playTracks, player, trackFromEntry } from '../lib/player.svelte'
  import { cancelSearch, runSearch, search } from '../lib/search.svelte'
  import type { Entry } from '../lib/types'

  const t = $derived(current())
  const watching = $derived(player.stageOpen && !!t)
  const music = $derived(search.kind === 'music')

  const suggestions = $derived(
    music ? ['lofi para concentrarse', 'rock clásico en español', 'música instrumental', 'salsa de los 70'] : ['documentales de ciencia', 'conciertos en vivo', 'recetas fáciles', 'misiones a Marte'],
  )

  function play(i: number) {
    playTracks(search.results.map(trackFromEntry), i, music ? 'audio' : 'video')
  }

  function menu(e: MouseEvent, entry: Entry, i: number) {
    const tr = trackFromEntry(entry)
    openMenu(e, [
      { label: 'Reproducir', action: () => play(i) },
      { label: 'Reproducir a continuación', action: () => playNext(tr) },
      { label: 'Agregar a la cola', action: () => enqueue(tr) },
      { label: '', separator: true },
      { label: 'Descargar video', action: () => downloadTrack(tr, 'video') },
      { label: 'Descargar audio', action: () => downloadTrack(tr, 'audio') },
      copyLink(entry.url),
    ])
  }

  async function paste() {
    try {
      const text = (await ClipboardGetText()).trim()
      if (!text) return
      search.q = text
      runSearch()
    } catch (e) {
      search.error = 'No se pudo leer el portapapeles: ' + String(e)
    }
  }

  function suggest(s: string) {
    search.q = s
    runSearch()
  }

  // changing between Video and Música re-runs the same query for the new kind
  function kindChanged() {
    if (search.searched) {
      search.q = search.searched
      runSearch()
    }
  }
</script>

<div class="page">
  {#if watching && t}
    <div class="watch" in:fly={{ y: 8, opacity: 1, duration: 200, easing: cubicOut }}>
      <div class="main">
        <button class="ghost back" onclick={() => (player.stageOpen = false)}><Icon name="back" size={18} />Volver a los resultados</button>
        <Stage />
        <h2 class="clamp2" title={t.title}>{t.title}</h2>
        <p class="muted meta">
          {[t.channel, views(t.viewCount)].filter(Boolean).join(' · ')}
        </p>
        <div class="actions">
          <button onclick={() => downloadTrack(t, 'video')}><Icon name="download" size={18} />Descargar video</button>
          <button onclick={() => downloadTrack(t, 'audio')}><Icon name="music" size={18} />Descargar audio</button>
        </div>
      </div>
      <aside class="queue">
        <h3>A continuación <span class="muted num">{player.index + 1} de {player.queue.length}</span></h3>
        <div class="qscroll"><QueueList /></div>
      </aside>
    </div>
  {:else}
    <h1 class="slim">Reproductor</h1>

    <div class="top">
      <form class="bar" onsubmit={(e) => { e.preventDefault(); runSearch() }}>
        <span class="lead-ico"><Icon name="search" size={20} /></span>
        <input
          type="text"
          bind:value={search.q}
          placeholder={music ? 'Busca canciones, artistas o álbumes' : 'Busca videos o pega un enlace'}
          aria-label="Buscar"
          onkeydown={(e) => e.key === 'Escape' && cancelSearch()}
        />
        <button type="button" class="ghost paste" onclick={paste}><Icon name="paste" size={18} />Pegar</button>
        {#if search.loading}
          <button type="button" class="cancel" onclick={cancelSearch}>Cancelar</button>
        {:else}
          <button class="primary" type="submit" disabled={!search.q.trim()}>Buscar</button>
        {/if}
      </form>
      <Choice label="Qué buscar" bind:value={search.kind} options={[{ value: 'video', label: 'Video' }, { value: 'music', label: 'Música' }]} onchange={kindChanged} />
    </div>

    {#if search.error}
      <p class="error problem" role="alert"><Icon name="warning" size={16} />{search.error}</p>
    {/if}

    {#if search.loading && search.results.length === 0}
      <p class="working muted" role="status"><OrbitLoader size={22} />Buscando… <span>Esc para cancelar</span></p>
      <div class={music ? 'list' : 'grid'} aria-busy="true">
        {#each Array(music ? 8 : 6) as _, i (i)}
          {#if music}
            <div class="skeleton srow"></div>
          {:else}
            <div class="scard"><div class="skeleton sthumb"></div><div class="skeleton sline" style="width: 85%"></div><div class="skeleton sline" style="width: 50%"></div></div>
          {/if}
        {/each}
      </div>
    {:else if search.results.length > 0}
      {#if music}
        <ul class="list">
          {#each search.results as e, i (e.id)}
            <li in:fly={{ y: 8, opacity: 0.4, duration: i < 10 ? 200 : 0, delay: Math.min(i, 8) * 18, easing: cubicOut }} oncontextmenu={(ev) => menu(ev, e, i)}>
              <button class="song" class:now={isCurrent(e.id)} onclick={() => play(i)}>
                <span class="cover">
                  <Cover src={e.thumbnail} icon="music" size={22} />
                  <span class="hov"><Icon name="play" size={22} /></span>
                </span>
                <span class="stext">
                  <span class="title clamp2">{e.title}</span>
                  {#if e.channel}<span class="muted ch">{e.channel}</span>{/if}
                </span>
                {#if e.duration}<span class="dur num muted">{duration(e.duration)}</span>{/if}
              </button>
              <button class="icon more" title="Más opciones" aria-label="Más opciones" onclick={(ev) => menu(ev, e, i)}><Icon name="more" size={20} /></button>
            </li>
          {/each}
        </ul>
      {:else}
        <div class="grid" role="list">
          {#each search.results as e, i (e.id)}
            <div class="vcard" role="listitem" in:fly={{ y: 8, opacity: 0.4, duration: i < 12 ? 200 : 0, delay: Math.min(i, 10) * 20, easing: cubicOut }} oncontextmenu={(ev) => menu(ev, e, i)}>
              <button class="open" class:now={isCurrent(e.id)} onclick={() => play(i)} title={e.title}>
                <span class="thumb">
                  <Cover src={e.thumbnail} />
                  {#if e.duration}<span class="vdur num">{duration(e.duration)}</span>{/if}
                  <span class="hov"><Icon name="play" size={34} /></span>
                </span>
                <span class="vtitle clamp2">{e.title}</span>
                <span class="muted vmeta">{[e.channel, views(e.viewCount)].filter(Boolean).join(' · ')}</span>
              </button>
            </div>
          {/each}
        </div>
      {/if}
      {#if search.hasMore}
        <div class="more-row"><button onclick={() => runSearch(true)} disabled={search.loading}>{search.loading ? 'Buscando…' : 'Más resultados'}</button></div>
      {/if}
    {:else if search.searched && !search.loading && !search.error}
      <EmptyState title="Sin resultados" text={`No encontramos nada para «${search.searched}». Prueba con otras palabras.`} />
    {:else if !search.loading && !search.error}
      <div class="intro">
        <EmptyState title={music ? 'Música sin descargar' : 'Mira sin descargar'} text={music ? 'Busca una canción o un artista y escúchalo al instante. Puedes seguir navegando mientras suena.' : 'Busca un video o pega un enlace. Lo ves aquí y, si te gusta, lo descargas con un clic.'} />
        <div class="chips" role="group" aria-label="Sugerencias">
          {#each suggestions as s}<button class="sug" onclick={() => suggest(s)}>{s}</button>{/each}
        </div>
      </div>
    {/if}
  {/if}
</div>

<style>
  .slim {
    font-size: 22px;
    margin-bottom: 18px;
  }
  .top {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 14px;
  }
  .bar {
    flex: 1 1 420px;
    max-width: 700px;
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 7px 7px 7px 16px;
    background: var(--surface-low);
    border: 1px solid var(--border-strong);
    border-radius: 999px;
    box-shadow: var(--shadow-card);
    transition: border-color var(--dur) var(--ease);
  }
  .bar:focus-within {
    border-color: rgba(255, 255, 255, 0.55);
    box-shadow:
      0 0 0 4px rgba(255, 255, 255, 0.07),
      var(--shadow-card);
  }
  .lead-ico {
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
  }
  .working {
    display: flex;
    align-items: center;
    gap: 10px;
    margin: 18px 0 14px;
  }
  .working span {
    opacity: 0.7;
    font-size: 12px;
  }

  .intro {
    margin-top: 4px;
  }
  .chips {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    margin: 18px 0 0 8px;
  }
  .sug {
    padding: 7px 14px;
    background: var(--surface-low);
    border: 1px solid var(--border);
    color: var(--text-muted);
    font-weight: 400;
  }
  .sug:hover:not(:disabled) {
    color: var(--text);
  }

  /* video results */
  .grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(230px, 1fr));
    gap: 24px 16px;
    margin-top: 24px;
  }
  .vcard .open {
    display: flex;
    flex-direction: column;
    align-items: stretch;
    gap: 8px;
    width: 100%;
    padding: 0;
    background: transparent;
    text-align: left;
    font-weight: 400;
    border-radius: var(--radius-card);
  }
  .vcard .open:hover:not(:disabled) {
    background: transparent;
  }
  .thumb {
    position: relative;
    aspect-ratio: 16 / 9;
    background: var(--surface-low);
    border-radius: var(--radius-card);
    overflow: hidden;
    transition:
      transform var(--dur) var(--ease),
      box-shadow var(--dur) var(--ease);
  }
  .vcard .open:hover .thumb {
    transform: translateY(-2px);
    box-shadow: var(--shadow-card);
  }
  .now .thumb,
  .now .cover {
    box-shadow: 0 0 0 2px var(--primary);
  }
  .vdur {
    position: absolute;
    right: 6px;
    bottom: 6px;
    padding: 1px 6px;
    border-radius: 5px;
    background: rgba(0, 0, 0, 0.82);
    font-size: 12px;
    font-weight: 500;
  }
  .hov {
    position: absolute;
    inset: 0;
    display: grid;
    place-items: center;
    background: rgba(0, 0, 0, 0.35);
    opacity: 0;
    transition: opacity var(--dur) var(--ease);
  }
  .open:hover .hov,
  .open:focus-visible .hov,
  .song:hover .hov,
  .song:focus-visible .hov {
    opacity: 1;
  }
  .vtitle {
    font-weight: 500;
    line-height: 1.3;
  }
  .vmeta {
    font-size: 12px;
  }
  .scard {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }
  .sthumb {
    aspect-ratio: 16 / 9;
  }
  .sline {
    height: 12px;
    border-radius: 6px;
  }

  /* music results */
  .list {
    list-style: none;
    margin: 22px 0 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 2px;
    max-width: 820px;
  }
  .list li {
    display: flex;
    align-items: center;
    border-radius: var(--radius-card);
  }
  .song {
    flex: 1;
    min-width: 0;
    display: flex;
    align-items: center;
    gap: 14px;
    padding: 6px 8px;
    background: transparent;
    text-align: left;
    font-weight: 400;
    border-radius: var(--radius-card);
  }
  .song:hover:not(:disabled) {
    background: rgba(255, 255, 255, 0.07);
  }
  .cover {
    position: relative;
    flex: none;
    width: 56px;
    aspect-ratio: 1;
    border-radius: 8px;
    overflow: hidden;
    background: var(--surface-low);
  }
  .stext {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }
  .title {
    font-weight: 500;
    line-height: 1.3;
  }
  .ch,
  .dur {
    font-size: 12px;
  }
  .more {
    opacity: 0;
  }
  .list li:hover .more,
  .list li:focus-within .more {
    opacity: 1;
  }
  .srow {
    height: 68px;
    border-radius: var(--radius-card);
    max-width: 820px;
  }
  .more-row {
    margin-top: 26px;
  }

  /* watching */
  .watch {
    display: grid;
    grid-template-columns: minmax(0, 1.9fr) minmax(280px, 1fr);
    gap: 28px;
    align-items: start;
  }
  .main {
    display: flex;
    flex-direction: column;
    gap: 10px;
    min-width: 0;
  }
  .back {
    align-self: flex-start;
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 6px 12px 6px 8px;
  }
  .main h2 {
    font-size: 20px;
    line-height: 1.25;
    margin-top: 6px;
  }
  .meta {
    font-size: 13px;
  }
  .actions {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    margin-top: 6px;
  }
  .actions button {
    display: inline-flex;
    align-items: center;
    gap: 8px;
  }
  .queue {
    min-width: 0;
    padding: 14px 10px 10px;
    background: var(--surface-low);
    border: 1px solid var(--border);
    border-radius: var(--radius-panel);
    max-height: calc(100vh - 150px);
    display: flex;
    flex-direction: column;
  }
  .queue h3 {
    font-size: 15px;
    padding: 0 8px 10px;
  }
  .queue h3 span {
    font-family: var(--font);
    font-weight: 400;
    font-size: 12px;
    margin-left: 6px;
  }
  .qscroll {
    overflow-y: auto;
  }

  @media (max-width: 960px) {
    .watch {
      grid-template-columns: 1fr;
    }
    .queue {
      max-height: none;
    }
  }
</style>
