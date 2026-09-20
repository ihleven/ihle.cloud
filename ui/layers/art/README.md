# art

The art archive's front end, brought over from the old `art` repo
(`~/src/art`, `github.com/ihleven/art`) when its Go packages were migrated into
this app, and wired up here.

It serves `/werke`: a facet search over the artworks by Gattung, Technik and
Träger.

## How it is wired

- The layer is listed in `extends` in `ui/nuxt.config.ts`. (Nuxt also
  auto-registers anything under `layers/`, so the entry is for legibility rather
  than effect — a layer here cannot be switched off by leaving it out.)
- The page fetches `/art/search` against `apiBaseURL`, i.e.
  `/api/v1/art/search`, served by `app/art/api.go` and mounted in `main.go`
  beside the familie routes. The old `/art-api/` mux is gone; `/art-api` was
  never proxied in dev, so it could not have worked from the SPA.
- Filtering and the counts come from the same place: each filter sends a query
  parameter declared in `ExtParams` (`app/art/bleve.go`), and each count is read
  from the facet of the same name requested in `api.go`.

## It needs `SEARCH_LEVEL=fulltext` or higher

Everything the page filters and counts on lives in the `artwork` sub-document,
written by `AugmentSearchDoc` — which the indexer calls only at `fulltext` and
above. At `basic`, which is the flag's default in `main.go`, every facet comes
back with no terms and every filter matches nothing.

So `/api/v1/art/search` refuses with 501 below that level rather than answering
emptily, and the page shows what it said. Measured: at `basic` the forms facet
has 0 terms, at `fulltext` it has 4 over 308 works.

## Values are the archive's own

The indexed terms are the old database's spellings, and they are compared as
keywords — so the options say `M`, `Z`, `P` for the Gattung and `Öl`,
`Leinwand` and so on for Technik and Träger. Lower-casing them finds nothing:
measured against the live index, `form=M` returns 275 works and `form=m`
returns none.

## Facets are lists, not maps

`facets.<name>.terms` is bleve's shape — an array of `{term, count}` — so
`countOf()` in the page looks a value up rather than indexing it. The page
originally read `data.aggs.<name>[value]`, which matched a `Result` type that
still exists in `app/art/bleve.go` but is not what the mounted handler returns.

## Searching for deleted works

The old database marks some works `DEL`. They are imported only with
`./ihlvn import --bilder --deleted`, arrive as INACTIVE entries keeping their
`DEL` status, and are indexed — so `?status=DEL` finds them. Nothing in the
repository carries that status yet, so that query currently returns nothing.
