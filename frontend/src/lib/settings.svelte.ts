import { GetSettings } from '../../wailsjs/go/main/App'
import type { Settings } from './types'

// Loaded once at startup; Settings edits this object and saves it, Download reads its defaults from it.
export const settings = $state<{ cfg: Settings | null }>({ cfg: null })

export async function loadSettings() {
  settings.cfg = (await GetSettings()) as Settings
}
