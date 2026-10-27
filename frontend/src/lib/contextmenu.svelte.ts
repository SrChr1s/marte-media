import { ClipboardGetText, ClipboardSetText } from '../../wailsjs/runtime/runtime'

export interface MenuItem {
  label: string
  hint?: string // shortcut shown on the right
  action?: () => void | Promise<void>
  disabled?: boolean
  danger?: boolean
  separator?: boolean
}

export const menu = $state<{ open: boolean; x: number; y: number; items: MenuItem[] }>({ open: false, x: 0, y: 0, items: [] })

// Opens the app's own menu at the pointer. Stops the event so the global handler doesn't also act.
export function openMenu(e: MouseEvent, items: MenuItem[]) {
  e.preventDefault()
  e.stopPropagation()
  menu.items = items
  menu.x = e.clientX
  menu.y = e.clientY
  menu.open = true
}

export function closeMenu() {
  menu.open = false
}

type TextField = HTMLInputElement | HTMLTextAreaElement
const TEXT_INPUTS = new Set(['text', 'search', 'url', 'email', 'tel', 'password'])

export function asTextField(el: EventTarget | null): TextField | null {
  if (el instanceof HTMLTextAreaElement) return el
  if (el instanceof HTMLInputElement && TEXT_INPUTS.has(el.type)) return el
  return null
}

// Cut / copy / paste / select-all for an input or textarea. The selection is read now, because
// by the time an item is clicked the field may have lost it.
export function textFieldItems(el: TextField): MenuItem[] {
  const editable = !el.readOnly && !el.disabled
  const start = el.selectionStart ?? 0
  const end = el.selectionEnd ?? 0
  const selected = el.value.slice(start, end)
  const singleLine = el instanceof HTMLInputElement

  const replace = (text: string) => {
    el.focus()
    el.setRangeText(text, start, end, 'end')
    el.dispatchEvent(new Event('input', { bubbles: true })) // keeps bind:value in sync
  }

  return [
    { label: 'Cortar', hint: 'Ctrl+X', disabled: !editable || !selected, action: async () => { await ClipboardSetText(selected); replace('') } },
    { label: 'Copiar', hint: 'Ctrl+C', disabled: !selected, action: () => void ClipboardSetText(selected) },
    {
      label: 'Pegar',
      hint: 'Ctrl+V',
      disabled: !editable,
      action: async () => {
        const text = await ClipboardGetText()
        if (text) replace(singleLine ? text.replace(/[\r\n]+/g, ' ') : text)
      },
    },
    { label: '', separator: true },
    { label: 'Seleccionar todo', hint: 'Ctrl+A', action: () => { el.focus(); el.select() } },
  ]
}

export function selectionItems(text: string): MenuItem[] {
  return [{ label: 'Copiar', hint: 'Ctrl+C', action: () => void ClipboardSetText(text) }]
}

export const copyLink = (url: string): MenuItem => ({ label: 'Copiar enlace', action: () => void ClipboardSetText(url) })
