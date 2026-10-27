<script lang="ts">
  import { onMount } from 'svelte'
  import { asTextField, closeMenu, menu, openMenu, selectionItems, textFieldItems, type MenuItem } from '../contextmenu.svelte'

  let el = $state<HTMLDivElement>()
  let pos = $state({ x: 0, y: 0 })

  // Right-click anywhere that no component handled itself: text fields and selected text get a menu,
  // everything else gets nothing (the native menu is hidden via --default-contextmenu in style.css).
  function onContext(e: MouseEvent) {
    const field = asTextField(e.target)
    const selection = window.getSelection()?.toString() ?? ''
    if (field) openMenu(e, textFieldItems(field))
    else if (selection) openMenu(e, selectionItems(selection))
    else {
      e.preventDefault()
      closeMenu()
    }
  }

  onMount(() => {
    const onPointerDown = (e: PointerEvent) => {
      if (menu.open && !el?.contains(e.target as Node)) closeMenu()
    }
    const onKey = (e: KeyboardEvent) => {
      if (menu.open && e.key === 'Escape') closeMenu()
    }
    window.addEventListener('contextmenu', onContext)
    window.addEventListener('pointerdown', onPointerDown, true)
    window.addEventListener('keydown', onKey, true)
    window.addEventListener('blur', closeMenu)
    window.addEventListener('resize', closeMenu)
    window.addEventListener('scroll', closeMenu, true)
    return () => {
      window.removeEventListener('contextmenu', onContext)
      window.removeEventListener('pointerdown', onPointerDown, true)
      window.removeEventListener('keydown', onKey, true)
      window.removeEventListener('blur', closeMenu)
      window.removeEventListener('resize', closeMenu)
      window.removeEventListener('scroll', closeMenu, true)
    }
  })

  // Keep the menu inside the window once its size is known.
  $effect(() => {
    if (!menu.open || !el) return
    const r = el.getBoundingClientRect()
    pos = {
      x: Math.max(8, Math.min(menu.x, window.innerWidth - r.width - 8)),
      y: Math.max(8, Math.min(menu.y, window.innerHeight - r.height - 8)),
    }
  })

  async function run(item: MenuItem) {
    closeMenu()
    await item.action?.()
  }

  function onMenuKey(e: KeyboardEvent) {
    const buttons = [...(el?.querySelectorAll<HTMLButtonElement>('button:not(:disabled)') ?? [])]
    const i = buttons.indexOf(document.activeElement as HTMLButtonElement)
    if (e.key === 'ArrowDown') buttons[(i + 1) % buttons.length]?.focus()
    else if (e.key === 'ArrowUp') buttons[(i - 1 + buttons.length) % buttons.length]?.focus()
    else return
    e.preventDefault()
  }
</script>

{#if menu.open}
  <!-- mousedown is cancelled so a text field keeps its focus and selection while a command runs -->
  <div
    class="menu"
    role="menu"
    tabindex="-1"
    bind:this={el}
    style="left: {pos.x}px; top: {pos.y}px"
    onmousedown={(e) => e.preventDefault()}
    onkeydown={onMenuKey}
  >
    {#each menu.items as item}
      {#if item.separator}
        <div class="sep" role="separator"></div>
      {:else}
        <button class="item" class:danger={item.danger} role="menuitem" disabled={item.disabled} onclick={() => run(item)}>
          <span>{item.label}</span>
          {#if item.hint}<span class="hint">{item.hint}</span>{/if}
        </button>
      {/if}
    {/each}
  </div>
{/if}

<style>
  .menu {
    position: fixed;
    z-index: var(--z-menu);
    min-width: 200px;
    padding: 6px 0;
    background: var(--surface);
    border: 1px solid #3f3f3f;
    border-radius: var(--radius-card);
    box-shadow: var(--shadow-menu);
  }
  .item {
    display: flex;
    justify-content: space-between;
    gap: 24px;
    width: 100%;
    padding: 8px 16px;
    border-radius: 0;
    background: none;
    font-weight: 400;
    text-align: left;
  }
  .item:hover:not(:disabled),
  .item:focus-visible {
    background: var(--surface-hover);
    outline: none;
  }
  .item:disabled {
    color: var(--text-muted);
    opacity: 0.5;
  }
  .item.danger {
    color: #ff6b6b;
  }
  .hint {
    color: var(--text-muted);
    font-size: 12px;
  }
  .sep {
    height: 1px;
    margin: 6px 0;
    background: #3f3f3f;
  }
</style>
