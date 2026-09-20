// Keeps areas the account is not entitled to out of direct navigation.
//
// A UX guard, not a security boundary: it runs in the browser and is bypassable,
// and the API enforces its own access. It exists so a wrong link or an old
// bookmark lands somewhere that explains itself instead of on a page the person
// was never meant to be offered.
//
// Which path belongs to which area is utils/modules, so a new area is guarded
// by existing — the archive was reachable unguarded for months because this
// file kept its own copy of the list and nobody added /retro to it.
//
// Runs after auth.global (alphabetical order), so the session is already loaded.
export default defineNuxtRouteMiddleware((to) => {
  const module = moduleAt(to.path)
  if (!module) return

  const { session } = useAuth()
  const { may } = useModules()

  // Not signed in at all is not this guard's business: app.vue shows the
  // sign-in overlay over whatever route was asked for, and reveals that same
  // route once the person signs in. There is nothing to redirect to.
  if (!session.value) return

  if (!may(module.id)) {
    // The app's own error page, rather than a redirect to a made-up route: this
    // is a refusal, so it should read as one and carry the status.
    return abortNavigation(createError({
      statusCode: 403,
      statusMessage: `Kein Zugriff auf «${module.id}».`,
      fatal: true,
    }))
  }
})
