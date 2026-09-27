// A page may declare itself reachable without a session. app.vue gates the whole
// app on one, so a page that is meant to be opened by someone not signed in —
// an enrollment link, a public film — has to say so.
// A page may also drop parts of the shell its layout would otherwise give it.
// The layout decides which parts exist at all; this turns off ones it supports.
// A layout without a footer simply ignores `footer: false`.
//
// definePageMeta is a compile-time macro, so these have to be written out
// literally — they cannot depend on the session. Anything conditional on who is
// signed in belongs in the layout, as geheimtipp's footer does.
interface ShellParts {
  /**
   * The strip along the top: the wordmark and, from lg, the areas.
   *
   * It scrolls away with the page by default, because iOS hazes the top of
   * scrolling content and nothing a page can write prevents it — so anything
   * anchored there is permanently under it. A page that would rather keep the
   * strip in reach than avoid that band says `'fixed'`, and it then stays put
   * while the page scrolls. `'static'` is the default spelled out.
   *
   * The round menu button shares that band and follows this unless `menu` says
   * otherwise: the two leaving at different moments is what made one band look
   * like two decisions.
   */
  bar?: boolean | 'fixed' | 'static'
  footer?: boolean
  /**
   * The main menu: the round button in the corner and what it opens.
   *
   * This is AppMenu, which the layout renders. The other menu — the slideover of
   * areas — belongs to the bar along the foot of the screen, which is shown
   * below lg and carries its own way in.
   *
   * Where it sits follows `bar`, because the two share the band at the top: by
   * default both scroll away with the page, and a page that pins the strip pins
   * this with it. iOS hazes that band and nothing a page can write prevents it,
   * which is why leaving is the default.
   *
   * `'fixed'` or `'static'` decides for the button alone, for the pages where
   * the two genuinely differ — the reader has no strip at all and still wants
   * the way out to leave with the page.
   *
   * Only the button moves. What it opens stays fixed over the viewport — an
   * overlay that scrolled with the page behind it would be no overlay at all.
   */
  menu?: boolean | 'fixed' | 'static'
  /**
   * The band between the strip and the page: how tall, as any CSS length.
   *
   * It is a gradient, from the strip's own background down to the page's, so
   * the two meet by fading rather than by a line. What it is for is a page whose
   * content begins hard against the strip — the front page's wall of tiles does
   * — where the strip and the first row read as one block of colour and the eye
   * cannot find the join.
   *
   * `false` or `'0'` removes it. It is drawn only where there is a strip to
   * fade from, so a page with `bar: false` has no band whatever this says.
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
    public?: boolean
    shell?: ShellParts
  }
}

export {}
