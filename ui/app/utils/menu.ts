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
 * The circle grows from the button, which sits at the panel's top right, so
 * the furthest corner is always the bottom left one — this is the panel's own
 * diagonal, give or take the button's inset from the corner.
 *
 * Taken from the button's measured position rather than from the lengths it is
 * positioned with, so a safe-area inset is accounted for without this having to
 * know that safe areas exist.
 */
export function discRadius(panel: Box, button: Box, gutter = 24): number {
  const x = (button.left + button.right) / 2
  const y = (button.top + button.bottom) / 2

  return Math.hypot(x - panel.left, panel.bottom - y) + gutter
}
