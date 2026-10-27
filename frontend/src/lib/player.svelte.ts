import { Resolve } from '../../wailsjs/go/main/App'
import { correction } from './avsync'
import { indexAfterRemove, pickNext, pickPrev, type Repeat } from './queue'
import { toast } from './toasts.svelte'
import type { Entry, StreamInfo, Track } from './types'

export type Mode = 'video' | 'audio'
type Status = 'idle' | 'loading' | 'playing' | 'paused' | 'error'

// ---- remembered between runs (per-viewer convenience, so every access is guarded) ----
interface Saved {
  volume: number
  muted: boolean
  repeat: Repeat
  shuffle: boolean
  height: number
  autoplay: boolean
}
const KEY = 'marte-desc.player'
const defaults: Saved = { volume: 0.8, muted: false, repeat: 'off', shuffle: false, height: 720, autoplay: true }

function loadSaved(): Saved {
  try {
    return { ...defaults, ...JSON.parse(localStorage.getItem(KEY) ?? '{}') }
  } catch {
    return { ...defaults }
  }
}
const saved = loadSaved()

export const persist = () => save()

function save() {
  try {
    const { volume, muted, repeat, shuffle, height, autoplay } = player
    localStorage.setItem(KEY, JSON.stringify({ volume, muted, repeat, shuffle, height, autoplay }))
  } catch {
    /* private mode or blocked storage: the player just forgets */
  }
}

export const player = $state({
  queue: [] as Track[],
  index: -1,
  mode: 'video' as Mode,
  status: 'idle' as Status,
  error: '',
  time: 0,
  duration: 0,
  buffered: 0,
  volume: saved.volume,
  muted: saved.muted,
  repeat: saved.repeat,
  shuffle: saved.shuffle,
  autoplay: saved.autoplay,
  rate: 1,
  height: saved.height,
  heights: [] as number[],
  stream: null as StreamInfo | null,
  hasVideo: false, // the current stream has a picture to show
  videoShown: false, // a stage is on screen and the video is playing into it
  stageOpen: false, // the Reproductor view shows its "watching" layout
})

// Diagnostics (see ensure()): a short trail of what the player did and why, kept only in debug mode.
const trail: string[] = []
let debugOn = false
const note = (m: string) => {
  if (debugOn) trail.push(`${(performance.now() / 1000).toFixed(1)}s ${m}`)
}

export const current = (): Track | undefined => player.queue[player.index]
export const isCurrent = (id: string) => current()?.id === id

export function trackFromEntry(e: Entry): Track {
  return {
    id: e.id,
    url: e.url,
    title: e.title,
    channel: e.channel,
    duration: e.duration,
    thumbnail: e.thumbnail || `https://i.ytimg.com/vi/${e.id}/hqdefault.jpg`,
    viewCount: e.viewCount,
  }
}

// ---- what this webview can decode (the Go side only offers formats from these lists) ----
export function capabilities() {
  const v = document.createElement('video')
  const ok = (t: string) => v.canPlayType(t) !== ''
  const video: string[] = []
  if (ok('video/webm; codecs="vp9"')) video.push('vp9')
  // WebKitGTK answers "probably" for H.264 but decoding depends on GStreamer plugins and often fails (0x0 video)
  if (!/linux/i.test(navigator.platform) && ok('video/mp4; codecs="avc1.640028"')) video.push('avc1')
  if (ok('video/mp4; codecs="av01.0.05M.08"')) video.push('av01')
  const audio: string[] = []
  if (ok('audio/webm; codecs="opus"')) audio.push('opus')
  if (ok('audio/mp4; codecs="mp4a.40.2"')) audio.push('aac')
  return { video, audio }
}

function prefs() {
  const c = capabilities()
  return { height: player.height, videoCodecs: c.video, audioCodecs: c.audio }
}

