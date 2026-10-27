import { CancelProbe, CancelSearch, Probe, Search } from '../../wailsjs/go/main/App'
import { playTrack, playTracks, trackFromEntry } from './player.svelte'
import type { Entry, MediaInfo } from './types'

// Lives outside the view so the results are still there when you come back to it.
export const search = $state({
  q: '',
  kind: 'video' as 'video' | 'music',
  results: [] as Entry[],
  loading: false,
  error: '',
  searched: '',
  limit: 30,
  hasMore: false,
})

let run = 0 // a result that arrives after cancel, or after a newer search, is dropped

const PAGE = 30
export const looksLikeLink = (s: string) => /^(https?:\/\/)?((www|m|music)\.)?(youtube\.com|youtu\.be)\//i.test(s.trim())

function message(e: unknown): string {
  const m = String(e)
  return m.includes('cancelado') ? '' : m.replace(/^Error: /, '')
}

export async function runSearch(more = false) {
  const q = search.q.trim()
  if (!q) return
  if (looksLikeLink(q)) return openLink(q)
  const my = ++run
  if (more) search.limit += PAGE
  else {
    search.limit = PAGE
    search.results = []
  }
  search.loading = true
  search.error = ''
  search.searched = q
  try {
    const r = ((await Search(q, search.kind, search.limit)) as Entry[]) ?? []
    if (my !== run) return
    search.results = r
    // YouTube Music returns what it has in one page; video search can be asked for more
    search.hasMore = search.kind === 'video' && r.length >= search.limit - 3
  } catch (e) {
    if (my === run) search.error = message(e)
  } finally {
    if (my === run) search.loading = false
  }
}

// A pasted video or playlist link plays directly (a playlist becomes the queue).
export async function openLink(url: string) {
  const my = ++run
  search.loading = true
  search.error = ''
  try {
    const info = (await Probe(url.trim())) as MediaInfo
    if (my !== run) return
    const mode = search.kind === 'music' ? 'audio' : 'video'
    if (info.kind === 'playlist') {
      const tracks = (info.entries ?? []).map(trackFromEntry)
      if (tracks.length === 0) throw new Error('La playlist está vacía.')
      playTracks(tracks, 0, mode)
    } else {
      playTrack(trackFromEntry(info), mode)
    }
    search.q = ''
  } catch (e) {
    if (my === run) search.error = message(e)
  } finally {
    if (my === run) search.loading = false
  }
}

export function cancelSearch() {
  if (!search.loading) return
  run++
  search.loading = false
  CancelSearch()
  CancelProbe()
}
