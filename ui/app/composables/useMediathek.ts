// Reaching the shared video library, and making sense of what is on it.
//
// The endpoints mirror the file browser's and are reached the same way; what is
// different is everything below them. The library is a folder somebody has been
// filling for years, not a catalogue somebody designed, so its shape has to be
// read out of it rather than assumed. Each rule here is a reading of the real
// shelf, and each is wrong somewhere — which is why they live in the client,
// where being wrong is cheap to correct.

function base(): string {
  return useRuntimeConfig().public.apiBaseURL as string
}

/** A path as the API wants it in a route: each segment encoded, slashes kept. */
function encodePath(path: string): string {
  return path.split('/').filter(Boolean).map(encodeURIComponent).join('/')
}

export interface Series {
  /** The directory, which is the path to everything in it. */
  dir: string
  /** What to call it: the directory with its year taken off, tidied. */
  title: string
  /** The year, where the directory names one. */
  year?: string
  /** Playable files, with the copies of one episode gathered together. */
  episodes: Episode[]
  cover: string
}

export interface Episode {
  /** "1.1", "01", … where the filename numbers itself. */
  number?: string
  /** What to call it: the filename, less its number, variant and extension. */
  title: string
  /** The same episode at different resolutions, largest file first. */
  variants: DriveMeta[]
}

export function useMediathek() {
  /** A directory in the library: what it holds, with sizes and types. */
  function meta(path: MaybeRefOrGetter<string>) {
    return useFetch<DriveMeta>(() => `${base()}/mediathek/meta/${encodePath(toValue(path))}/`, {
      credentials: 'include',
    })
  }

  /** The bytes of a file, ranged and seekable. */
  function stream(path: string): string {
    return `${base()}/mediathek/stream/${encodePath(path)}`
  }

  /**
   * The shelf: every series, with its cover and its playable files.
   *
   * One request for the root and one per directory, in parallel. The second
   * round cannot be avoided — a listing gives one level, so a directory's own
   * members are not in it, and both questions the wall asks are about those
   * members. `nmembers` counts the archive masters and the cover too, so it
   * answers neither "is there a cover" nor "how many episodes".
   *
   * Ten directories, so the fan-out is bounded by the shelf rather than by
   * anything a visitor does. If it ever grows past that, the answer is for the
   * server to assemble it, not for this to fetch harder.
   */
  async function shelf(): Promise<Series[]> {
    const root = await $fetch<DriveMeta>(`${base()}/mediathek/meta/`, { credentials: 'include' })

    const dirs = (root.members ?? []).filter(m => m.category === 'directory')
    const listings = await Promise.all(dirs.map(async (d) => {
      const dir = driveName(d)
      try {
        return { dir, listing: await $fetch<DriveMeta>(`${base()}/mediathek/meta/${encodePath(dir)}/`, { credentials: 'include' }) }
      }
      catch {
        // One unreadable directory should cost its own tile, not the wall.
        return { dir, listing: null }
      }
    }))

    return listings.flatMap(({ dir, listing }) => {
      if (!listing || !hasCover(listing)) return []
      const { title, year } = titleOf(dir)

      return [{ dir, title, year, episodes: episodesOf(listing), cover: stream(`${dir}/cover.jpg`) }]
    })
  }

  return { meta, stream, shelf }
}

/**
 * Whether a directory carries a cover, which is what makes it a series.
 *
 * Nobody wrote that convention down — it is what the shelf turns out to look
 * like, and it happens to sort the real entries from the leftovers exactly.
 * Beside the eight covered directories sit two uppercase copies of series that
 * already have one, and three loose files nobody filed. Requiring a cover
 * leaves all of them out without naming any of them, and lets a new series join
 * by being given one.
 */
function hasCover(listing: DriveMeta): boolean {
  return (listing.members ?? []).some(m => driveName(m).toLowerCase() === 'cover.jpg')
}

/**
 * A directory name read as a title and a year.
 *
 * "2001-walking-with-beasts" is the shape most of them take, so the leading
 * year becomes a year and the rest becomes words. "captain-future" has none and
 * keeps its name. Nothing here is authoritative: it is a label, and a shelf
 * that wanted better labels would carry them in a file.
 */
export function titleOf(dir: string): { title: string, year?: string } {
  const match = /^(\d{4})[-_ ](.+)$/.exec(dir)
  const [year, rest] = match ? [match[1], match[2]!] : [undefined, dir]

  return { year, title: words(rest) }
}

function words(slug: string): string {
  return slug
    .replace(/[-_]+/g, ' ')
    .replace(/\s+/g, ' ')
    .trim()
    .replace(/(^|\s)\S/g, c => c.toUpperCase())
}

/**
 * The files in a directory a browser can actually play.
 *
 * Matroska is excluded, and that is not a detail: these directories hold the
 * archive master beside the viewable copy, and in one of them the masters
 * outnumber the copies five to two. Listing them would fill a page with rows
 * that open and then play nothing.
 */
