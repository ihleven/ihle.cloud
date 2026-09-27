/**
 * Weight, given to something that has none.
 *
 * The tiles' labels are painted on the page and move with it exactly, which is
 * what makes a grid of them feel like a printed sheet. Giving them a little
 * mass makes the page feel like a thing rather than a picture of one: scroll,
 * and they swing; stop, and they settle back where they belong.
 *
 * Every label gets its own weight, and that is the difference between a grid of
 * things and one big thing. Sharing a single number moved all twelve exactly
 * together, which reads as one slab sliding about behind the tiles — the eye
 * reads *common* motion as one object. Given their own, they answer the same
 * scroll at different strengths and settle at different moments, and the grid
 * ripples.
 *
 * Not a scroll-driven animation, and it cannot be one. `view()` and `scroll()`
 * map a *position* — where a thing is in its scroller — and this is about
 * *speed*: how fast the page is moving and how hard it stopped. There is no
 * position that answers that, so it is worked out a frame at a time.
 *
 * One loop for the whole grid, and one property written per label. None of it
 * goes through Vue, which would mean a re-render sixty times a second to move
 * something by four pixels.
 */
export interface ScrollGravity {
  /**
   * How far a given speed throws them, in pixels per pixel-of-scroll-per-frame.
   * The whole feel is in this number, and it has to be small: an ordinary
   * trackpad scroll covers twenty to forty pixels a frame and a flick covers a
   * hundred and fifty, so anything near 1 pins every scroll at the limit and
   * the labels stop answering how fast the page is moving — which is the one
   * thing they are for. At 0.12 an ordinary scroll is worth three or four
   * pixels and only a flick reaches the end of the travel.
   */
  gain?: number
  /** The furthest any of them will go, however hard the page is flung. */
  max?: number
  /** How strongly they are pulled back. Higher settles faster. */
  stiffness?: number
  /**
   * How much of the swing is taken out per frame. Below about 0.5 they
   * overshoot and come back, which is the part that reads as weight; above it
   * they slide home and the whole thing looks like easing.
   */
  damping?: number
  /**
   * How much the strength differs from tile to tile, 0 to 1.
   *
   * Strength and nothing else: they are pulled home equally hard, so they
   * settle together and differ only in how far they went. At 0 they are
   * identical and move as one slab; at 1 the lightest takes about a third of
   * the throw the heaviest does.
   */
  spread?: number
  /** Which elements are given weight, within the container. */
  selector?: string
}

/** One label, its weight, and where its spring currently is. */
interface Weighted {
  el: HTMLElement
  /** Its share of the throw. Around 1, above or below by up to `spread`. */
  weight: number
  shift: number
  speed: number
}

/**
 * A number in 0..1 for index i, spread evenly and never repeating.
 *
 * The golden ratio rather than a random number, and rather than a ramp. Random
 * would give a different grid on every reload, which is a page that cannot be
 * looked at twice. A ramp — heavier the further down — turns out to read as a
 * mistake in the layout rather than as character, because the eye finds the
 * order and then expects it to mean something. This has neither problem: it is
 * the same every time and it has no pattern to find.
 */
function scatter(i: number): number {
  return (i * 0.618033988749895) % 1
}

/**
 * Writes `--gravity` onto each label: the pixels it is currently displaced by.
 *
 * Positive is downwards. Scrolling towards the top of the page throws them
 * down and scrolling away throws them up — the direction the content itself
 * appears to travel, so they read as being carried along by it rather than
 * dragging behind it. The other sign is the classic inertia feel, a mass
 * lagging its container, and it is the single minus below.
 */
