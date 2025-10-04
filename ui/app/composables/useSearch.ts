import type Search from '~/pages/search.vue'

type Search = {
  params: SearchParams
  result: SearchResult
  search: () => void
}

type SearchParams = {
  q?: string
  size?: number
  from?: number
  sort?: string
  fields?: string
  type?: string[]
  version?: string
  facet?: string[]
  sorting?: string
}

export const useSearch = () => {
  const config = useRuntimeConfig()

  const params = useState<SearchParams>('params', () => ({
    q: '',
    size: 10,
    from: 0,
    sort: 'score',
    fields: '*',
    type: [],
    version: undefined,
    facet: ['type', 'version'],
    sorting: '-modified',
    highlight: true,
  }))

  const result = useState<SearchResult>('result', () => ({
    total_hits: 0,
    cost: 0,
    max_score: 0,
    took: 0,
    took_ms: '',
    tookms: '',
    hits: [] as SearchResultHit[] } as SearchResult))

  const facets = ref({
    type: [
      { label: 'Page', count: 0 },
      { label: 'Badge', count: 0 },
    ],
  })

  async function search(p: SearchParams) {
    // console.log('SearchParams', p)
    const data = await $fetch<SearchResult>(config.public.api.base + '/api/v1/search', {
      query: p, // ? p : params.value,
      method: 'GET',
    })

    result.value = data

    const tmp = {}
    for (const term of data.facets.type.terms || []) {
      tmp[term.term] = term.count
    }
    // console.log('tmp', tmp)
    for (const term of facets.value.type) {
      term.count = tmp[term.label]
      // console.log('facet', term.label, tmp[term.label], term.count)
    }
    // console.log('data', data.facets.type)
    // for (const [name, facet] in Object.entries(data.facets)) {
    //   console.log('facet', name, facet)

    // }
  }

  watchEffect(() => {
    console.log(`watch: params changed to ${params.value}`)
    search(params.value)
  })
  // watch(params.value, (newValue, oldValue) => {
  //   console.log(`watch: params changed from ${oldValue} to ${newValue}`)
  //   search(params.value)
  // }, { immediate: true })

  return {
    params,
    result,
    search,
    facets,
  } as unknown as Search
}
