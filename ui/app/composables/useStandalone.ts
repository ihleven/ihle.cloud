// Whether the app is running as an installed app rather than in a browser tab.
//
// It matters because an installed app has no browser chrome: no address bar, no
// back button, and on iOS no pull-to-reload. A tab can recover from anything by
// going back or reloading; an installed app cannot, so anything that would
// navigate away from it is a dead end rather than a detour.
//
// Read from the display mode rather than from the platform, because the
// condition is the missing chrome and not the operating system — an installed
// app on Android or the desktop is in the same position.
//
// `navigator.standalone` is checked as well: iOS has reported installed apps
// that way since long before it supported the display-mode query, and a phone
// that predates the query is exactly the one most likely to strand somebody.
export function useStandalone() {
  const standalone = ref(false)

  onMounted(() => {
    standalone.value
      = window.matchMedia?.('(display-mode: standalone)').matches
        || (window.navigator as Navigator & { standalone?: boolean }).standalone === true
  })

  /**
   * Open something that is not a page of this app.
   *
   * In a browser tab this does nothing: the link works as written, opening a
   * tab the visitor can close. Installed, the same link would replace the app
   * with the document and leave no way back, so it is handed to the browser
   * instead — which puts the document in a window of its own and leaves the app
   * where it was.
   *
   * The default is only suppressed once the browser has actually taken it. If
   * opening is refused the link is left to do what it would have done, because
   * a link that does nothing at all is worse than one that navigates.
   *
   * `noopener` is deliberately not in the feature string: window.open returns
   * null whenever it is given, by specification and not by accident, so asking
   * for it makes the test above answer "refused" every single time and the
   * whole guard a no-op. The opener is severed afterwards instead, which is the
   * same protection and leaves the return value meaning what it says.
   */
  function openOutside(event: MouseEvent, href: string): void {
    if (!standalone.value) return

    const opened = window.open(href, '_blank')
    if (!opened) return

    opened.opener = null
    event.preventDefault()
  }

  return { standalone, openOutside }
}