// ---- the media elements live here, not in a view, so music keeps playing while you browse ----
let audio!: HTMLAudioElement
let video!: HTMLVideoElement
let master!: HTMLMediaElement // the element that keeps time: the audio, or the video when the file is muxed
let slave: HTMLVideoElement | null = null // the muted picture that follows the master
let wantPlay = false // what the user asked for, as opposed to what the elements happen to be doing
let pausedForVideo = false // the audio waits while the picture buffers
let pausedSince = 0
let slaveAdvancedAt = 0 // when the picture last moved: WebKit leaves readyState at 2 even while it plays
let token = 0 // identifies the load in progress; an older answer is ignored
let history: number[] = []
let played: number[] = []
let syncTimer: ReturnType<typeof setInterval> | undefined

function ensure() {
  if (audio) return
  audio = new Audio()
  audio.preload = 'auto'
  video = document.createElement('video')
  video.muted = true
  video.playsInline = true
  video.preload = 'auto'
  video.disablePictureInPicture = true
  video.style.cssText = 'width:100%;height:100%;object-fit:contain;background:#000;display:block'
  master = audio
  for (const el of [audio, video]) wire(el)
  video.addEventListener('timeupdate', () => (slaveAdvancedAt = performance.now()))
  applyVolume()
  setupMediaSession()
  // Diagnostics: with localStorage "marte-desc.debug" = "1" the elements and controls are reachable from the
  // console (used to measure audio/video drift). Off by default.
  try {
    debugOn = localStorage.getItem('marte-desc.debug') === '1'
    if (debugOn) Object.assign(window, { __marte: {
          trail,
          audio, video, player, seek, toggle, next, setQuality, setMode,
          flags: () => ({ wantPlay, pausedForVideo, slave: !!slave, master: master === audio ? 'audio' : 'video', aPaused: audio.paused, aReady: audio.readyState, aNet: audio.networkState, aErr: audio.error?.code ?? 0, vPaused: video.paused, vReady: video.readyState, vSeeking: video.seeking, vNet: video.networkState, status: player.status }),
        } })
  } catch {
    /* ignore */
  }
}

function wire(el: HTMLMediaElement) {
  const mine = () => el === master
  el.addEventListener('timeupdate', () => {
    if (!mine()) return
    player.time = el.currentTime
    // WebKit does not always send "playing" again after "waiting": if time is moving, we are playing
    if (player.status === 'loading' && wantPlay && !el.paused && !pausedForVideo) player.status = 'playing'
  })
  el.addEventListener('durationchange', () => {
    if (mine() && isFinite(el.duration) && el.duration > 0) player.duration = el.duration
  })
  el.addEventListener('progress', () => {
    if (!mine()) return
    const b = el.buffered
    for (let i = 0; i < b.length; i++) if (b.start(i) <= el.currentTime + 0.5 && b.end(i) >= el.currentTime) player.buffered = b.end(i)
  })
  el.addEventListener('playing', () => {
    if (mine() && wantPlay) player.status = 'playing'
    if (el === slave) maybeResume()
  })
  el.addEventListener('waiting', () => {
    if (mine() && wantPlay) player.status = 'loading'
    // Keep the sound from running ahead of a picture that is really out of data. Decided from the element's
    // state, not from this event alone: "waiting" can arrive after the picture is already playing again.
    if (el === slave && wantPlay && !master.paused && !slaveReady()) {
      pausedForVideo = true
      pausedSince = Date.now()
      master.pause()
    }
  })
  for (const ev of ['canplay', 'seeked']) el.addEventListener(ev, () => el === slave && maybeResume())
  el.addEventListener('ended', () => {
    if (mine()) onEnded()
  })
  el.addEventListener('error', () => {
    note(`${el === audio ? 'audio' : 'video'} element error code=${el.error?.code} ${el.error?.message ?? ''} src=${(el.currentSrc || '').slice(0, 60)}`)
    if (el === slave) {
      // lose the picture, keep the sound
      toast('error', 'El video no se pudo reproducir; sigue el audio.')
      dropSlave()
    } else if (mine() && player.status !== 'idle') {
      fail('No se pudo reproducir este contenido.')
    }
  })
}

