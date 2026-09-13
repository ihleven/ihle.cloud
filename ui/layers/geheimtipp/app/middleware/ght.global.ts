// The pool's own gate.
//
// Global, but only ever acts on /geheimtipp — the same shape as modules.global,
// and the reason it is not a named middleware is that a page added later would
// silently be ungated until someone remembered to name it.
//
// This is deliberately not the family app's sign-in. The pool has its own users
// in its own database, and an account here grants nothing there. The pages stay
// `public: true` so app.vue's overlay never covers them, and this stands in its
// place.
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

  const { signedIn, load } = useGhtSession()
  await load()
  if (signedIn.value) return

  // Where they were going, so signing in lands there rather than on the
  // landing page — the link that brought them here is usually the point.
  return navigateTo({ path: LOGIN, query: { weiter: to.fullPath } })
})
