export default defineNuxtRouteMiddleware(async () => {
  await callOnce(async () => {
    const { session, loadSession } = useAuth({ init: true, interval: 1, debug: false })
    await loadSession()
    const where = import.meta.client ? 'client' : import.meta.server ? 'server' : 'neither client nor server'
    console.log(` *** auth.global.ts (${where}) session =>`, JSON.parse(JSON.stringify(session.value ?? null)))
  })
})