function applyVolume() {
  const v = player.muted ? 0 : player.volume
  audio.volume = v
  video.volume = v // only audible when the video is the master (muxed file)
}

// The picture can keep up: no seek in progress, and it either reports enough data or is visibly moving.
// (readyState alone is not enough: this WebKit keeps it at HAVE_CURRENT_DATA while the video plays.)
function slaveReady() {
  if (!slave || slave.seeking) return !slave
  return slave.readyState >= 3 || (!slave.paused && performance.now() - slaveAdvancedAt < 500)
}

// Resumes the sound once the picture is ready, or after a few seconds regardless so playback can never freeze.
function maybeResume() {
  if (!pausedForVideo) return
  if (slaveReady() || Date.now() - pausedSince > 6000) {
    pausedForVideo = false
    if (wantPlay) master.play().catch(() => {})
  }
}

function startSync() {
  stopSync()
  syncTimer = setInterval(() => {
    maybeResume() // runs while the sound is paused too: this is what un-sticks a missed event
    if (!slave || !wantPlay || master.paused || !player.videoShown) return
    const c = correction(master.currentTime, slave.currentTime, player.rate)
    if (c.seekTo !== undefined) slave.currentTime = c.seekTo
    if (slave.playbackRate !== c.rate) slave.playbackRate = c.rate
  }, 300)
}

function stopSync() {
  if (syncTimer) clearInterval(syncTimer)
  syncTimer = undefined
}

function dropSlave() {
  if (!slave) return
  slave.pause()
  slave.removeAttribute('src')
  slave.load()
  slave = null
  player.hasVideo = false
  if (pausedForVideo) {
    pausedForVideo = false
    if (wantPlay) master.play().catch(() => {})
  }
}

function playSlave() {
  if (!slave || !player.videoShown || !wantPlay) return
  // only seek when it is really out of place: a needless seek makes the picture stall and the sound wait
  if (Math.abs(slave.currentTime - master.currentTime) > 0.15) slave.currentTime = master.currentTime
  slave.playbackRate = player.rate
  slave.play().catch(() => {})
}

// ---- loading a track ----
async function load(i: number, startAt = 0) {
  ensure()
  const t = player.queue[i]
  if (!t) return
  const my = ++token
  player.index = i
  player.status = 'loading'
  player.error = ''
  player.time = startAt
  player.buffered = 0
  player.duration = t.duration || 0
  stopSync()
  audio.pause()
  video.pause()
  pausedForVideo = false

  let s: StreamInfo
  const t0 = performance.now()
  note(`load #${i} ${t.id} "${t.title}"`)
  try {
    s = (await Resolve(t.id, prefs())) as StreamInfo
  } catch (e) {
    note(`resolve FAILED ${t.id} after ${Math.round(performance.now() - t0)}ms: ${e}`)
    if (my === token) fail(String(e))
    return
  }
  note(`resolved ${t.id} in ${Math.round(performance.now() - t0)}ms${my !== token ? ' (stale, ignored)' : ''}`)
  if (my !== token) return

  player.stream = s
  player.heights = s.heights ?? []
  // Music search only returns titles; the rest of the metadata arrives with the stream.
  player.queue[i] = {
    ...t,
    title: s.track || s.title || t.title,
    channel: s.artist || s.channel || t.channel,
    duration: s.duration || t.duration,
    thumbnail: s.thumbnail || t.thumbnail,
  }
  attach(s, startAt)
  updateSession()
  if (wantPlay) await play()
}

function attach(s: StreamInfo, startAt: number) {
  const wantVideo = player.mode === 'video'
  master = audio
  slave = null
  if (s.audio) audio.src = s.audio.url
  else audio.removeAttribute('src')

  if (wantVideo && s.video && s.audio) {
    video.src = s.video.url
    video.muted = true
    slave = video
  } else if (s.muxed && (wantVideo || !s.audio)) {
    // a single file with both tracks: it keeps its own time
    audio.removeAttribute('src')
    video.src = s.muxed.url
    video.muted = false
    master = video
  } else {
    video.removeAttribute('src')
  }
  player.hasVideo = slave !== null || master === video && wantVideo
  applyVolume()
  master.playbackRate = player.rate
  if (startAt > 0) {
    master.currentTime = startAt
    if (slave) slave.currentTime = startAt
  }
  startSync()
}

