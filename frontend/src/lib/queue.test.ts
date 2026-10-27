import { describe, expect, it } from 'vitest'
import { indexAfterRemove, pickNext, pickPrev } from './queue'

describe('pickNext without shuffle', () => {
  it('walks the list and stops at the end', () => {
    expect(pickNext(3, 0, 'off', false, []).index).toBe(1)
    expect(pickNext(3, 2, 'off', false, []).index).toBeNull()
  })
  it('wraps with repeat all', () => {
    expect(pickNext(3, 2, 'all', false, []).index).toBe(0)
  })
  it('handles an empty queue', () => {
    expect(pickNext(0, -1, 'all', false, []).index).toBeNull()
  })
})

describe('pickNext with shuffle', () => {
  it('never repeats a track within a round and ends when all were played', () => {
    const seen = new Set<number>([1])
    let index: number | null = 1
    let played: number[] = []
    for (let i = 0; i < 3; i++) {
      const s = pickNext(4, index!, 'off', true, played)
      expect(s.index).not.toBeNull()
      expect(seen.has(s.index!)).toBe(false)
      seen.add(s.index!)
      index = s.index
      played = s.played
    }
    expect(pickNext(4, index!, 'off', true, played).index).toBeNull()
  })
  it('starts a new round with repeat all, never picking the current track twice in a row', () => {
    const s = pickNext(3, 2, 'all', true, [0, 1], () => 0)
    expect(s.index).not.toBe(2)
    expect(s.index).not.toBeNull()
  })
  it('a single track repeats itself with repeat all', () => {
    expect(pickNext(1, 0, 'all', true, []).index).toBe(0)
  })
})

describe('pickPrev', () => {
  it('goes back through the history first', () => {
    expect(pickPrev(5, 3, 'off', [0, 4])).toEqual({ index: 4, history: [0] })
  })
  it('falls back to the previous item, clamping or wrapping at the start', () => {
    expect(pickPrev(5, 3, 'off', []).index).toBe(2)
    expect(pickPrev(5, 0, 'off', []).index).toBe(0)
    expect(pickPrev(5, 0, 'all', []).index).toBe(4)
  })
})

describe('indexAfterRemove', () => {
  it('shifts the current index when an earlier item goes away', () => {
    expect(indexAfterRemove(3, 1)).toBe(2)
    expect(indexAfterRemove(3, 3)).toBe(3)
    expect(indexAfterRemove(3, 5)).toBe(3)
  })
})
