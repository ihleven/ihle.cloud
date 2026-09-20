interface Person {

  key: string
  geburtstag: string // date
  todestag: string // date
  vater: string
  mutter: string
  // The markdown below the frontmatter, carried by a content.Body on the Go
  // side. Named `body` and not `markdown`: the CMS's generic markdown codec
  // fills a content.Body, and the field this replaced was never populated.
  body: string
}

// A journey, as app/familie/reise.go stores it. See Person.body for why the
// entry's markdown arrives as `body`.
interface Reise {
  ziel: string
  jahr: number
  von: string
  bis: string
  body: string
}
