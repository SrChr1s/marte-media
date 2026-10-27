export interface Toast {
  id: number
  kind: 'ok' | 'error' | 'info'
  text: string
  action?: { label: string; run: () => void }
}

export const toasts = $state<{ list: Toast[] }>({ list: [] })
let next = 1

export function toast(kind: Toast['kind'], text: string, action?: Toast['action'], ms = 5000) {
  const t: Toast = { id: next++, kind, text, action }
  toasts.list = [...toasts.list, t].slice(-4)
  setTimeout(() => dismiss(t.id), ms)
}

export function dismiss(id: number) {
  toasts.list = toasts.list.filter((t) => t.id !== id)
}
