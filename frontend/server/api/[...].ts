export default defineEventHandler(event => {
  const { apiBaseUrl } = useRuntimeConfig()
  const target = new URL(event.context.params._, apiBaseUrl)
  // const target = new URL(event.node.req.url, apiBaseUrl)

  return proxyRequest(event, target.toString(), {
    headers: {
      host: target.host, // if you need to bypass host security
    },
  })
})

// import { defineEventHandler, getCookie, getHeaders, getMethod, getQuery, readBody } from 'h3'
// const config = useRuntimeConfig()
// const baseURL = config.apiBaseUrl
// export default defineEventHandler(async (event) => {
//   const method = getMethod(event)
//   const params = getQuery(event)

//   const headers = getHeaders(event)
//   const authorization = headers.Authorization || getCookie(event, 'auth._token.local')

//   const url = event.node.req.url as string

//   const body = method === 'GET' ? undefined : await readBody(event)

//   return await $fetch(url, {
//     headers: {
//       'Content-Type': headers['content-type'] as string,
//       Authorization: authorization as string,
//     },
//     baseURL,
//     method,
//     params,
//     body,
//   })
// })
