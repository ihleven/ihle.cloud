// Reaching the DJ archive.
//
// The same arrangement as the music shelf and the magazines: one fixed tree on
// the family storage, the same for everybody entitled to it, so the answers can
// be cached by address. What differs is that it is arranged — series, then CDs,
// then the files — so this walks two levels and no further, and never searches.
//
// Cost is why it stops there. Eight series of ten CDs is nine requests to draw
// the whole overview, because a sleeve is an address rather than something that
// has to be looked up; listing every CD to draw it would be eighty-nine.
//
// What the folder names mean is read in utils/djvet, away from here, so it can
// be tested against the real names without a browser.

function base(): string {
  return useRuntimeConfig().public.apiBaseURL as string
}

/** A path as the API wants it in a route: each segment encoded, slashes kept. */
function encodePath(path: string): string {
  return path.split('/').filter(Boolean).map(encodeURIComponent).join('/')
}

export function useDjvet() {
  /** A directory in the archive: what it holds, one level down. */
  function meta(path: string) {
    return $fetch<DriveMeta>(`${base()}/djvet/meta/${encodePath(path)}/`, { credentials: 'include' })
  }

  /**
   * What the files of one CD say about themselves.
   *
   * A CD at a time, because it is a request per file on the server. Asked for
   * only when a CD is opened: an overview of sleeves needs none of it.
   */
  function tags(dir: string) {
    return $fetch<{ tracks: MusikTags[] }>(`${base()}/djvet/tags/${encodePath(dir)}`, {
      credentials: 'include',
    })
  }

  /** A picture from the archive, at the width it is shown. */
  function image(path: string, width = 400): string {
    return `${base()}/djvet/thumb/${encodePath(path)}?width=${width}`
  }

  /** A track's bytes, through the streaming route a player wants. */
  function track(path: string): string {
    return `${base()}/djvet/stream/${encodePath(path)}`
  }

  /**
   * The archive: every series, and the sleeves of its CDs.
   *
   * The sleeve is built rather than found — every CD folder holds a cover.jpeg,
   * so its address is known without listing the folder, which is what keeps this
   * to one request per series. A CD that turns out not to have one shows its
   * title instead; the page handles that, because it is the only place that
   * finds out.
   */
  async function archive(width = 320): Promise<DjSeries[]> {
    const top = await meta('')

    return Promise.all(
      (top.members ?? [])
        .filter(m => m.category === 'directory')
        .map(async (m) => {
          const slug = driveName(m)

          return {
            slug,
            title: seriesTitle(slug),
            description: seriesDescription(slug),
            cds: await cdsOf(slug, width),
          }
        }),
    )
  }

  /** The CDs of one series. */
  async function cdsOf(series: string, width = 320): Promise<Cd[]> {
    try {
      const listing = await meta(series)

      return (listing.members ?? [])
        .filter(m => m.category === 'directory')
        .map((m) => {
          const slug = driveName(m)
          const dir = `${series}/${slug}`

          return { dir, slug, title: cdTitle(slug), cover: image(`${dir}/cover.jpeg`, width) }
        })
    }
    catch (err) {
      // One unreadable series should cost what is in it, not the archive.
      console.warn('skipping a series of the archive:', series, err)

      return []
    }
  }

  /**
   * One CD: what is on it, and the two pictures beside it.
   *
   * Both requests go out together — the listing says which files there are, the
   * tags say what they are called, and neither waits on the other.
   *
   * The pictures are looked for by name and then by kind. cover.jpeg and
   * tracks.png is what the archive uses, and falling back to whatever pictures
   * are there costs nothing, because this listing has been fetched anyway.
   */
  async function cd(series: string, slug: string, width = 800): Promise<CdDetail> {
    const dir = `${series}/${slug}`
    const [listing, tagged] = await Promise.all([meta(dir), tags(dir)])

    const members = listing.members ?? []
    const said = new Map((tagged.tracks ?? []).map(t => [t.name, t]))

    const tracks: DjTrack[] = members
      .filter(m => isAudio(driveName(m)))
      .map((m) => {
        const name = driveName(m)
        const tag = said.get(name)
        const { title, number } = named(name, tag)

        return { name, path: `${dir}/${name}`, title, number, artist: tag?.artist, length: duration(tag?.seconds), size: m.size }
      })

    const pictures = members.map(driveName).filter(isPicture)
    const sleeve = pictures.find(n => n.toLowerCase().startsWith('cover.')) ?? pictures.find(isCover)
    const written = pictures.find(n => n.toLowerCase().startsWith('tracks.'))
      ?? pictures.find(n => n !== sleeve)

    return {
      series,
      dir,
      slug,
      title: cdTitle(slug),
      cover: sleeve ? image(`${dir}/${sleeve}`, width) : '',
      tracklist: written ? image(`${dir}/${written}`, width) : '',
      tracks: sortTracks(tracks) as DjTrack[],
    }
  }

  return { archive, cdsOf, cd, meta, tags, image, track }
}

const series = [
  { slug: 'pool-party', name: 'Pool Party' },
  { slug: 'lunar-suite', name: 'Lunar Suite' },
  { slug: 'rxc', name: 'RXC' },
  { slug: 'lounge-club', name: 'Lounge Club' },
  { slug: 'rudi-bar', name: 'Rudi Bar' },
  { slug: 'ambient-meets-acid', name: 'Ambient meets Acid' },
]
