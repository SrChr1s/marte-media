<script lang="ts">
  import { fly } from 'svelte/transition'
  import { dismiss, toasts } from '../toasts.svelte'
  import Icon from './Icon.svelte'
</script>

<div class="toasts" aria-live="polite">
  {#each toasts.list as t (t.id)}
    <div class="toast {t.kind}" in:fly={{ y: 14, duration: 220 }} out:fly={{ y: 8, duration: 160 }}>
      <span class="ico"><Icon name={t.kind === 'ok' ? 'check' : 'warning'} size={16} /></span>
      <span class="text">{t.text}</span>
      {#if t.action}<button class="ghost" onclick={() => { t.action?.run(); dismiss(t.id) }}>{t.action.label}</button>{/if}
      <button class="icon" aria-label="Cerrar aviso" onclick={() => dismiss(t.id)}><Icon name="close" size={16} /></button>
    </div>
  {/each}
</div>

<style>
  .toasts {
    position: fixed;
    right: 20px;
    bottom: 20px;
    z-index: var(--z-toast);
    display: flex;
    flex-direction: column;
    gap: 8px;
    max-width: min(420px, calc(100vw - 40px));
  }
  .toast {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 8px 8px 8px 14px;
    background: var(--surface);
    border-radius: var(--radius-card);
    box-shadow: var(--shadow-menu);
  }
  .ico {
    display: grid;
    place-items: center;
    width: 22px;
    height: 22px;
    flex: none;
    border-radius: 50%;
    background: var(--ok);
    color: #fff;
  }
  .error .ico {
    background: var(--primary);
  }
  .text {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .toast button.ghost {
    padding: 4px 10px;
    color: var(--text);
  }
  .toast button.icon {
    width: 28px;
    height: 28px;
  }
</style>
