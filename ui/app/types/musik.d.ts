// What a music file says about itself, as the tags route answers it.
//
// The shelf albums none of this in its layout: a folder carries a year and a
// title, a filename carries a position and a title, and the artist appears
// nowhere at all. These come from inside the files.
interface MusikTags {
  /** The file, so these can be lined up with a listing rather than by order. */
  name: string
  /**
   * Whether the file had tags at all.
   *
   * Without it an empty title cannot be told from a file that failed to read,
   * and the two want different things on screen.
   */
  read: boolean
  title?: string
  artist?: string
  album_artist?: string
  album?: string
  genre?: string
  year?: number
  track?: number
  tracks?: number
  disc?: number
  /** What carried them — ID3v2.3, ID3v1, VORBIS. */
  format?: string
  /**
   * How long the track runs, where the file says so exactly.
   *
   * Absent where it does not, and the page then falls back to what it can work
   * out from the file's size — which is a guess, and was 12% short on one of
   * these albums.
   */
  seconds?: number
}
