import { ref } from 'vue'
import { beforeEach, describe, expect, it } from 'vitest'

// The composable reaches for Nuxt's auto-imports and for the browser. Standing
// both up here rather than mocking the composable means these tests run the
// spring that ships, a frame at a time, with the frames under our control.
// eslint-disable-next-line @typescript-eslint/no-explicit-any
const g = globalThis as any

let mounted: (() => void)[] = []
let unmounted: (() => void)[] = []
let frames: ((at: number) => void)[] = []
let clock = 0
let listeners: Record<string, (() => void)[]> = {}
let reduced = false

g.onMounted = (fn: () => void) => mounted.push(fn)
g.onBeforeUnmount = (fn: () => void) => unmounted.push(fn)
g.requestAnimationFrame = (fn: (at: number) => void) => {
  frames.push(fn)

  return frames.length
}
g.cancelAnimationFrame = () => {}
g.window = {
  scrollY: 0,
  matchMedia: () => ({ matches: reduced }),
  addEventListener: (name: string, fn: () => void) => {
    (listeners[name] ||= []).push(fn)
  },
  removeEventListener: (name: string, fn: () => void) => {
    listeners[name] = (listeners[name] || []).filter(f => f !== fn)
  },
}

const { useScrollGravity } = await import('../app/composables/useScrollGravity')

interface FakeLabel { style: { setProperty: (k: string, v: string) => void }, props: Record<string, string> }

/**
 * A stand-in for the grid and the labels inside it, each remembering what was
 * last written to it. The composable finds the labels by querying, so the fake
 * has to answer a query.
 */
function grid(labels = 4) {
  const kids: FakeLabel[] = Array.from({ length: labels }, () => {
    const props: Record<string, string> = {}
    const setProperty = (k: string, v: string) => {
      props[k] = v
    }

    return { style: { setProperty }, props }
  })

  return ref({ querySelectorAll: () => kids, kids } as never) as never
}

/** What each label is currently displaced by, in pixels. */
function gravityOf(el: never): number[] {
  const { kids } = (el as unknown as { value: { kids: FakeLabel[] } }).value

  return kids.map(k => parseFloat(k.props['--gravity'] ?? '0'))
}

/** The displacement of the first label, for the tests about one spring. */
function firstOf(el: never): number {
  return gravityOf(el)[0] ?? 0
}

/**
 * Run the frames the loop has asked for, moving the page as we go.
 *
 * ms is how long each frame takes: 16.667 is a sixtieth of a second, 8.333 a
 * hundred-and-twentieth. The spring is meant to feel the same either way, which
 * is a thing this can ask it.
 */
function run(times: number, scrollBy = 0, ms = 16.667) {
  for (let i = 0; i < times; i++) {
    g.window.scrollY += scrollBy
    clock += ms
    const due = frames
    frames = []
    // Only one frame is ever outstanding; draining is how a burst is avoided.
    due.forEach(fn => fn(clock))
    if (!due.length) break
  }
}

function scroll() {
  (listeners.scroll || []).forEach(fn => fn())
}

