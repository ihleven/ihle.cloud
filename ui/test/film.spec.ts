import { describe, expect, it } from 'vitest'
import { frameSeconds, stepTo, timecode } from '../app/utils/film'

describe('frameSeconds', () => {
  // The old buttons stepped 0.05s, which is a frame and a bit short: over
  // eighteen presses it loses a whole frame.
  it('is a real frame of Super 8, not a round number near one', () => {
    expect(frameSeconds('super8')).toBeCloseTo(1 / 18, 10)
    expect(frameSeconds('super8')).not.toBe(0.05)
  })

  it('offers nothing for a format whose rate is unknown', () => {
    expect(frameSeconds('16mm')).toBeUndefined()
    expect(frameSeconds('')).toBeUndefined()
  })
})

describe('stepTo', () => {
  it('moves forward and back', () => {
    expect(stepTo(10, 1, 60)).toBe(11)
    expect(stepTo(10, -1, 60)).toBe(9)
  })

  it('stops at the beginning rather than going negative', () => {
    expect(stepTo(0.02, -1, 60)).toBe(0)
    expect(stepTo(0, -1 / 18, 60)).toBe(0)
  })

  it('stops at the end rather than running past it', () => {
    expect(stepTo(59.9, 1, 60)).toBe(60)
  })

  // Before the metadata loads, duration is NaN and there is no end to clamp to.
  it('clamps only the start while the duration is unknown', () => {
    expect(stepTo(10, 5, Number.NaN)).toBe(15)
    expect(stepTo(10, 5, 0)).toBe(15)
    expect(stepTo(10, 5, undefined)).toBe(15)
    expect(stepTo(0.5, -1, Number.NaN)).toBe(0)
  })

  it('steps by whole frames at eighteen a second', () => {
    const frame = frameSeconds('super8')!
    expect(stepTo(1, frame, 60)).toBeCloseTo(1 + 1 / 18, 10)
    // Eighteen steps is one second, which 0.05 would have missed by a frame.
    let t = 0
    for (let i = 0; i < 18; i++) t = stepTo(t, frame, 60)
    expect(t).toBeCloseTo(1, 10)
  })
})

describe('timecode', () => {
  it('reads minutes, seconds and frames', () => {
    expect(timecode(0, 18)).toBe('0:00.00')
    expect(timecode(67 + 12 / 18, 18)).toBe('1:07.12')
  })

  // The reason it counts in frames rather than taking the fraction of a second:
  // a frame length does not divide a second evenly, so eighteen of them land
  // beside one second rather than on it — above or below, depending on the
  // arithmetic. Either way a fraction-based reading misreports the frame.
  it('survives the drift of adding up frame lengths', () => {
    let t = 0
    for (let i = 0; i < 18; i++) t += 1 / 18

    expect(t).not.toBe(1)
    expect(timecode(t, 18)).toBe('0:01.00')
    // And from the other side of one second, the same reading.
    expect(timecode(0.99999999, 18)).toBe('0:01.00')
  })

  it('advances one frame at a time', () => {
    expect(timecode(0, 18)).toBe('0:00.00')
    expect(timecode(1 / 18, 18)).toBe('0:00.01')
    expect(timecode(2 / 18, 18)).toBe('0:00.02')
    expect(timecode(17 / 18, 18)).toBe('0:00.17')
  })

  it('never reads a negative position', () => {
    expect(timecode(-0.4, 18)).toBe('0:00.00')
  })
})
