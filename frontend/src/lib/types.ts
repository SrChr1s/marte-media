export interface Entry {
  id: string
  url: string
  title: string
  channel: string
  duration: number
  thumbnail: string
  viewCount?: number
}

export interface MediaInfo extends Entry {
  kind: 'video' | 'playlist'
  heights: number[] | null
  entries: Entry[] | null
  playlistCount: number
}

export interface Request {
  url: string
  title: string
  channel: string
  thumbnail: string
  mode: string
  height: number
  videoFormat: string
  audioFormat: string
  audioBitrate: number
  dir: string
}

export interface Job {
  id: string
  request: Request
  status: 'queued' | 'downloading' | 'processing' | 'done' | 'error' | 'canceled'
  percent: number
  downloaded: number
  total: number
  speed: number
  eta: number
  path: string
  error: string
}

export interface HistoryEntry {
  id: string
  url: string
  title: string
  channel: string
  thumbnail: string
  mode: string
  format: string
  path: string
  finishedAt: string
}

export interface Settings {
  downloadDir: string
  maxConcurrent: number
  mode: string
  videoHeight: number
  videoFormat: string
  audioFormat: string
  audioBitrate: number
}

export interface YtdlpStatus {
  installed: boolean
  path: string
  version: string
  ffmpeg: string
  jsRuntime: string
  jsPath: string
}

// Player
export interface Track {
  id: string
  url: string
  title: string
  channel: string
  duration: number
  thumbnail: string
  viewCount?: number
}

export interface SourceInfo {
  url: string
  mime: string
  height: number
  codec: string
}

export interface StreamInfo {
  id: string
  title: string
  channel: string
  artist: string
  track: string
  thumbnail: string
  duration: number
  heights: number[] | null
  video: SourceInfo | null
  audio: SourceInfo | null
  muxed: SourceInfo | null
}
