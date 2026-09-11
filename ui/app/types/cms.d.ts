// The Entry envelope itself comes from the CMS entry layer, which owns the
// editing shell; these name the content shapes this app stores in it.
type PersonEntry = Entry & { content: Person }
type ArtworkEntry = Entry & { content: Artwork }
type Super8Entry = Entry & { content: Super8 }
