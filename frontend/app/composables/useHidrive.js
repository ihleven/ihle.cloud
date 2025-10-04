export default function () {
  // const route = useRoute()

  // const path = computed(() => route.params.slug.join('/'))

  // console.log('slug', `${path.value}`, route.query)

  // const { data, refresh } = await useFetch(`/api/${path.value}`, {
  //   server: false,
  //   key: path.value,
  //   // query: route.query,
  // })
  // const files = useState('files', () => data)

  // files.value = data

  // watch(
  //   () => path,
  //   () => refresh()
  // )
  const config = useRuntimeConfig()

  function metaURL() {
    const { params } = useRoute()
    const path = typeof params.slug === 'string' ? params.slug : params.slug.join('/')
    console.log('hidirive path =>', path, params, config.public.hidrive.api)

    return `${config.public.hidrive.api}/meta/${path}`
  }

  function media(slugs) {
    return 'http://localhost:8000/' + pathJoin(['hi/media/', slugs], '/')
  }
  function pathJoin(parts, sep) {
    const separator = sep || '/'
    const replace = new RegExp(separator + '{1,}', 'g')
    return parts.join(separator).replace(replace, separator)
  }
  return {
    media,
    pathJoin,
    metaURL,
    // path,
    // files,
    // refresh,
  }
}
function pathJoin(parts, sep) {
  const separator = sep || '/'
  const replace = new RegExp(separator + '{1,}', 'g')
  return parts.join(separator).replace(replace, separator)
}
