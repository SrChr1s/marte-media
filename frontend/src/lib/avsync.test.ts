import { describe, expect, it } from 'vitest'
import { correction } from './avsync'

describe('correction', () => {
  it('leaves a video that is in sync alone', () => {
    expect(correction(10, 10.02)).toEqual({ rate: 1 })
  })
  it('speeds up a video that is a little late and slows one that is early', () => {
    expect(correction(10, 9.9).rate).toBeCloseTo(1.05)
    expect(correction(10, 10.12).rate).toBeCloseTo(0.95)
  })
  it('jumps when the gap is large', () => {
    expect(correction(10, 9.5)).toEqual({ seekTo: 10, rate: 1 })
    expect(correction(10, 11)).toEqual({ seekTo: 10, rate: 1 })
  })
  it('keeps the user playback speed as the base', () => {
    expect(correction(10, 10, 1.5)).toEqual({ rate: 1.5 })
    expect(correction(10, 9.9, 1.5).rate).toBeCloseTo(1.575)
  })
})
