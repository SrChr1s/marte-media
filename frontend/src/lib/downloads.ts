import { Enqueue } from '../../wailsjs/go/main/App'
import { settings } from './settings.svelte'
import { toast } from './toasts.svelte'
import type { Request, Track } from './types'

// Sends a track to the existing download queue, with the user's default formats.
export async function downloadTrack(t: Track, mode: 'video' | 'audio') {
  const c = settings.cfg
  if (!c) return
  const r: Request = {
    url: t.url,
    title: t.title,
    channel: t.channel,
    thumbnail: t.thumbnail,
    mode,
    height: c.videoHeight,
    videoFormat: c.videoFormat,
    audioFormat: c.audioFormat,
    audioBitrate: c.audioBitrate,
    dir: c.downloadDir,
  }
  await Enqueue([r])
  toast('ok', mode === 'audio' ? `Descargando el audio: ${t.title}` : `Descargando el video: ${t.title}`)
}
