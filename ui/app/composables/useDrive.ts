// Reaching the file browser's endpoints.
//
// One place holds the base and the shapes of the three URLs, so a page never
// builds them by hand. The legacy browser hardcoded the host into its gallery
// twice, which is why its pictures only ever loaded on one laptop.

function base(): string {
  return useRuntimeConfig().public.apiBaseURL as string
}

/**
 * An entry's name, as text.
 *
 * The provider returns names percent-encoded — pkg/hi calls the field
 * NameURLEncoded and means it — so a name has to be decoded before it is shown
 * or joined into a path. Encoding it again without decoding first turns a space
 * into %2520, which the old browser did and which resolves to nothing.
 */
export function driveName(entry: Pick<DriveMeta, 'name'>): string {
  try {
    return decodeURIComponent(entry.name)
  }
  catch {
    // A name with a stray % is not encoded at all; show it as it came.
    return entry.name
  }
}

/** A path as the API wants it in a route: each segment encoded, slashes kept. */
function encodePath(path: string): string {
  return path.split('/').filter(Boolean).map(encodeURIComponent).join('/')
}

export function useDrive() {
  /** The entry at a path: a file, or a directory and its members. */
  function meta(path: MaybeRefOrGetter<string>) {
    return useFetch<DriveMeta>(() => `${base()}/drive/meta/${encodePath(toValue(path))}`, {
      credentials: 'include',
    })
  }

  /** The bytes of a file. */
  function media(path: string): string {
    return `${base()}/drive/media/${encodePath(path)}`
  }

  /**
   * A preview of an image. The width is a request, not a promise: the server
   * rounds it to a fixed set, so asking for an arbitrary one only costs a
   * cache miss.
   */
  function thumb(path: string, width = 200): string {
    return `${base()}/drive/thumb?path=${encodeURIComponent(path)}&width=${width}`
  }

  return { meta, media, thumb }
}

/**
 * Bytes, as the storage provider counts them: powers of 1000, not 1024.
 *
 * Carried over from the old browser rather than rewritten, so the same file
 * reports the same size in both.
 */
export function driveSize(num?: number): string {
  if (!num) return ''

  const units = ['B', 'kB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB']
  const neg = num < 0
  if (neg) num = -num
  if (num < 1) return `${neg ? '-' : ''}${num} B`

  const exponent = Math.min(Math.floor(Math.log(num) / Math.log(1000)), units.length - 1)
  const value = Number((num / Math.pow(1000, exponent)).toPrecision(3))

  return `${neg ? '-' : ''}${value} ${units[exponent]}`
}

/** A file's text, for the source view. */
export function driveText(url: string): Promise<string> {
  return $fetch<string>(url, { credentials: 'include', responseType: 'text' })
}

/**
 * When an entry was last touched, or nothing.
 *
 * Takes whichever of the two the provider sent. A directory listing carries
 * `ctime` for its members and not `mtime` — the field list /dir asks for says
 * so — while the entry fetched on its own has both. The legacy browser read
 * only one of them and printed "Invalid Date" down the whole column.
 */
export function driveDate(entry?: Pick<DriveMeta, 'mtime' | 'ctime'>): string {
  const seconds = entry?.mtime ?? entry?.ctime
  if (!seconds) return ''
  return new Intl.DateTimeFormat('de-DE', {
    year: 'numeric', month: 'numeric', day: 'numeric', hour: '2-digit', minute: '2-digit',
  }).format(new Date(seconds * 1000))
}
