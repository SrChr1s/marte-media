import { EventsOn } from '../../wailsjs/runtime/runtime'
import { History } from '../../wailsjs/go/main/App'
import type { HistoryEntry } from './types'

// Loaded once at startup so a view can paint with its data on the first frame (no pop-in).
export const history = $state<{ items: HistoryEntry[]; loaded: boolean }>({ items: [], loaded: false })

export async function loadHistory() {
  history.items = ((await History()) as HistoryEntry[]) ?? []
  history.loaded = true
}

export function initHistory() {
  loadHistory()
  return EventsOn('history:changed', loadHistory)
}
