import { defineEventHandler, getCookie, getHeaders, getMethod, getQuery, readBody } from 'h3'

export default defineEventHandler(event => {
  const { apiBaseUrl } = useRuntimeConfig()
  const url = event.node.req.url as string
  // const target = new URL(url.slice(4), apiBaseUrl)
  const target = new URL(url, apiBaseUrl)

  // const target = new URL(event.node.req.url, apiBaseUrl)

  return proxyRequest(event, target.toString(), {
    headers: {
      host: target.host, // if you need to bypass host security
    },
  })
})
// const { apiBaseUrl } = useRuntimeConfig()
// // const baseURL = config.apiBaseUrl
// export default defineEventHandler(async event => {
//   const method = getMethod(event)
//   const params = getQuery(event)

//   const headers = getHeaders(event)
//   // const authorization = headers.Authorization || getCookie(event, 'auth._token.local')

//   let url = event.node.req.url as string
//   if (url.startsWith('/api')) {
//     // PREFIX is exactly at the beginning
//     url = url.slice(4)
//   }

//   proxyRequest

//   const body = method === 'GET' ? undefined : await readBody(event)
//   return await $fetch(url, {
//     headers: {
//       'Content-Type': headers['content-type'] as string,
//       // Authorization: authorization as string,
//     },
//     baseURL: apiBaseUrl,
//     method,
//     params,
//     body,
//   })
// })
