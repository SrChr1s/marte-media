<script lang="ts">
  import { duration } from '../format'
  import { copyLink, openMenu } from '../contextmenu.svelte'
  import { downloadTrack } from '../downloads'
  import { jump, player, removeAt } from '../player.svelte'
  import Cover from './Cover.svelte'
  import Icon from './Icon.svelte'
</script>

{#if player.queue.length === 0}
  <p class="empty muted">La cola está vacía.</p>
{:else}
  <ul>
    {#each player.queue as t, i (t.id + ':' + i)}
      {@const now = i === player.index}
      <li
        class:now
        oncontextmenu={(e) =>
          openMenu(e, [
            { label: 'Reproducir ahora', action: () => jump(i) },
            { label: 'Quitar de la cola', action: () => removeAt(i) },
            { label: '', separator: true },
            { label: 'Descargar video', action: () => downloadTrack(t, 'video') },
            { label: 'Descargar audio', action: () => downloadTrack(t, 'audio') },
            copyLink(t.url),
          ])}
      >
        <button class="row" onclick={() => jump(i)} title={t.title}>
          <span class="lead">
            {#if now}
              <span class="eq" class:still={player.status !== 'playing'} aria-label="Reproduciendo"><i></i><i></i><i></i></span>
            {:else}
              <span class="num">{i + 1}</span>
            {/if}
          </span>
          <span class="thumb"><Cover src={t.thumbnail} size={18} /></span>
          <span class="text">
            <span class="title clamp2">{t.title}</span>
            {#if t.channel}<span class="muted ch">{t.channel}</span>{/if}
          </span>
          {#if t.duration}<span class="dur num muted">{duration(t.duration)}</span>{/if}
        </button>
        <button class="icon rm" title="Quitar de la cola" aria-label="Quitar de la cola" onclick={() => removeAt(i)}><Icon name="close" size={18} /></button>
      </li>
    {/each}
  </ul>
{/if}

<style>
  .empty {
    padding: 14px 4px;
  }
  ul {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 2px;
  }
  li {
    display: flex;
    align-items: center;
    border-radius: var(--radius-card);
  }
  li.now {
    background: rgba(255, 0, 0, 0.1);
  }
  .row {
    flex: 1;
    min-width: 0;
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 6px 4px 6px 8px;
    background: transparent;
    text-align: left;
    font-weight: 400;
    border-radius: var(--radius-card);
  }
  .row:hover:not(:disabled) {
    background: rgba(255, 255, 255, 0.06);
  }
  .lead {
    width: 22px;
    flex: none;
    display: grid;
    place-items: center;
  }
  .num {
    font-size: 12px;
    color: var(--text-muted);
  }
  .eq {
    display: inline-flex;
    align-items: flex-end;
    gap: 2px;
    height: 14px;
  }
  .eq i {
    width: 3px;
    height: 100%;
    background: var(--primary);
    transform-origin: bottom;
    animation: bounce 0.9s ease-in-out infinite;
  }
  .eq i:nth-child(2) {
    animation-delay: -0.3s;
  }
  .eq i:nth-child(3) {
    animation-delay: -0.6s;
  }
  .eq.still i {
    animation-play-state: paused;
    transform: scaleY(0.35);
  }
  @keyframes bounce {
    0%,
    100% {
      transform: scaleY(0.3);
    }
    50% {
      transform: scaleY(1);
    }
  }
  .thumb {
    flex: none;
    width: 72px;
    aspect-ratio: 16 / 9;
    border-radius: 6px;
    overflow: hidden;
    background: var(--surface-low);
  }
  .text {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 1px;
  }
  .title {
    font-weight: 500;
    font-size: 13px;
    line-height: 1.25;
  }
  .ch,
  .dur {
    font-size: 12px;
  }
  .rm {
    width: 30px;
    height: 30px;
    margin-right: 4px;
    opacity: 0;
  }
  li:hover .rm,
  li:focus-within .rm {
    opacity: 1;
  }
</style>
