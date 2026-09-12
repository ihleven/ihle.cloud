// Keeps areas the account is not entitled to out of direct navigation.
//
// A UX guard, not a security boundary: it runs in the browser and is bypassable,
// and the API enforces its own access. It exists so a wrong link or an old
// bookmark lands somewhere that explains itself instead of on a page the person
// was never meant to be offered.
//
// Runs after auth.global (alphabetical order), so the session is already loaded.
const AREA_BY_PREFIX: Record<string, string> = {
  '/filme': 'filme',
  '/entries': 'content',
  '/famihlie': 'familie',
  '/kalender': 'kalender',
  '/mediathek': 'mediathek',
  '/musik': 'musik',
  '/search': 'search',
}

export default defineNuxtRouteMiddleware((to) => {
  const prefix = Object.keys(AREA_BY_PREFIX).find(p => to.path === p || to.path.startsWith(p + '/'))
  if (!prefix) return

  const { session } = useAuth()
  const { may } = useModules()

  // Not signed in at all: the sign-in page is the answer, not a refusal.
  if (!session.value) {
    return navigateTo(`/login?redirect=${encodeURIComponent(to.fullPath)}`)
  }

  const area = AREA_BY_PREFIX[prefix]!
  if (!may(area)) {
    // The app's own error page, rather than a redirect to a made-up route: this
    // is a refusal, so it should read as one and carry the status.
    return abortNavigation(createError({
      statusCode: 403,
      statusMessage: `Kein Zugriff auf «${area}».`,
      fatal: true,
    }))
  }
})
