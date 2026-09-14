// The pool's own gate.
//
// Global, but only ever acts on /geheimtipp — the same shape as modules.global,
// and the reason it is not a named middleware is that a page added later would
// silently be ungated until someone remembered to name it.
//
// The two sign-ins are now one. An account here is how someone reaches the pool:
// the proxy mints the pool's own credential from the session, so being signed in
// to this app *is* being signed in to the pool, and that is the whole of what
// this decides.
//
// It does two jobs and they are not the same job. Deciding access is one; the
// other is loading the pool's landing payload, which nothing else does. Reading
// the second as an expensive way of doing the first is a mistake worth naming,
// because making it leaves every page asking for a URL with the edition missing
// out of the middle.
//
// The pages stay `public: true`. That is about app.vue's overlay, not about
// access: without it every visitor to the pool would meet this app's sign-in
// dialog laid over the pool's pages, including the ones they are welcome to
// read.
//
// A UX guard, not a security boundary: it runs in the browser. The pool's
// backend checks its own token on every write.
const PREFIX = '/geheimtipp'
const LOGIN = '/geheimtipp/login'

export default defineNuxtRouteMiddleware(async (to) => {
  // Generation writes each route as a directory with an index.html, so a
  // browser can arrive at either "/geheimtipp/login" or "/geheimtipp/login/".
  // Comparing the raw path lets the second one past the exemption below and
  // sends the sign-in page to itself.
  const path = to.path.length > 1 ? to.path.replace(/\/+$/, '') : to.path

  if (path !== PREFIX && !path.startsWith(PREFIX + '/')) return
  if (path === LOGIN) return

  // Access is decided here, from this app's session alone. The pool cannot see
  // anyone it was not told about: the proxy forwards no token but the one it
  // mints, and it mints only for an account entitled to the pool. So asking the
  // pool who is signed in could only ever repeat this answer.
  const { session } = useAuth()
  if (!session.value?.modules.includes('geheimtipp')) {
    // Where they were going, so signing in lands there rather than on the
    // landing page — the link that brought them here is usually the point.
    return navigateTo({ path: LOGIN, query: { weiter: to.fullPath } })
  }

  // Then the payload, for everyone let through. /aktuell is not a session probe
  // to be skipped once the session is known — it carries the current edition,
  // this person's registration and the matchday windows, and it is the only
  // thing that loads them. Pages build their own URLs out of the edition, so
  // skipping this leaves them asking for `/compseasons/BL/` with no season.
  await useGhtSession().load()
})