export function useScrollGravity(container: Ref<HTMLElement | null>, options: ScrollGravity = {}) {
  const {
    gain = 0.12,
    max = 16,
    stiffness = 0.14,
    damping = 0.55,
    spread = 0.45,
    selector = '.tile-label',
  } = options

  let last = 0
  let stamp = 0
  let frame = 0
  let weighted: Weighted[] = []

  const clamp = (n: number) => Math.min(max, Math.max(-max, n))

  /**
   * Find the labels and hand each its weight.
   *
   * Done when a burst of scrolling begins rather than once on mount: which
   * tiles exist follows from what the account is entitled to, and that arrives
   * after this composable does. Re-reading costs one query per burst — not per
   * frame — and it is what keeps the weights attached to the right tiles when
   * the grid changes underneath.
   */
  function collect() {
    const root = container.value
    if (!root) {
      weighted = []

      return
    }

    const found = [...root.querySelectorAll<HTMLElement>(selector)]
    if (found.length === weighted.length && found.every((el, i) => weighted[i]?.el === el)) return

    weighted = found.map((el, i) => ({
      el,
      // Around 1, so `gain` still means what it says on average and the spread
      // is a spread rather than a quiet change of strength.
      weight: 1 + (scatter(i) - 0.5) * 2 * spread,
      shift: 0,
      speed: 0,
    }))
  }

  function rest() {
    frame = 0
    for (const w of weighted) {
      w.shift = 0
      w.speed = 0
      w.el.style.setProperty('--gravity', '0px')
    }
  }

  function tick(at: number) {
    if (!container.value) return rest()

    // How much of a sixtieth of a second this frame was. A phone and a laptop
    // both draw at 120Hz now, which halves the distance the page covers between
    // frames — so a spring counted in frames is twice as gentle on the good
    // screen and the whole effect is tuned for whichever machine it was tuned
    // on. Measured against time instead, it feels the same on both.
    //
    // Clamped, because a tab returning from the background reports a frame
    // several seconds long, and an integration step that size throws the spring
    // somewhere it cannot come back from.
    const step = Math.min(2, Math.max(0.2, (at - stamp) / 16.667))
    stamp = at

    // Read the speed here rather than in the listener: scroll events arrive in
    // bursts of several per frame and a browser may coalesce them differently
    // from one to the next, so counting them measures the browser. The distance
    // covered since the last frame is the thing that is actually being asked
    // about.
    const now = window.scrollY
    const travelled = now - last
    last = now

    // The page's own speed, worked out once. What differs between the labels is
    // what each of them makes of it.
    const thrown = -(travelled / step) * gain

    let moving = Math.abs(travelled) >= 0.05
    for (const w of weighted) {
      const target = clamp(thrown * w.weight)

      // A spring, integrated once per frame. The throw is where it is being
      // pulled *towards*, so when the page stops moving it falls to nothing and
      // the same spring carries the label home — one behaviour rather than a
      // movement and a separate rule about returning.
      w.speed += ((target - w.shift) * stiffness - w.speed * damping) * step
      w.shift += w.speed * step

      // The limit is on where they end up, not only on where they are pulled
      // towards. A spring loose enough to feel like weight overshoots what it
      // is aiming at, so clamping the aim alone lets the swing past it. The
      // velocity goes with it: kept, it would go on pressing against the stop
      // and the label would sit pinned there instead of falling back.
      if (Math.abs(w.shift) > max) {
        w.shift = Math.sign(w.shift) * max
        w.speed = 0
      }

      w.el.style.setProperty('--gravity', `${w.shift.toFixed(2)}px`)

      if (Math.abs(w.shift) >= 0.05 || Math.abs(w.speed) >= 0.05) moving = true
    }

    // Stopped, and all of them back where they started. Ending the loop matters
    // more than it looks: a rAF that never returns keeps a phone's display
    // pipeline awake for as long as the page is open, to animate nothing.
    if (!moving) return rest()

    frame = requestAnimationFrame(tick)
  }

  /**
   * The first frame of a burst has no previous one to be measured against, so
   * it sets the clock and does nothing else. Left to tick, it would measure the
   * gap since the last burst ended — a pause of any length reading as one
   * enormous frame.
   */
  function first(at: number) {
    stamp = at
    frame = requestAnimationFrame(tick)
  }

  function onScroll() {
    if (frame) return
    collect()
    frame = requestAnimationFrame(first)
  }

  onMounted(() => {
    // Asked not to animate, this does not start at all. There is nothing to
    // degrade to — the labels simply stay where they are drawn, which is where
    // they were before any of this.
    if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) return

    last = window.scrollY
    window.addEventListener('scroll', onScroll, { passive: true })
  })

  onBeforeUnmount(() => {
    window.removeEventListener('scroll', onScroll)
    if (frame) cancelAnimationFrame(frame)
    frame = 0
  })
}
