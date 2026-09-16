// Reaching the music shelf — two ways, on purpose.
//
// The shelf is one fixed tree on the family storage, the same for everybody
// entitled to it, so the same arrangement as the Mediathek and the archive. The
// difference is that the storage offers two ways of reading it, and neither is
// simply better:
//
//   - Listing reads one level per request. A shelf of folders therefore costs a
//     request per folder, but each is fast — twenty-five milliseconds warm —
//     and they all go at once.
//   - Searching reads the whole tree in one request whatever its shape, but
//     takes seconds rather than milliseconds, and the cost does not follow the
//     size of the answer: the whole of this shelf took as long as the whole of
//     a drive holding a hundred times more.
//
// Measured against six albums, listing wins by two orders of magnitude. Which
// way it goes when there are five hundred is not something to guess at, so both
// are built and the page switches between them. They are written to produce the
// same albums from the same shelf, because a comparison of two things that
// disagree about the answer is not a comparison of anything.
//
// What the names and folders mean is read in utils/musik, away from here, so it
// can be tested against real filenames without a browser.

/** Which reading of the shelf a page is using. */
export type Strategy = 'dir' | 'search'

/** A shelf, and what it cost to read it. */
export interface Shelf {
  albums: Album[]
  /** How many requests went to the storage, which is the thing being compared. */
  requests: number
  /** How long they took, in milliseconds, wall clock. */
  took: number
}

function base(): string {
  return useRuntimeConfig().public.apiBaseURL as string
}

/** A path as the API wants it in a route: each segment encoded, slashes kept. */
function encodePath(path: string): string {
  return path.split('/').filter(Boolean).map(encodeURIComponent).join('/')
}