async function play() {
  ensure()
  wantPlay = true
  try {
    await master.play()
    playSlave()
  } catch (e) {
    note(`play() rejected: ${(e as Error).name} ${(e as Error).message}`)
    if ((e as DOMException).name !== 'AbortError') fail('No se pudo iniciar la reproducción.')
  }
}

function pause() {
  wantPlay = false
  pausedForVideo = false
  master.pause()
  slave?.pause()
  if (player.status === 'playing' || player.status === 'loading') player.status = 'paused'
}

let retriedKey = '' // the track (index + id) already given a second chance

function fail(msg: string) {
  note(`FAIL: ${msg}`)
  // a transient failure (expired link, hiccup) deserves one more try before giving up on the track
  const cur = current()
  const key = cur ? `${player.index}:${cur.id}` : ''
  if (cur && key !== retriedKey && player.status !== 'error') {
    retriedKey = key
    note(`retrying ${cur.id}`)
    load(player.index, player.time) // plays by itself when the user wanted playback
    return
  }
  player.status = 'error'
  player.error = msg
  wantPlay = false
  const t = current()
  toast('error', t ? `No se pudo reproducir «${t.title}»: ${msg}` : msg)
  // keep a playlist moving past a track that will not play
  const my = token
  if (player.autoplay && player.index + 1 < player.queue.length) setTimeout(() => my === token && next(), 1200)
}

function onEnded() {
  if (player.repeat === 'one') {
    seek(0)
    play()
    return
  }
  if (!player.autoplay) {
    wantPlay = false
    player.status = 'paused'
    return
  }
  next(true)
}

// ---- queue control ----
export function playTracks(tracks: Track[], start = 0, mode: Mode = 'video') {
  if (tracks.length === 0) return
  ensure()
  player.queue = tracks.map((t) => ({ ...t }))
  player.mode = mode
  player.stageOpen = mode === 'video'
  history = []
  played = []
  wantPlay = true
  load(Math.max(0, Math.min(start, tracks.length - 1)))
}

export function playTrack(t: Track, mode: Mode = 'video') {
  playTracks([t], 0, mode)
}

export function jump(i: number) {
  if (i === player.index) return toggle()
  history.push(player.index)
  played = []
  wantPlay = true
  load(i)
}

export function enqueue(t: Track) {
  if (player.queue.length === 0) return playTrack(t, player.mode)
  if (player.queue.some((q) => q.id === t.id)) return toast('info', 'Ya está en la cola.')
  player.queue.push({ ...t })
  toast('ok', `Agregada a la cola: ${t.title}`)
}

export function playNext(t: Track) {
  if (player.queue.length === 0) return playTrack(t, player.mode)
  player.queue.splice(player.index + 1, 0, { ...t })
  toast('ok', `Sonará a continuación: ${t.title}`)
}

export function removeAt(i: number) {
  if (i < 0 || i >= player.queue.length) return
  const wasCurrent = i === player.index
  player.queue.splice(i, 1)
  history = history.filter((h) => h !== i).map((h) => (h > i ? h - 1 : h))
  played = []
  if (!wasCurrent) {
    player.index = indexAfterRemove(player.index, i)
    return
  }
  if (player.queue.length === 0) return clear()
  load(Math.min(i, player.queue.length - 1))
}

export function clear() {
  token++
  stopSync()
  if (audio) {
    audio.pause()
    audio.removeAttribute('src')
    video.pause()
    video.removeAttribute('src')
  }
  slave = null
  wantPlay = false
  player.queue = []
  player.index = -1
  player.status = 'idle'
  player.stream = null
  player.hasVideo = false
  player.stageOpen = false
  player.time = player.duration = player.buffered = 0
  history = []
  played = []
  updateSession()
}

