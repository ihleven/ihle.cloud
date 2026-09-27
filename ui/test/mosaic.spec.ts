import { describe, expect, it } from 'vitest'

import { gapClass, mosaic, rng, shapeClass, type TileShape } from '../app/utils/mosaic'

interface Tile { id: string, shapes?: TileShape[] }

const ten: Tile[] = Array.from({ length: 10 }, (_, i) => ({ id: `t${i}` }))

/** Every seed that matters, so a property is asserted rather than one lucky roll. */
const seeds = Array.from({ length: 60 }, (_, i) => (i + 1) / 61)

/** The cells that hold a tile, as opposed to the holes between them. */
function tilesOf(laid: ReturnType<typeof mosaic<Tile>>) {
  return laid.filter(l => l.tile).map(l => l.tile as Tile)
}

describe('mosaic', () => {
  it('lays out every tile it is given, once each', () => {
    for (const seed of seeds) {
      const held = tilesOf(mosaic(ten, seed))
      expect(held).toHaveLength(ten.length)
      expect(new Set(held.map(t => t.id)).size).toBe(ten.length)
    }
  })

  // The list handed in is very often a computed property, and shuffling it in
  // place would reorder the thing the caller is deriving from.
  it('leaves the list it was given alone', () => {
    const before = ten.map(t => t.id)
    mosaic(ten, 0.42)
    expect(ten.map(t => t.id)).toEqual(before)
  })

  it('puts them in a different order for a different seed', () => {
    const one = tilesOf(mosaic(ten, 0.1)).map(t => t.id).join()
    const two = tilesOf(mosaic(ten, 0.9)).map(t => t.id).join()
    expect(one).not.toBe(two)
  })

  // The same number has to give the same page, or nothing downstream can hold
  // still: the arrangement is rolled once and read on every render after it.
  it('gives the same arrangement for the same seed', () => {
    const one = mosaic(ten, 0.37)
    const two = mosaic(ten, 0.37)
    expect(one.map(l => `${l.tile?.id ?? 'gap'}:${l.shape}`))
      .toEqual(two.map(l => `${l.tile?.id ?? 'gap'}:${l.shape}`))
  })

  it('gives out more than one shape', () => {
    const shapes = new Set(seeds.flatMap(seed => tilesOf(mosaic(ten, seed)).length
      ? mosaic(ten, seed).filter(l => l.tile).map(l => l.shape)
      : []))
    expect(shapes.size).toBeGreaterThan(1)
  })

  // Two big tiles is two focal points, which is none — and on a phone, where
  // the grid is two columns, a big one is already the whole screen.
  it('never deals more than one big tile', () => {
    for (const seed of seeds) {
      const bigs = mosaic(ten, seed).filter(l => l.tile && l.shape === 'big')
      expect(bigs.length).toBeLessThanOrEqual(1)
    }
  })

  // The calendar's date block wants height and is squashed by width.
  it('gives a tile only the shapes it will accept', () => {
    const fussy: Tile[] = [{ id: 'kalender', shapes: ['small', 'tall'] }]
    for (const seed of seeds) {
      const held = mosaic(fussy, seed).filter(l => l.tile)
      expect(['small', 'tall']).toContain(held[0]?.shape)
    }
  })

  it('hands back the spans for the shape it chose', () => {
    for (const laid of mosaic(ten, 0.5).filter(l => l.tile)) {
      expect(laid.class).toBe(shapeClass[laid.shape])
    }
  })

  // On a phone the wall is two columns, so a hole is half a row: the page would
  // open on the wordmark beside a large grey nothing.
  it('holds the holes back until there is room for them', () => {
    for (const hole of mosaic(ten, 0.5).filter(l => !l.tile)) {
      expect(hole.class).toBe(gapClass)
      expect(hole.class).toContain('hidden')
    }
  })

  // Holes are what separates a wall from a slab. Left to itself the grid packs
  // forward and leaves one only when a wide block will not fit at a row's end,
  // which on a wide window almost never happens.
  it('leaves holes in the wall', () => {
    for (const seed of seeds) {
      const holes = mosaic(ten, seed).filter(l => !l.tile)
      expect(holes.length).toBeGreaterThan(0)
    }
  })

  // ...but a page made mostly of absence is worse than the grid it replaced.
  it('leaves far more tiles than holes', () => {
    for (const seed of seeds) {
      const laid = mosaic(ten, seed)
      const holes = laid.filter(l => !l.tile).length
      expect(holes).toBeLessThanOrEqual(Math.ceil(ten.length / 4))
    }
  })

  // Two together read as the page having failed to load rather than as space,
  // and one at the front is a wall that begins with nothing.
  it('never opens with a hole or puts two side by side', () => {
    for (const seed of seeds) {
      const laid = mosaic(ten, seed)
      expect(laid[0]?.tile).toBeDefined()
      laid.forEach((cell, i) => {
        if (!cell.tile && i > 0) expect(laid[i - 1]?.tile).toBeDefined()
      })
    }
  })

  // A page with nothing on it is a page, not a crash.
  it('copes with an empty page', () => {
    expect(mosaic([], 0.5)).toEqual([])
  })

  it('draws numbers in range and repeats them for one seed', () => {
    const a = rng(0.25)
    const b = rng(0.25)
    for (let i = 0; i < 50; i++) {
      const n = a()
      expect(n).toBeGreaterThanOrEqual(0)
      expect(n).toBeLessThan(1)
      expect(n).toBe(b())
    }
  })
})
