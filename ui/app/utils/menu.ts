// The geometry of the main menu's background.
//
// The menu is backed by a circle grown out of the button that opens it. How
// big that circle has to be is a question about the content it stands behind,
// so it is answered here rather than written into the stylesheet as a length:
// the menu's height is the number of areas an account may see, and CSS cannot
// know that.

/** The part of a DOMRect this needs, so a DOMRect satisfies it as it is. */
export interface Box {
  left: number
  top: number
  right: number
  bottom: number
}

/**
 * How far the circle must reach from the button's centre to stand behind the
 * whole panel: the distance to the panel's furthest corner, plus a gutter so
 * the content is not left sitting on the edge.
 *
 * All four corners are measured rather than the one the button is diagonally
 * opposite. That used to be written down as a fact — the button sat at the
 * panel's top right, so the far corner was always the bottom left — and it
 * stopped being true the first time the button moved: below lg it is in the
 * bottom corner, and the panel is above it rather than below. Asking which
 * corner is furthest costs three more hypotenuses and cannot be wrong.
 *
 * Taken from the button's measured position rather than from the lengths it is
 * positioned with, so a safe-area inset is accounted for without this having to
 * know that safe areas exist.
 */
export function discRadius(panel: Box, button: Box, gutter = 24): number {
  const x = (button.left + button.right) / 2
  const y = (button.top + button.bottom) / 2

  const corners = [
    Math.hypot(x - panel.left, y - panel.top),
    Math.hypot(x - panel.left, y - panel.bottom),
    Math.hypot(x - panel.right, y - panel.top),
    Math.hypot(x - panel.right, y - panel.bottom),
  ]

  return Math.max(...corners) + gutter
}
