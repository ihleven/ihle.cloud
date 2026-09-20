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
  /** The strip along the top: the wordmark, the areas, the menu button. */
  bar?: boolean
  footer?: boolean
  /**
   * The main menu: the round button in the corner and what it opens.
   *
   * This is AppMenu, which the layout renders. The other menu — the one in the
   * top strip's slideover — comes with the strip, so it is governed by `bar`.
   */
  menu?: boolean
}

declare module '#app' {
  interface PageMeta {
    public?: boolean
    shell?: ShellParts
  }
}

export {}
