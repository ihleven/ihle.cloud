// Laying the front page out as a wall rather than a table.
//
// A grid of equal squares reads as a control panel: every door the same size
// says every door is the same, and there is nothing for the eye to land on. A
// wall of mixed blocks reads as a place. What it costs is that the arrangement
// has to be *worked out* rather than written down — the tiles somebody sees
// follow from what they are entitled to, so a layout drawn by hand has a hole
// in it the moment one is missing.
//
// So the shapes are dealt out instead. Nothing here knows which tiles it is
// dealing with, only how many and what each of them will accept, which is what
// makes it survive a page with two tiles on it and a page with twelve.

/**
 * What a tile may be, in cells of the grid.
 *
 * Four rather than a free choice of width and height, because the gaps are the
 * point: shapes that all tessellate leave a solid slab, and a small vocabulary
 * of blocks that do not quite fit each other is what produces the holes.
 */
export type TileShape = 'small' | 'wide' | 'tall' | 'big'

/**
 * The grid spans each shape claims.
 *
 * Written out as whole class names rather than built from pieces, because
 * Tailwind finds the classes it must generate by reading the source: a name
 * assembled at runtime is a name it never sees, and the rule is simply absent.
 */
export const shapeClass: Record<TileShape, string> = {
  small: 'col-span-1 row-span-1',
  wide: 'col-span-2 row-span-1',
  tall: 'col-span-1 row-span-2',
  big: 'col-span-2 row-span-2',
}

/**
 * How often each shape comes up.
 *
 * Mostly small, and that is not timidity: a wall of large blocks is as
 * monotonous as a wall of equal squares, and the big ones only read as big
 * while they are outnumbered.
 */
const odds: [TileShape, number][] = [
  ['small', 6],
  ['wide', 2],
  ['tall', 2],
  ['big', 1],
]

/**
 * What a hole is drawn as.
 *
 * Hidden until there is room for three columns. A hole is air, and air is a
 * luxury of a wide window: on a phone the wall is two columns, so one hole is
 * half a row, the page opens on the wordmark beside a large grey nothing, and
 * everything below it is pushed another screen down. The shapes still vary
 * there — that is what makes it a wall — but the gaps between them wait for the
 * room to be worth it.
 */
export const gapClass = 'hidden md:block col-span-1 row-span-1'

/**
 * Roughly one hole for every this many tiles.
 *
 * The holes are what the exercise is for — they are the difference between a
 * wall and a slab — but they are also doors that are not there, so a page made
 * mostly of absence is worse than the grid it replaced.
 */
const perGap = 5

/**
 * At most one tile is allowed to be the big one.
 *
 * Two of them and the page has two focal points, which is none. It is also the
 * shape that hurts most on a phone, where the grid is two columns wide and a
 * big tile is the whole screen.
 */
const bigLimit = 1

/**
 * A deterministic pseudo-random sequence from one number.
 *
 * `Math.random()` cannot be asked to repeat itself, and every part of this has
 * to: the shapes and the order are two draws from one arrangement, and a test
 * that cannot fix the seed can only assert that something happened.
 */
export function rng(seed: number): () => number {
  let a = Math.floor(seed * 0xffffffff) || 1

  return () => {
    a |= 0
    a = (a + 0x6d2b79f5) | 0
    let t = Math.imul(a ^ (a >>> 15), 1 | a)
    t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t

    return ((t ^ (t >>> 14)) >>> 0) / 4294967296
  }
}

/**
 * One cell of the wall: a tile, or a hole.
 *
 * A hole has no tile and is rendered as nothing — the page's own background,
 * showing through. Left to itself the grid packs forward and leaves a gap only
 * when a wide block will not fit at the end of a row, which on a wide window
 * almost never happens: eleven blocks across nine columns come out as a solid
 * band. The holes have to be dealt, or there are none.
 */
export interface Laid<T> {
  tile?: T
  shape: TileShape
  /** The Tailwind spans for that shape, ready to put on the element. */
  class: string
  /** Stable within one arrangement, for keying the list. */
  key: string
}

/**
 * What a tile will accept, where it cannot take any shape.
 *
 * Most tiles are a word on a background and do not care. Some have something
 * inside them with a shape of its own — the calendar's date block wants height
 * and is squashed by width — and say so.
 */
export interface Shaped {
  shapes?: TileShape[]
}

/**
 * Deal out an arrangement: the tiles in a shuffled order, each with a shape.
 *
 * The order is shuffled rather than sorted, and the caller decides how often
 * that happens by choosing when the seed changes. Nothing is placed: the shapes
 * are spans on a CSS grid and the browser packs them, which is what keeps this
 * working when the window is resized and when there are three tiles instead of
 * ten.
 *
 * Gaps are left rather than filled. A shape that will not fit at the end of a
 * row pushes to the next one and leaves the remainder empty, and that hole is
 * the whole point — it is the difference between a wall and a slab. Asking the
 * grid to pack densely would close them all up again.
 */
export function mosaic<T extends Shaped>(tiles: T[], seed: number): Laid<T>[] {
  const next = rng(seed)

  // Fisher–Yates, on a copy: the list handed in belongs to the caller and is
  // very often a computed property, which must not be shuffled in place.
  const order = [...tiles]
  for (let i = order.length - 1; i > 0; i--) {
    const j = Math.floor(next() * (i + 1))
    const a = order[i] as T
    order[i] = order[j] as T
    order[j] = a
  }

  let bigs = 0
  const laid: Laid<T>[] = []

  order.forEach((tile, i) => {
    const allowed = tile.shapes ?? (['small', 'wide', 'tall', 'big'] as TileShape[])
    const pool = odds.filter(([shape]) =>
      allowed.includes(shape) && (shape !== 'big' || bigs < bigLimit))

    const shape = pick(pool, next) ?? 'small'
    if (shape === 'big') bigs++

    laid.push({ tile, shape, class: shapeClass[shape], key: `t${i}` })
  })

  // The holes, dropped in afterwards so that dealing them cannot use up a place
  // a tile needed. Never first — the wall should begin with something — and
  // never two together, which reads as the page having failed to load rather
  // than as space.
  const holes = Math.floor(laid.length / perGap)
  for (let h = 0; h < holes; h++) {
    const at = 1 + Math.floor(next() * (laid.length - 1))
    if (laid[at - 1]?.tile === undefined || laid[at]?.tile === undefined) continue
    laid.splice(at, 0, { shape: 'small', class: gapClass, key: `gap${h}` })
  }

  return laid
}

/** One weighted draw. */
function pick(pool: [TileShape, number][], next: () => number): TileShape | undefined {
  const total = pool.reduce((n, [, weight]) => n + weight, 0)
  if (!total) return undefined

  let roll = next() * total
  for (const [shape, weight] of pool) {
    roll -= weight
    if (roll <= 0) return shape
  }

  return pool[pool.length - 1]?.[0]
}
