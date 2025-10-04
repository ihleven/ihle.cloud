export default defineNuxtRouteMiddleware(async (to) => {
  try {
    await callOnce(async () => {
      const { session, loadSession } = useAuth({ init: true, interval: 1, debug: true })
      await loadSession()
      console.log(' *** auth mw (', import.meta.client ? 'client' : import.meta.server ? 'server' : 'neither client nor server', ') => session: ', session.value?.sub, session.value?.expires_in,
      )
    })
  }
  catch (e) {
    console.log(' *** auth.global.session error = ', e.status, e.message)
  }
})
