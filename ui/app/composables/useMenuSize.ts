/**
 * How big the round menu button is, and everything that has to agree with it.
 *
 * One table, because three places depend on these numbers and two of them are
 * not the button: the strip leaves a hole for it so the areas do not run
 * underneath, and the rule under the strip stops short of it. Those used to be
 * written out in `Header` with a comment pointing at `Menu` — which is a
 * dependency that cannot be checked and was wrong the first time the button
 * changed size.
 *
 * The class names are written out rather than built: Tailwind reads the source
 * for candidates, and a name assembled at runtime is a name it never generates.
 */
export type MenuSize = 'sm' | 'md' | 'lg'

export type MenuMetric = {
  /** The button itself, and the mark inside it. */
  button: string
  glyph: string
  bar: string
  /** How far the outer bars sit from the centre, closed. */
  off: number
  /** The wordmark, kept level with the button: the two bracket the menu. */
  name: string
  /** What the menu's first row leaves to clear the button. */
  clear: string
  /** The hole the strip leaves in its row, the button's own width. */
  hole: string
  /**
   * Where the rule under the strip stops: the button, its inset, and a gap.
   * From lg only — below it the button is in the corner of the window rather
   * than in the strip's row, so the rule has nothing to stop short of.
   */
  rule: string
}

const metrics: Record<MenuSize, MenuMetric> = {
  sm: { button: 'size-10', glyph: 'h-4 w-5', bar: 'h-[2px]', off: 5, name: 'h-10', clear: '4.5rem', hole: 'w-10', rule: 'lg:right-[4rem]' },
  md: { button: 'size-12', glyph: 'h-[18px] w-6', bar: 'h-[3px]', off: 6, name: 'h-12', clear: '5rem', hole: 'w-12', rule: 'lg:right-[4.5rem]' },
  lg: { button: 'size-14', glyph: 'h-5 w-7', bar: 'h-[3px]', off: 7, name: 'h-14', clear: '5.5rem', hole: 'w-14', rule: 'lg:right-[5rem]' },
}

/**
 * The numbers for the size this page asked for.
 *
 * A page says it in `shell` meta, like everything else about the top of the
 * screen, and both the header and the menu read the answer here rather than
 * passing it between themselves. It used to be shared state nothing ever set,
 * so only the middle size existed.
 */
export function useMenuSize() {
  const { shell } = useShell()
  const metric = computed(() => menuMetric(shell.value.menuSize))

  return { metric, sizes: Object.keys(metrics) as MenuSize[] }
}

/**
 * The numbers for one size, named.
 *
 * A function rather than the table itself: a `const` export sitting beside an
 * auto-imported composable is one more name in the same namespace, and the one
 * that wins is whichever Nuxt wrote last.
 */
export function menuMetric(size: MenuSize): MenuMetric {
  return metrics[size]
}
