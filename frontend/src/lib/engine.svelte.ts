import { YtdlpStatus } from '../../wailsjs/go/main/App'
import type { YtdlpStatus as Status } from './types'

// yt-dlp / ffmpeg availability, shared by the rail, the download view and settings.
export const engine = $state<{ status: Status | null }>({ status: null })

export async function refreshEngine() {
  engine.status = (await YtdlpStatus()) as Status
}
