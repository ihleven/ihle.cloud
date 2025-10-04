interface SearchResult {
  total_hits: number
  cost: number
  max_score: number
  took: number
  took_ms: string
  hits: SearchResultHit[]
  facets?: Record<string, SearchResultFacet>
}

interface SearchResultHit {

  id: string
  score: number
  fields: Record<string, string | string[] | number | number[] | boolean>
  fragments: Record<string, string[]>
}

interface SearchResultFacet {
  field: string
  missing: number
  total: number
  other: number
  terms: { count: number, term: string }[]
}
