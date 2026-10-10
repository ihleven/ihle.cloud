// A page may declare that this app does not guard it. app.vue gates everything
// on a session, so a page reached without one — an enrolment link, the pool,
// which has an account system of its own — has to say so.
//
// Not "public": the pool is not open to the world, every page of it sends you
// to its own login, and that word was read here once as meaning it was and
// written into a plan as fact.
// A page may also say what the shell should look like where it is. Whether a
// part exists at all is the layout's: a layout without a footer simply ignores
// `footer: false`, and a layout that draws no header cannot be given one here.
//
// definePageMeta is a compile-time macro, so these have to be written out
// literally — they cannot depend on the session. Anything conditional on who is
// signed in belongs in the layout, as geheimtipp's footer does.
interface ShellParts {
  /**
   * The top of the screen: the colour of it, and what can be read against it.
   *
   * The header below lg is a fixed strip, and on an installed iPhone it is the
   * status bar — iOS takes the bar's colour from the pixels the page paints
   * there. It samples the strip where the page has been scrolled and the
   * document behind the page where it has not, so both are painted from this
   * one value and a page cannot set them apart.
   *
   * The ink is stated rather than worked out, so that choosing a colour means
   * looking at what has to be read against it. iOS picks its own indicators
   * from the same surface and will not be told, which is why a mid-tone is a
   * poor choice here whatever is written.
   *
   * Any CSS colour, the palette included: `var(--color-blue-800)`.
   */
  top?: {
    colour: string
    ink: 'light' | 'dark'
    /**
     * Whether the header leaves with the page instead of staying at the top.
     *
     * For a page that is one picture and a list under it — a film, an episode:
     * the picture sticks to the top of the window and what is up there is the
     * picture rather than a bar over it. The way back and the menu are a scroll
     * away, which on such a page is the trade worth making.
     *
     * A page that asks for this should stick its own picture at
     * `top-(--status-bar)` rather than at `top-(--top-height)`: the row is
     * gone, but the band under the system's own clock and battery stays, and
     * a picture that rose past it would leave the page showing through.
     */
    scrolls?: boolean
  }
  /**
   * How big the round menu button is: 'sm', 'md' or 'lg'.
   *
   * What it opens does not change. The menu is laid out against the screen and
   * the circle behind it is measured from whatever the button turns out to be,
   * so this is a statement about the corner of the page rather than about the
   * menu — and the header reads it too, because it leaves the button a hole and
   * stops its rule short of it.
   */
  menuSize?: 'sm' | 'md' | 'lg'
  footer?: boolean
  /**
   * The band between the strip and the page: how tall, as any CSS length.
   *
   * It is a gradient, from the strip's own background down to the page's, so
   * the two meet by fading rather than by a line. What it is for is a page whose
   * content begins hard against the strip — the front page's wall of tiles does
   * — where the strip and the first row read as one block of colour and the eye
   * cannot find the join.
   *
   * `false` or `'0'` removes it. It hangs under the wide rendering of the
   * header only; below lg there is no band.
   *
   * Any CSS length: '2rem', '5vh', '48px'.
   */
  band?: string | false
  /**
   * The page's own background, and by default the colour the band arrives at.
   *
   * One value for both: a band that faded to something the page is not would
   * put back the line it exists to remove.
   *
   * Any CSS colour, which includes the palette — `var(--color-ral-7035)` — so a
   * page can name a colour this app already has rather than repeat its hex.
   */
  background?: string
  /**
   * Where the band arrives, when that is not the page's own background.
   *
   * Most pages never set this: the band meets the page, and the page is one
   * colour. A page whose *first block* is a colour of its own is the exception —
   * the calendar opens with a green band saying what day it is, and a fade that
   * arrived at the grey behind it would put a line between the two. What the
   * band has to meet is whatever is directly under it, which is not always what
   * the page is made of.
   *
   * Set this rather than lying with `background`. They are separate because the
   * page's background is also what <main> is painted, so a page claiming to be
   * green to get a green fade would turn green the day somebody removed the
   * block that was covering it.
   */
  bandTo?: string
  /**
   * How much of the band keeps the strip's own colour before the fade starts.
   *
   * A gradient that begins fading immediately reads as a smear: the strip has
   * no edge, and the colour is already changing where the eye is still looking
   * for one. Holding the strip's colour for the first part of the band gives it
   * a shoulder to sit on, and the fade then has less distance to cover, which
   * makes it look deliberate rather than gradual.
   *
   * Any CSS length or percentage; a percentage is of the band's own height, so
   * it keeps its proportions whatever the band is set to.
   */
  bandHold?: string
  /**
   * A third colour, part of the way down.
   *
   * Two colours fade in a straight line through whatever lies between them,
   * which for white to a saturated green runs through a grey-green nobody
   * chose. A stop in the middle says what it should pass through instead.
   *
   * A CSS gradient stop, so it may carry its own position — `'var(--color-green-200) 70%'`
   * puts that colour seventy per cent of the way down. Without one the browser
   * spaces it evenly, which is usually what is wanted.
   */
  bandVia?: string
}

declare module '#app' {
  interface PageMeta {
    unguarded?: boolean
    shell?: ShellParts
  }
}

export {}
