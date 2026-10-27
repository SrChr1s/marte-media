export type Repeat = 'off' | 'all' | 'one'

export interface Step {
  index: number | null // null: nothing more to play
  played: number[] // indices already played in this shuffle round
}

// Next track. Without shuffle it walks the list; with shuffle it picks randomly among those not played yet.
// When everything was played, repeat "all" starts a new round and anything else stops.
export function pickNext(length: number, index: number, repeat: Repeat, shuffle: boolean, played: number[], rnd = Math.random): Step {
  if (length === 0) return { index: null, played: [] }
  if (!shuffle) {
    if (index + 1 < length) return { index: index + 1, played }
    return repeat === 'all' ? { index: 0, played } : { index: null, played }
  }
  const done = new Set([...played, index])
  let pool = [...Array(length).keys()].filter((i) => !done.has(i))
  if (pool.length === 0) {
    if (repeat !== 'all') return { index: null, played: [...done] }
    done.clear()
    done.add(index)
    pool = [...Array(length).keys()].filter((i) => i !== index)
    if (pool.length === 0) return { index, played: [index] } // a single track repeats itself
  }
  const pick = pool[Math.floor(rnd() * pool.length)]
  return { index: pick, played: [...done] }
}

// Previous track: the one played before (history), or the previous in the list when there is none.
export function pickPrev(length: number, index: number, repeat: Repeat, history: number[]): { index: number | null; history: number[] } {
  if (history.length > 0) return { index: history[history.length - 1], history: history.slice(0, -1) }
  if (length === 0) return { index: null, history }
  if (index > 0) return { index: index - 1, history }
  return { index: repeat === 'all' ? length - 1 : 0, history }
}

// Keeps the current track playing when the list changes under it (an item before it was removed, etc).
export function indexAfterRemove(index: number, removed: number): number {
  if (removed < index) return index - 1
  return index
}
