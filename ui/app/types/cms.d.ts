interface Entry {
  path: string
  repo: string

  id: string
  locale: string
  version: string
  space: string
  collection: string

  name: string
  notes: string
  slug: string
  tags: string[]
  full_slug: string
  status: string

  created: string
  modified: string
  published: string
  owner: string
  group: string
  permissions: string

  mime: string
  type: string
  content: object
}

type PersonEntry = Entry & { content: Person }
type ArtworkEntry = Entry & { content: Artwork }
type Super8Entry = Entry & { content: Super8 }
