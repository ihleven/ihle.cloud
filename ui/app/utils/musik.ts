// Reading a music shelf out of its folder and file names.
//
// Nothing on the shelf is labelled. A folder is called "1987 - Appetite For
// Destruction (lameV3A)" and that string carries three separate facts — a year,
// a title, and which encoder ripped it — none of which is marked as such. The
// artist is carried nowhere at all, which is why it is not read here: it comes
// from the tags inside the files, or from a folder level the shelf does not yet
// have.
//
// Every rule below is a reading of what the shelf actually looks like today,
// and each is wrong somewhere, which is why they live apart from the fetching
// where they can be tested against real names.

export interface Track {
  /** The file, as it is stored. */
  name: string
  /** Where it is, relative to the shelf — what the routes address. */
  path: string
  /** What to call it: the filename, less its number and extension. */
  title: string
  /** Its place on the album, where the filename gives one. */
  number?: number
  size?: number
}

export interface Album {
  /** The folder, relative to the shelf. Its identity, since nothing else is. */
  path: string
  /** What to call it: the folder name, less the year and the encoder. */
  title: string
  /** The year on the folder, where it carries one. */
  year?: string
  /** The cover's path, where the folder holds a picture. */
  cover?: string
  tracks: Track[]
}

const audio = /\.(mp3|m4a|flac|ogg|opus|wav|aac|wma)$/i

export function isAudio(name: string): boolean {
  return audio.test(name)
}

/**
 * Whether a file is the album's picture.
 *
 * "cover.jpeg" is what this shelf uses and what the old page addressed
 * directly, but the name is a convention rather than a rule, so anything called
 * cover or folder counts, and failing that the first picture in the folder is
 * taken. An album with no picture at all is still an album.
 */
export function isCover(name: string): boolean {
  return /^(cover|folder|front|album)\.(jpe?g|png|webp|gif)$/i.test(name)
}

export function isPicture(name: string): boolean {
  return /\.(jpe?g|png|webp|gif)$/i.test(name)
}

/**
 * What to call an album, from the folder holding it.
 *
 * Two things are stripped, and only two. The leading year is shown separately,
 * so repeating it in the title would say it twice. The trailing parenthesis is
 * the encoder the rip was made with — "(lameV3A)" — which is a fact about the
 * file and not about the album; it is only removed when it reads like one,
 * because a title may legitimately end in brackets and "(Live At Budokan)" must
 * survive.
 */
export function albumTitle(dir: string): string {
  return lastSegment(dir)
    .replace(/^\s*(?:19|20)\d{2}\s*[-–—.]\s*/, '')
    .replace(/\s*\((?:lame|flac|mp3|ape|wav|aac|ogg|v\d)[^)]*\)\s*$/i, '')
    .trim() || lastSegment(dir)
}

/** The year on a folder, where the name begins with one. */
export function albumYear(dir: string): string | undefined {
  return /^\s*((?:19|20)\d{2})\b/.exec(lastSegment(dir))?.[1]
}

/**
 * A filename read as a track.
 *
 * "01. Hey Stoopid.mp3" is a position and a title run together. The separator
 * varies — a dot, a dash, or nothing but a space — so what is taken is the
 * leading run of digits and whatever punctuation follows it, and a name that
 * starts with a digit for another reason ("1984.mp3") keeps its title, because
 * removing the number would leave nothing behind.
 */
export function trackTitle(name: string): { number?: number, title: string } {
  const stem = name.replace(/\.[^.]+$/, '')
  const match = /^\s*(\d{1,3})\s*(?:[.\-–—_)]\s*|\s)\s*(.+)$/.exec(stem)
  if (!match?.[2]?.trim()) {
    return { title: stem }
  }

  return { number: Number(match[1]), title: match[2].trim() }
}

/** Tracks in the order the album plays, which is not the order a listing gives. */
export function sortTracks(tracks: Track[]): Track[] {
  return [...tracks].sort((a, b) =>
    (a.number ?? Number.MAX_SAFE_INTEGER) - (b.number ?? Number.MAX_SAFE_INTEGER)
    || a.name.localeCompare(b.name))
}

/** Albums in the order the shelf reads: newest year first, then by title. */
export function sortAlbums(albums: Album[]): Album[] {
  return [...albums].sort((a, b) =>
    (b.year ?? '').localeCompare(a.year ?? '') || a.title.localeCompare(b.title))
}

/**
 * Which of a set of folders hold albums rather than other folders.
 *
 * An album is a folder with nothing but files in it. Defining it that way
 * rather than as "a folder holding audio" is what lets an incomplete album —
 * a cover and no tracks yet, which several on this shelf are — still appear,
 * and what lets both ways of reading the shelf agree: one of them sees the
 * folders before it sees anything inside them.
 */
export function leafDirs(dirs: string[]): string[] {
  return dirs.filter(dir => !dirs.some(other => other !== dir && other.startsWith(dir + '/')))
}

/** Whether an album answers to what somebody typed. Accent- and case-blind. */
export function albumMatches(album: Album, query: string): boolean {
  const wanted = fold(query)
  if (!wanted) return true

  return fold(album.title).includes(wanted)
    || fold(album.year ?? '').includes(wanted)
    || album.tracks.some(track => fold(track.title).includes(wanted))
}

/** The last part of a path: the folder itself, without what it sits in. */
export function lastSegment(path: string): string {
  const parts = path.split('/').filter(Boolean)

  return parts[parts.length - 1] ?? ''
}

/** The folder a path sits in, or '' for one at the top of the shelf. */
export function parentDir(path: string): string {
  const cut = path.lastIndexOf('/')

  return cut < 0 ? '' : path.slice(0, cut)
}

/**
 * A string reduced to what a typist and a filename have in common: no case, no
 * accents. Somebody looking for Motörhead types Motorhead.
 */
function fold(s: string): string {
  return s.normalize('NFD').replace(/[\u0300-\u036f]/g, '').toLowerCase().trim()
}

/** What the window already says about an album, above its track list. */
export interface AlbumHeader {
  artist?: string
  album?: string
  year?: string
  genre?: string
}

/**
 * What one track's tags say that the album's own heading does not.
 *
 * Printing every field on every row said the same four things ten times over —
 * the heading has just said them, and a line that repeats it is not metadata, it
 * is wallpaper. What is worth a row of its own is disagreement: the guest on one
 * song, the compilation whose every track is somebody else, the track filed
 * under a year the rest of the album is not.
 *
 * So a uniform album shows nothing here, which is the correct amount to say
 * about an album that has nothing further to say.
 */
export function trackMeta(tag: {
  artist?: string
  album?: string
  year?: number
  genre?: string
}, header: AlbumHeader): string {
  const differs = (value: string | undefined, shown: string | undefined) =>
    !!value && value !== (shown ?? '')

  const parts: string[] = []
  if (differs(tag.artist, header.artist)) parts.push(tag.artist!)
  if (differs(tag.album, header.album)) parts.push(tag.album!)
  if (differs(tag.year ? String(tag.year) : '', header.year)) parts.push(String(tag.year))
  if (differs(tag.genre, header.genre)) parts.push(tag.genre!)

  return parts.join(' · ')
}

/**
 * The one value a set of tracks agrees on, or nothing where they disagree.
 *
 * Disagreement is why this returns nothing rather than the commonest value:
 * where the tracks do not agree there is no album-level answer, and taking the
 * first one would suppress exactly the rows that ought to show.
 */
export function agreed(values: (string | undefined)[]): string {
  const found = new Set(values.filter(Boolean) as string[])

  return found.size === 1 ? [...found][0]! : ''
}
