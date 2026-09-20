import { describe, expect, it } from 'vitest'
import { appModules, moduleAt, navModules, splitForBar, tileModules } from '../app/utils/modules'

describe('moduleAt', () => {
  it('finds the module a path belongs to', () => {
    expect(moduleAt('/musik')?.id).toBe('musik')
    expect(moduleAt('/musik/irgendwas')?.id).toBe('musik')
    expect(moduleAt('/djvet/pool-party/001')?.id).toBe('djvet')
  })

  // A module owns its own path and what is under it, and nothing else. The
  // guard drew this line correctly before; stating it here keeps it drawn now
  // that one function serves the guard, the navigation and the shell.
  it('does not claim a path that merely starts with the same letters', () => {
    expect(moduleAt('/musikschule')).toBeUndefined()
    expect(moduleAt('/admins')).toBeUndefined()
  })

  it('leaves common ground to nobody', () => {
    expect(moduleAt('/')).toBeUndefined()
    expect(moduleAt('/enroll')).toBeUndefined()
  })

  // The bug this registry exists to prevent: /retro was in the navigation and
  // in the tiles but missing from the guard, so the archive was unguarded.
  it('covers every module the navigation offers', () => {
    for (const module of navModules()) {
      expect(moduleAt(module.at)?.id).toBe(module.id)
    }
  })
})

describe('the registry', () => {
  it('gives every module an id, a label and a path', () => {
    for (const module of appModules) {
      expect(module.id).toBeTruthy()
      expect(module.label).toBeTruthy()
      expect(module.at.startsWith('/')).toBe(true)
    }
  })

  it('has no two modules at the same path', () => {
    const paths = appModules.map(m => m.at)
    expect(new Set(paths).size).toBe(paths.length)
  })

  // Presence of the field is what decides, so these are not the same list:
  // geheimtipp has a tile and no navigation entry, because it is reached from
  // the front page and then keeps its own chrome.
  it('separates what the navigation offers from what the front page shows', () => {
    expect(navModules().map(m => m.id)).not.toContain('geheimtipp')
    expect(tileModules().map(m => m.id)).toContain('geheimtipp')
    expect(tileModules().map(m => m.id)).not.toContain('admin')
  })
})

describe('splitForBar', () => {
  it('holds nothing back when everything fits', () => {
    const { shown, rest } = splitForBar(navModules().slice(0, 4))
    expect(shown).toHaveLength(4)
    expect(rest).toHaveLength(0)
  })

  it('keeps the first few in order and leaves the rest to the menu', () => {
    const { shown, rest } = splitForBar(navModules())
    expect(shown.map(m => m.id)).toEqual(navModules().slice(0, 4).map(m => m.id))
    expect(rest).toHaveLength(navModules().length - 4)
  })

  it('copes with an account entitled to nothing', () => {
    const { shown, rest } = splitForBar([])
    expect(shown).toHaveLength(0)
    expect(rest).toHaveLength(0)
  })

  it('copes with fewer than the bar has room for', () => {
    const { shown, rest } = splitForBar(navModules().slice(0, 3))
    expect(shown).toHaveLength(3)
    expect(rest).toHaveLength(0)
  })
})

// The art archive is registered for one reason only: so the route guard knows
// /werke belongs to somebody. Before it was listed here the page was reachable
// by any signed-in account, because the guard gates only the prefixes this
// registry names. Keeping it out of the navigation and the front page is what
// leaves it unadvertised — and an icon or a tile added later would undo that
// silently, which is what these assertions are for.
describe('the art archive', () => {
  it('claims /werke, so the guard covers it', () => {
    expect(moduleAt('/werke')?.id).toBe('art')
    expect(moduleAt('/werke/etwas')?.id).toBe('art')
  })

  it('is offered nowhere', () => {
    expect(navModules().map(m => m.id)).not.toContain('art')
    expect(tileModules().map(m => m.id)).not.toContain('art')
  })
})

// The journeys were put under the familie area rather than given a route of
// their own, so that they are covered by the familie entitlement instead of
// needing one of their own. That only holds while the path stays beneath it.
describe('the journeys', () => {
  it('are guarded as part of familie', () => {
    expect(moduleAt('/famihlie/reisen/2025-bretagne')?.id).toBe('familie')
  })
})
