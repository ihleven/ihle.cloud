export default defineNuxtRouteMiddleware(async (to, from) => {
  // console.log('A', to.path)
  // if (to.path.startsWith('/api')) {
  //   console.log('A')

  //   return
  // }

  // console.log(
  //   'AUTH MIDDLEWARE',
  //   from.path.length,
  //   ' => ',
  //   to.path, // .slice(0, 20),
  //   import.meta.client ? 'CLIENT' : import.meta.server ? 'SERVER' : 'NEITHER CLIENT NOR SERVER',
  // )

  if (!to.path.startsWith('/hidrive')) {
    return
  }

  if (import.meta.server) {
    // const myauth = useAuth()
    // if (!myauth.value) return
    // else
    //   return navigateTo('/login', {
    //     external: true,
    //   })
    return
  }

  let session
  try {
    const headers = useRequestHeaders()
    session = await $fetch('http://localhost:8000/auth/session', { headers: { ...headers }, credentials: 'include' })
    const myauth = useAuth()

    if (session) {
      myauth.auth.value = session
    }

    console.log('session = ', myauth.auth)
  }
  catch (e) {
    console.log(' *** session error:', e)
    return navigateTo('/login?redirect=' + to.path)
  }

  console.log(` *** server side auth middleware ${session?.id} => ${to.path.length} ***`)
})
