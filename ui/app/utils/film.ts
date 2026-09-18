// Moving about inside a film by whole frames.
//
// A film is a sequence of still pictures, and the useful unit for finding an
// exact moment in one is a frame rather than a round number of seconds — which
// is why the step is derived from the format's frame rate instead of being a
// decimal somebody once found close enough.

/**
 * How fast each format ran.
 *
 * Super 8 ran at 18 frames a second for silent film, which is why a frame is
 * 0.0556 seconds and not the 0.05 the old buttons used — close enough to look
 * right and wrong by one frame every eighteen.
 *
 * A format that is not listed has no frame stepping rather than a guessed rate:
 * stepping by the wrong frame length is worse than not offering it, because it
 * lands between frames and looks as though the film is stuck.
 */
export const frameRates: Record<string, number> = { super8: 18 }

/** How long one frame lasts, or nothing if the format's rate is unknown. */
export function frameSeconds(format: string): number | undefined {
  const fps = frameRates[format]

  return fps ? 1 / fps : undefined
}

/**
 * Where a step of `by` seconds from `current` lands.
 *
 * Clamped at both ends, because a video element treats the two badly in
 * different ways: a negative time is refused outright, and a time past the end
 * leaves the film sitting on its last frame having fired `ended`.
 *
 * A duration that is not yet known — zero, or NaN before the metadata has
 * loaded — clamps only at the start, since the end is not knowable yet.
 */
export function stepTo(current: number, by: number, duration?: number): number {
  const next = current + by
  if (next < 0) return 0
  if (duration && Number.isFinite(duration) && next > duration) return duration

  return next
}

/**
 * A position in the film, to the frame: `1:07.12` is seven seconds and twelve
 * frames past the minute.
 *
 * Whole seconds are not enough next to the frame buttons — eighteen presses
 * would leave the reading unchanged and the buttons looking broken.
 *
 * Counted in frames throughout rather than by taking the fraction of a second,
 * because a frame length does not divide a second evenly: eighteen steps of
 * 1/18 land on 0.9999999999999999, and a fraction-based reading calls that
 * frame eighteen of a second that only has eighteen.
 */
export function timecode(seconds: number, fps: number): string {
  const frames = Math.max(0, Math.round(seconds * fps))
  const whole = Math.floor(frames / fps)

  return `${Math.floor(whole / 60)}:${String(whole % 60).padStart(2, '0')}.${String(frames % fps).padStart(2, '0')}`
}