export function next(auto = false) {
  const s = pickNext(player.queue.length, player.index, auto ? player.repeat : player.repeat === 'one' ? 'off' : player.repeat, player.shuffle, played)
  played = s.played
  if (s.index === null) {
    wantPlay = false
    player.status = 'paused'
    return
  }
  history.push(player.index)
  wantPlay = true
  load(s.index)
}

export function prev() {
  if (player.time > 3) return seek(0) // like every player: first go back to the start of this track
  const p = pickPrev(player.queue.length, player.index, player.repeat, history)
  history = p.history
  if (p.index === null) return seek(0)
  wantPlay = true
  load(p.index)
}

// ---- transport ----
export function toggle() {
  if (!audio || player.index < 0) return
  if (player.status === 'idle' || player.status === 'error') {
    wantPlay = true
    return void load(player.index)
  }
  if (wantPlay) pause()
  else play()
}

export function seek(t: number) {
  if (!audio) return
  const to = Math.max(0, Math.min(t, player.duration || t))
  master.currentTime = to
  if (slave) slave.currentTime = to
  player.time = to
}

export function setVolume(v: number) {
  player.volume = Math.max(0, Math.min(1, v))
  if (player.volume > 0) player.muted = false
  if (audio) applyVolume()
  save()
}

export function toggleMute() {
  player.muted = !player.muted
  if (audio) applyVolume()
  save()
}

export function setRepeat(r: Repeat) {
  player.repeat = r
  save()
}

export const cycleRepeat = () => setRepeat(player.repeat === 'off' ? 'all' : player.repeat === 'all' ? 'one' : 'off')

export function toggleShuffle() {
  player.shuffle = !player.shuffle
  played = []
  save()
}

export function setAutoplay(v: boolean) {
  player.autoplay = v
  save()
}

export function setRate(r: number) {
  player.rate = r
  if (!audio) return
  master.playbackRate = r
  if (slave) slave.playbackRate = r
}

// Quality only swaps the picture: the sound is not interrupted.
export async function setQuality(h: number) {
  player.height = h
  save()
  const t = current()
  if (!t || !slave || !player.stream) return
  const my = token
  try {
    const s = (await Resolve(t.id, prefs())) as StreamInfo
    if (my !== token || !s.video || !slave) return
    const at = master.currentTime
    slave.src = s.video.url
    slave.currentTime = at
    player.stream = { ...player.stream, video: s.video }
    if (wantPlay) slave.play().catch(() => {})
  } catch (e) {
    toast('error', String(e))
  }
}

// Video or sound only for the current track (reloads it at the same position).
export function setMode(m: Mode) {
  if (player.mode === m) return
  player.mode = m
  if (current()) {
    const at = player.time
    wantPlay = wantPlay || player.status === 'playing'
    load(player.index, at)
  }
}

// ---- the picture: a view hands the video element a place to live ----
export function stage(node: HTMLElement) {
  ensure()
  node.appendChild(video)
  player.videoShown = true
  playSlave()
  return {
    destroy() {
      player.videoShown = false
      if (video.parentNode === node) node.removeChild(video)
      slave?.pause() // nobody is looking: stop decoding, the sound goes on
    },
  }
}

// ---- keyboard media keys and the system's now-playing widget ----
function setupMediaSession() {
  try {
    const ms = navigator.mediaSession
    if (!ms) return
    ms.setActionHandler('play', () => play())
    ms.setActionHandler('pause', () => pause())
    ms.setActionHandler('previoustrack', () => prev())
    ms.setActionHandler('nexttrack', () => next())
    ms.setActionHandler('seekto', (d) => d.seekTime !== undefined && seek(d.seekTime))
  } catch {
    /* not supported by this webview */
  }
}

function updateSession() {
  try {
    const t = current()
    navigator.mediaSession.metadata = t ? new MediaMetadata({ title: t.title, artist: t.channel, artwork: t.thumbnail ? [{ src: t.thumbnail }] : [] }) : null
  } catch {
    /* ignore */
  }
}