export function playable(dir: DriveMeta): DriveMeta[] {
  return (dir.members ?? []).filter(m => /\.mp4$/i.test(driveName(m)))
}

/** The resolution a file was encoded at, where the name carries one. */
const variantSuffix = /\.Creator\d+p\d+$/i

/**
 * A leading episode number: "1.1", "01", "3".
 *
 * The separators after it are taken greedily, and the dash is both the ASCII
 * one and the typographic one — "1. – Vorläufer oder Blutsbrüder.mp4" puts an
 * en dash between the number and the title, and leaving it behind starts the
 * title with a stray dash.
 */
const leadingNumber = /^(\d+(?:\.\d+)?)[.\-–—_ ]+/

/**
 * The episodes in a directory, with the copies of one gathered together.
 *
 * A count of files is not a count of episodes. One series holds eleven files and
 * eight episodes, because its first is kept at four resolutions — and spells
 * itself "1.1_" once and "1.1-" three times, so the copies cannot be matched by
 * their names alone. The number in front is what identifies an episode, and it
 * survives that inconsistency.
 *
 * Where a file numbers itself, that number groups it. Where none does — one
 * directory is raw rip names like "B1_t00" — the filename stands in, so each is
 * its own episode and nothing is silently merged.
 */
export function episodesOf(dir: DriveMeta): Episode[] {
  const groups = new Map<string, Episode>()

  for (const file of playable(dir)) {
    const name = driveName(file).replace(/\.mp4$/i, '').replace(variantSuffix, '')
    const numbered = leadingNumber.exec(name)
    const number = numbered?.[1]
    const key = number ?? name

    const episode = groups.get(key)
    if (episode) {
      episode.variants.push(file)
      continue
    }
    groups.set(key, { number, title: episodeTitle(name, numbered?.[0]), variants: [file] })
  }

  for (const episode of groups.values()) {
    episode.variants.sort(bestFirst)
  }

  return [...groups.values()]
}

/**
 * A filename read as an episode title.
 *
 * Only separators that are standing in for spaces are replaced: "1.1-Eine-neue-
 * Zeit" wants them, "01. Die Rückverwandlung" already has spaces and its
 * hyphens, if any, are part of the words. Capitalisation is left alone, because
 * these are German titles and title-casing them would be wrong.
 */
function episodeTitle(name: string, numberPrefix?: string): string {
  const rest = numberPrefix ? name.slice(numberPrefix.length) : name

  return (rest.includes(' ') ? rest : rest.replace(/[-_]+/g, ' ')).trim() || name
}

/**
 * Which copy of an episode to offer first: the biggest file.
 *
 * Not what their names suggest. The four copies of "1.1" call themselves 720p,
 * 1080p, 1440p and 2160p, and all four are 702x406 — measured with ffprobe
 * through this very endpoint. What actually differs is the bitrate, from 3.5
 * down to 2.0 Mbit/s, and it runs *opposite* to the resolution in the name: the
 * file calling itself 2160p is the worst of the four.
 *
 * So those names are decoration from whatever upscaler wrote them, and reading
 * a quality out of them would confidently hand out the poorest copy. Same
 * picture and same length in every copy means the size is the quality — and it
 * is the only thing here that was measured rather than claimed.
 */
function bestFirst(a: DriveMeta, b: DriveMeta): number {
  return (b.size ?? 0) - (a.size ?? 0)
}

const unused_medien = [
  { dir: 'captain-future' },
  // {
  //   name: 'Die Erben der Saurier',
  //   dir: 'Die-Erben-der-Saurier',
  //   link: '/mediathek/Die-Erben-der-Saurier/1.1-Eine-neue-Zeit',
  //   img: '/public/mediathek/Die-Erben-der-Saurier/cover.jpg',
  //   class: "bg-[url('http://localhost:8000/hi/media/public/mediathek/Die-Erben-der-Saurier/cover.jpg')]",
  // },
  { dir: '2001-walking-with-beasts', titel: 'Die Erben der Saurier', orig: 'Walking with Beasts', jahr: 2001 },
  {
    dir: '2003-walking-with-cavemen',
    titel: 'Im Reich der Urmenschen',
    orig: 'Walking with Cavemen',
    jahr: 2003,
  },
  {
    dir: '2003-sea-monsters',
  },
  {
    dir: '2003-monsters-we-met', titel: 'Menschen gegen Monster', orig: 'Monsters We Met', jahr: 2003,
  },
  {
    dir: '2005-walking-with-monsters', titel: 'Die Ahnen der Saurier', orig: 'Walking with Monsters', jahr: 2005,
  },
  {
    dir: '2011-planet-of-the-apemen',
    titel: 'Kampf der Menschenaffen',
    orig: 'Planet of the Apemen: Battle for Earth',
    jahr: 2011,
  },
  { dir: '2013-ice-age-giants', orig: 'Ice Age Giants', jahr: 2013 },
]
