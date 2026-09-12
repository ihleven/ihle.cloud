// A digitised reel. Scenes segment it; annotations mark what is visible at a
// moment, optionally at a spot on the frame.

interface Scene {
  title: string
  start: number
  end?: number
  description?: string
}

interface Annotation {
  kind: 'person' | 'location' | 'object' | 'note'
  label: string
  start: number
  end?: number
  // Fractions of the frame's width and height, not pixels.
  at?: { x: number, y: number }
  author?: string
}

interface Film {
  format: string
  key: string
  poster?: string
  filmed_by?: string
  filmed_from?: string
  filmed_to?: string
  location?: string
  duration?: number
  description?: string
  tags?: string[]
  scenes?: Scene[]
  annotations?: Annotation[]
}

type FilmEntry = Entry & { content: Film }
