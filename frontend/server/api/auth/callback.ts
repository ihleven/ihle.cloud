export default defineEventHandler(async (event) => {
  console.log('asdf', 'callback')
  // const config = useRuntimeConfig()

  // const params = getQuery(event)

  //   const verfierCookie = getCookie(event, 'verifier')

  //   const verifier = decrypt(verfierCookie, config.cookie.key)
  //   if (!verifier) {
  //     return sendRedirect(event, '/welcome?error=callback:missing_verifier')
  //   }

  //   deleteCookie(event, 'verifier')

  //   const response = await $fetch(config.oauth.authorizeURL, {
  //     method: 'POST',
  //     headers: { 'content-type': 'application/x-www-form-urlencoded' },
  //     body: new URLSearchParams({
  //       code: params.code,
  //       verifier,
  //       state: params.state || '',
  //     }).toString(),
  //   }).catch((e) => {
  //     console.log('RESPONSE CATCH', e)
  //     return sendRedirect(event, '/logout?message=' + e.message)
  //   })

  //   response.expires = Math.floor(Date.now() / 1000) + 250

  const options = {
    // maxAge: 30,
    // expires: new Date(response.expires * 1000),

    sameSite: 'lax', // 'strict' prevents direct auth in case of new single sign login
    httpOnly: true,
    secure: process.env.NODE_ENV !== 'development',
  }

  setCookie(event, 'jwt', 'test', options)
  //   setCookie(event, 'cmsauth', encrypt(JSON.stringify(response), config.cookie.key), config.cookie.options)

  //   const target = params.state ? params.state.replace(/^(\/)/, '') : ''
  //   return sendRedirect(event, '/' + target)

  return { a: 123 }
})
