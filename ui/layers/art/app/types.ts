// What /api/v1/art/search answers with, as far as this page reads it.
//
// The facets are bleve's own shape — a list of terms with counts, not a map
// keyed by value — so the page looks a value up rather than indexing it.
export interface ArtFacet {
  field: string
  total: number
  terms?: Array<{ term: string, count: number }>
}

export interface ArtSearchResult {
  total_hits: number
  hits: Array<{ id: string, fields: Record<string, unknown> }>
  facets?: Record<string, ArtFacet>
  status?: unknown
}