export function useMusik() {
  /** A directory on the shelf: what it holds, one level down. */
  function meta(path: string) {
    return $fetch<DriveMeta>(`${base()}/musik/meta/${encodePath(path)}/`, { credentials: 'include' })
  }

  /**
   * Everything below a path, at any depth, in one request.
   *
   * What comes back has to be selected: `category: 'dir'` takes every folder
   * and needs no pattern, while files are reached by `pattern`, a plain
   * substring of the name. That asymmetry is the storage's, not ours — see the
   * Go client's Search.
   */
  function search(path: string, select: { pattern?: string, category?: string }) {
    return $fetch<{ result: DriveMeta[] }>(`${base()}/musik/search/${encodePath(path)}`, {
      credentials: 'include',
      query: select,
    })
  }

  /**
   * What the files of one album say about themselves.
   *
   * An album at a time rather than a track at a time, because it is a request
   * per file on the server and would be a request per track from here as well.
   * Only asked for when an album is opened: the wall shows no artists, so
   * drawing it needs none of this.
   */
  function tags(dir: string) {
    return $fetch<{ tracks: MusikTags[] }>(`${base()}/musik/tags/${encodePath(dir)}`, {
      credentials: 'include',
    })
  }

  /** A cover, at the width a tile shows it. */
  function cover(path: string, width = 400): string {
    return `${base()}/musik/thumb/${encodePath(path)}?width=${width}`
  }

  /**
   * A track's bytes.
   *
   * The streaming route rather than the archive's signed-URL one: a player asks
   * for one range at a time, which is the case that route serves, and it keeps
   * the storage's own address out of the browser.
   */
  function track(path: string): string {
    return `${base()}/musik/stream/${encodePath(path)}`
  }

  /**
   * The shelf, read a level at a time.
   *
   * Walked rather than listed once, so that a folder of artists each holding
   * their albums works without changing anything here. Today the shelf is flat
   * and the walk stops after one level; the cost of being ready for the other
   * shape is one request that finds no folders.
   *
   * Every level goes out at once. The depth is what costs, not the width: a
   * hundred albums under one folder is one round trip, the same as six.
   */
  async function byDir(): Promise<Shelf> {
    const started = performance.now()
    let requests = 0

    async function walk(dir: string): Promise<Album[]> {
      requests++
      const listing = await meta(dir)
      const members = listing.members ?? []
      const below = members.filter(m => m.category === 'directory')

      // A folder with folders in it is a shelf, not an album — the same rule
      // leafDirs applies to the other reading, so the two agree.
      if (below.length) {
        const deeper = await Promise.all(
          below.map(async (m) => {
            const under = dir ? `${dir}/${driveName(m)}` : driveName(m)
            try {
              return await walk(under)
            }
            catch (err) {
              // One unreadable folder should cost what is in it, not the shelf.
              console.warn('skipping a folder of the shelf:', under, err)

              return []
            }
          }),
        )

        return deeper.flat()
      }

      return dir ? [albumOf(dir, members.map(driveName), members)] : []
    }

    const albums = await walk('')

    return { albums: sortAlbums(albums), requests, took: performance.now() - started }
  }

  /**
   * The shelf, read in one sweep.
   *
   * Three requests whatever the shelf's shape, because the endpoint selects
   * rather than enumerates: one for the folders, one for the tracks, one for
   * the covers. They are three and not one for a reason worth knowing — only
   * directories can be asked for as a category, so everything else has to be
   * named by a substring of its filename.
   *
   * Which is also this reading's limitation: a substring is one string. Asking
   * for "mp3" finds mp3s, and a shelf holding FLACs as well would need a fourth
   * request for them. The listing above has no such trouble, since it sees
   * whatever is in the folder.
   */
  async function bySearch(): Promise<Shelf> {
    const started = performance.now()

    const [dirs, audio, pictures] = await Promise.all([
      search('', { category: 'dir' }),
      search('', { pattern: 'mp3' }),
      search('', { pattern: 'cover' }),
    ])

    // The shelf's own folder comes back among the hits, and comes back absolute
    // — it is not below itself, so it has no relative name. That is what marks
    // it: anything still starting with a slash is not inside the shelf.
    const folders = dirs.result.map(d => d.path ?? '').filter(p => p && !p.startsWith('/'))
    const files = [...audio.result, ...pictures.result]
      .map(f => f.path ?? '')
      .filter(p => p && !p.startsWith('/'))

    const held = new Map<string, string[]>()
    for (const path of files) {
      const dir = parentDir(path)
      const names = held.get(dir)
      if (names) names.push(path)
      else held.set(dir, [path])
    }

    const albums = leafDirs(folders).map((dir) => {
      const paths = held.get(dir) ?? []

      return albumOf(dir, paths.map(lastSegment))
    })

    return { albums: sortAlbums(albums), requests: 3, took: performance.now() - started }
  }

  /**
   * A folder and the names in it, read as an album.
   *
   * Both readings end here, which is what makes them comparable: they differ in
   * how the names were obtained and in nothing else. Sizes are passed where the
   * caller has them — the listing does, the search does too, but only for what
   * its patterns matched.
   */
  function albumOf(dir: string, names: string[], members?: DriveMeta[]): Album {
    const sized = new Map<string, number | undefined>(
      (members ?? []).map(m => [driveName(m), m.size]),
    )

    const tracks = names.filter(isAudio).map((name) => {
      const { number, title } = trackTitle(name)

      return { name, path: `${dir}/${name}`, title, number, size: sized.get(name) }
    })

    const picture = names.find(isCover) ?? names.find(isPicture)

    return {
      path: dir,
      title: albumTitle(dir),
      year: albumYear(dir),
      cover: picture ? `${dir}/${picture}` : undefined,
      tracks: sortTracks(tracks),
    }
  }

  /** The shelf, read whichever way was asked for. */
  function shelf(strategy: Strategy): Promise<Shelf> {
    return strategy === 'search' ? bySearch() : byDir()
  }

  return { shelf, byDir, bySearch, meta, search, tags, cover, track }
}