describe('scroll gravity', () => {
  beforeEach(() => {
    mounted = []
    unmounted = []
    frames = []
    listeners = {}
    reduced = false
    clock = 0
    g.window.scrollY = 0
  })

  function start(options = {}, labels = 4) {
    const el = grid(labels)
    useScrollGravity(el, options)
    mounted.forEach(fn => fn())

    return el
  }

  // The direction asked for: scrolling towards the top of the page throws the
  // labels down, the way the content itself appears to travel.
  it('throws the labels downwards when the page is scrolled towards the top', () => {
    g.window.scrollY = 2000
    const el = start()
    scroll()
    run(4, -40)

    expect(firstOf(el)).toBeGreaterThan(0)
  })

  it('throws them the other way when the page is scrolled away from the top', () => {
    const el = start()
    scroll()
    run(4, 40)

    expect(firstOf(el)).toBeLessThan(0)
  })

  // Speed is the whole input: a flick should move them further than a nudge.
  it('throws them further the faster the page moves', () => {
    g.window.scrollY = 5000
    const gentle = start()
    scroll()
    run(4, -10)
    const small = firstOf(gentle)

    beforeEachReset()
    g.window.scrollY = 5000
    const hard = start()
    scroll()
    run(4, -120)

    expect(firstOf(hard)).toBeGreaterThan(small)
  })

  // However hard it is flung. A label that leaves its tile is not weight, it
  // is a bug with a spring in it.
  it('never throws them past the limit', () => {
    g.window.scrollY = 90000
    const el = start({ max: 14 })
    scroll()
    run(14, -4000)

    expect(Math.max(...gravityOf(el).map(Math.abs))).toBeLessThanOrEqual(14)
  })

  // The part the whole thing is for: it comes back on its own, and then stops
  // asking for frames. A loop that never ends keeps a phone's display awake to
  // animate nothing.
  it('settles back to nothing and stops running', () => {
    g.window.scrollY = 2000
    const el = start()
    scroll()
    run(4, -60)
    expect(Math.abs(firstOf(el))).toBeGreaterThan(0.5)

    run(400, 0) // the page is still; let it settle

    expect(gravityOf(el)).toEqual([0, 0, 0, 0])
    expect(frames).toHaveLength(0)
  })

  // A phone and a laptop both draw at 120Hz now. Counted in frames rather than
  // in time, the same gesture covers half the distance between frames there and
  // the whole effect is half as strong on the better screen.
  it('feels the same on a 120Hz screen as on a 60Hz one', () => {
    g.window.scrollY = 5000
    const sixty = start()
    scroll()
    run(4, -48, 16.667)
    const atSixty = firstOf(sixty)

    beforeEachReset()
    clock = 0
    g.window.scrollY = 5000
    const oneTwenty = start()
    scroll()
    // Same gesture, same wall-clock time: half the distance, twice the frames.
    run(8, -24, 8.333)

    // Within a fifth, not to the pixel. A spring integrated a step at a time
    // does not land in exactly the same place when the steps are halved, and
    // chasing that would mean a solver rather than four lines of arithmetic —
    // for a difference of half a pixel nobody can see. What this is here to
    // catch is the factor of two that counting frames instead of time gives,
    // which is the difference between the effect being there and not.
    const ratio = firstOf(oneTwenty) / atSixty
    expect(ratio).toBeGreaterThan(0.8)
    expect(ratio).toBeLessThan(1.2)
  })

  // The point of the change: twelve labels moving exactly together read as one
  // slab sliding about behind the tiles, because the eye takes common motion
  // for one object.
  it('throws the tiles by different amounts', () => {
    g.window.scrollY = 5000
    const el = start({}, 6)
    scroll()
    run(4, -80)

    const thrown = gravityOf(el)
    expect(new Set(thrown.map(n => n.toFixed(2))).size).toBeGreaterThan(1)
  })

  // ...and asked for none, they are identical again, which is the setting to
  // reach for if the ripple turns out to be a distraction.
  it('throws them all the same when no spread is asked for', () => {
    g.window.scrollY = 5000
    const el = start({ spread: 0 }, 6)
    scroll()
    run(4, -80)

    const thrown = gravityOf(el)
    expect(new Set(thrown.map(n => n.toFixed(2))).size).toBe(1)
  })

  // Strength only. They are pulled home equally hard, so a lighter label does
  // not also linger — it simply went less far.
  it('settles every tile together, however far each went', () => {
    g.window.scrollY = 5000
    const el = start({}, 6)
    scroll()
    run(4, -80)
    expect(new Set(gravityOf(el).map(n => n.toFixed(2))).size).toBeGreaterThan(1)

    run(400, 0)
    expect(gravityOf(el)).toEqual([0, 0, 0, 0, 0, 0])
  })

  it('does not start at all when motion is not wanted', () => {
    reduced = true
    start()

    expect(listeners.scroll ?? []).toHaveLength(0)
  })

  it('lets go of the page when the component does', () => {
    start()
    expect(listeners.scroll).toHaveLength(1)

    unmounted.forEach(fn => fn())
    expect(listeners.scroll).toHaveLength(0)
  })
})

function beforeEachReset() {
  mounted = []
  unmounted = []
  frames = []
  listeners = {}
}
