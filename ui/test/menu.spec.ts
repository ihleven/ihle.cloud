import { describe, expect, it } from 'vitest'
import { discRadius, type Box } from '../app/utils/menu'

// The real geometry: a 3.5rem button inset 1rem from the top right, and a panel
// flush with the right edge running from the top of the screen to wherever its
// content ends.
function button(viewportWidth: number, safeTop = 0): Box {
  const left = viewportWidth - 16 - 56
  const top = safeTop + 16

  return { left, top, right: left + 56, bottom: top + 56 }
}

function panel(viewportWidth: number, width: number, height: number): Box {
  return { left: viewportWidth - width, top: 0, right: viewportWidth, bottom: height }
}

/** Does a circle of this radius, centred on the button, hold the panel's far corner? */
function holds(radius: number, p: Box, b: Box) {
  const x = (b.left + b.right) / 2
  const y = (b.top + b.bottom) / 2

  return Math.hypot(x - p.left, p.bottom - y) <= radius
}

describe('discRadius', () => {
  // Every shape the app is actually used in. The point of measuring rather
  // than writing a length is that this holds in all of them without anyone
  // having to pick a number that is generous enough for the worst.
  const screens: Array<[string, number, number, number]> = [
    // name, viewport width, panel width, panel height
    ['phone portrait', 390, 390, 477],
    ['phone landscape', 844, 273, 477],
    ['tablet portrait', 1024, 672, 477],
    ['laptop', 1440, 630, 477],
    ['desktop', 1920, 672, 477],
  ]

  it.each(screens)('stands behind the whole panel on a %s', (_name, vw, pw, ph) => {
    const b = button(vw)
    const p = panel(vw, pw, ph)

    expect(holds(discRadius(p, b), p, b)).toBe(true)
  })

  // The case a fixed 90vh radius could not serve: a phone held sideways is
  // only 390px tall, so 90vh is 351px, while the menu still needs to reach
  // more than 400px to stand behind its own areas.
  it('reaches further than 90vh would on a short landscape screen', () => {
    const b = button(844)
    const p = panel(844, 273, 477)

    expect(discRadius(p, b)).toBeGreaterThan(0.9 * 390)
    expect(holds(0.9 * 390, p, b)).toBe(false)
  })

  // An account entitled to four areas has a shorter menu than one entitled to
  // eleven, and gets a smaller circle for it.
  it('shrinks with the menu', () => {
    const b = button(1440)
    const few = panel(1440, 630, 265)
    const many = panel(1440, 630, 477)

    expect(discRadius(few, b)).toBeLessThan(discRadius(many, b))
    expect(holds(discRadius(few, b), few, b)).toBe(true)
  })

  // The gutter is what keeps the corner of the content off the curve.
  it('leaves the content clear of the edge', () => {
    const b = button(1440)
    const p = panel(1440, 630, 477)
    const x = (b.left + b.right) / 2
    const y = (b.top + b.bottom) / 2

    expect(discRadius(p, b, 24) - Math.hypot(x - p.left, p.bottom - y)).toBeCloseTo(24)
    expect(discRadius(p, b, 0)).toBeLessThan(discRadius(p, b, 24))
  })

  // The safe area moves the button down, and the circle grows from the button,
  // so the radius has to come off its measured position rather than off the
  // lengths it is positioned with.
  it('follows the button under a safe-area inset', () => {
    const p = panel(390, 390, 477)

    expect(discRadius(p, button(390, 59))).toBeLessThan(discRadius(p, button(390, 0)))
  })
})
