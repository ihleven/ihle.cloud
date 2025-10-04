export default defineEventHandler(async (event) => {
  const config = useRuntimeConfig()

  // const { data } = await useSession(event, config.h3session)

  // let accesstoken = data.token.access_token

  // const now = Math.floor(Date.now() / 1000)

  // const expired = data.token.expires < now // new Date().getTime() > (new Date(data.account.exp).getTime() - 120 * 1000)
  // console.log('token:', data.account.exp, expired, data.token.expires, now)

  // try {
  //   if (expired) {
  //     const headers = getHeaders(event)

  //     const resp = await $fetch('/api/auth/refresh', {
  //       headers: { Cookie: headers.cookie as string },
  //     })

  //     const session = await useSession(event, config.h3session)

  //     await session.update({ token: resp })
  //     accesstoken = resp.access_token
  //   }
  // } catch (err) {
  //   console.log(err)
  //   await useServerSideLogout()
  //   // return sendRedirect('/login')
  //   return {
  //     redirect: {
  //       statusCode: 302,
  //       destination: '/login',
  //     },
  //   }
  // }

  let body = null
  const headers = getHeaders(event)
  if (event.method === 'POST' || event.method === 'PUT') {
    if (headers['content-type']?.startsWith('multipart/form-data')) {
      body = new FormData()
      const mulitpart = await readMultipartFormData(event)
      console.log('form data', mulitpart)
      mulitpart?.forEach((part) => {
        console.log('part', part)
        if (part.name === 'file') {
          body.set(part.name, new Blob([part.data]))
        }
        else {
          body.set(part.name, part.data.toString())
        }
      })
    }
    else {
      body = await readBody(event)
    }
  }
  // try {
  //   body = await readRawBody(event)
  // } catch (e) {}

  const route = getRouterParam(event, 'route')
  const jwt = getCookie(event, 'jwt')
  return $fetch(config.contentApiUrl + '/content-api/v1/' + route, {
    method: event.method,
    headers: { Authorization: `Bearer ${jwt}` },
    query: getQuery(event),
    body,
    // ignoreResponseError: true,
    onRequest({ request }) {
      console.log('[proxy request]', route, '=>', request)
    },
  }).catch((error) => {
    throw createError({ status: error.status, statusText: error.data })
  })

  // .catch((error) => ({ status: error.status, msg: error.data }))
})
