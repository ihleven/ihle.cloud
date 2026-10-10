/**
 * Where this page goes back to, if anywhere.
 *
 * One fact, asked by two: the strip renders the link, and the layout decides
 * whether the strip may be hidden at all on a narrow screen.
 *
 * It matters more than it looks. A content entry is shown full-bleed and asks
 * for no round button, so this link is the only way out of it — and an
 * installed app has no browser chrome to fall back on: no address bar, no back
 * button, and on iOS no pull-to-reload. A page with no way out is a dead end
 * rather than a detour.
 *
 * The answer is the area this path sits inside, rather than a list of the
 * places that happen to have children. `/filme/schmalfilm-1974` goes back to
 * `/filme`; so does anything else one level under an area, without this having
 * to be told that the area exists. An area's own root goes nowhere: it is a
 * destination, and the menu reaches it.
 */
export function useBackLink() {
  const route = useRoute()
  const { areas } = useAreas()

  const back = computed(() => {
    const here = route.path

    // The longest match, so an area nested under another is preferred to the
    // one above it rather than whichever was declared first.
    const area = areas.value
      .filter(a => a.at !== '/' && here.startsWith(a.at + '/'))
      .sort((a, b) => b.at.length - a.at.length)[0]

    return area ? { to: area.at, label: area.label } : null
  })

  return { back }
}
