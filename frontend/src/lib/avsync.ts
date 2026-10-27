// Audio is the master clock; the (muted) video follows it. The player calls correction() a few times a second.
export const SEEK_THRESHOLD = 0.3 // seconds apart: jump
export const RATE_THRESHOLD = 0.08 // seconds apart: nudge the speed instead of jumping
export const RATE_NUDGE = 0.05

export interface Correction {
  seekTo?: number // set video.currentTime to this
  rate: number // video.playbackRate
}

export function correction(audioTime: number, videoTime: number, baseRate = 1): Correction {
  const behind = audioTime - videoTime // > 0: the video is late
  const gap = Math.abs(behind)
  if (gap > SEEK_THRESHOLD) return { seekTo: audioTime, rate: baseRate }
  if (gap > RATE_THRESHOLD) return { rate: baseRate * (behind > 0 ? 1 + RATE_NUDGE : 1 - RATE_NUDGE) }
  return { rate: baseRate }
}
