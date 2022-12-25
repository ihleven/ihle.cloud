export default defineEventHandler((event) => {
 
  const { apiBaseUrl } = useRuntimeConfig()
  const target = new URL(event.context.params._, apiBaseUrl)
  // const target = new URL(event.node.req.url, apiBaseUrl)

  return proxyRequest(event, target.toString(), {
    headers: {
      host: target.host // if you need to bypass host security
    }
  })
})