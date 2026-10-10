/**
 * What chrome this page asks for.
 *
 * Read in two places and so no longer owned by either. The layout draws the
 * page's own furniture — the strip, the band, the background, the footer — and
 * `app.vue` draws the shell, which is everything that must not be rebuilt on
 * the way from one layout to the next.
 *
 * A page keeps all of it unless it says otherwise, in `definePageMeta`. What it
 * can turn off is here; what it cannot is anything depending on who is signed
 * in, because page meta is resolved at build time — that belongs in a layout,
 * or in a layer of its own, as geheimtipp's does.
 */
/**
 * What the top of the screen is: a colour, and what can be read against it.
 *
 * Light ink is white and dark ink is black — two values rather than a free
 * colour, because the only thing read up there is a chevron and a menu button,
 * and a third option would be a decision nobody is asking to make.
 */
export type Top = {
  colour: string
  ink: 'light' | 'dark'
  /**
   * Whether the header leaves with the page instead of staying at the top.
   *
   * For a page that is one picture and a list under it: the film takes the top
   * of the screen, and what is left up there is the film rather than a bar over
   * it. The way back and the menu are a scroll away, which on such a page is
   * the trade worth making.
   */
  scrolls?: boolean
}

export function useShell() {
  const route = useRoute()

  const shell = computed(() => ({
    footer: true,
    // The strip across the top of a narrow screen, and the document behind the
    // page. Together, on an installed iPhone, they are the status bar: iOS takes
    // its colour from the pixels the page paints there, from the strip where the
    // page has been scrolled and from the document where it has not.
    //
    // The ink is stated rather than worked out, so that choosing a colour means
    // looking at what has to be read against it. iOS picks its own indicators
    // from the same surface and will not be told, which is why a mid-tone is a
    // poor choice here whatever we write.
    top: { colour: 'black', ink: 'light' } as Top,
    // How big the round button in the corner is. The middle size everywhere
    // unless a page says otherwise — a page that is mostly picture can afford a
    // smaller one, and one that is mostly nothing can carry a larger.
    menuSize: 'md' as MenuSize,
    // The colour the pages have always been. Named here rather than left to the
    // <main> element's class, because the band has to fade to exactly it and two
    // places saying what the page's colour is would eventually disagree.
    background: 'var(--color-ral-7004)',
    band: '2rem' as string | false,
    // The strip keeps its own colour for the first part of the band. Fading from
    // the very first pixel leaves the strip without an edge — the colour is
    // already going where the eye is still looking for one.
    bandHold: '40%',
    ...(route.meta.shell ?? {}),
  }))

  return { shell }
}
