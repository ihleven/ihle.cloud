// One entry in the playing queue.
//
// The address is a finished URL rather than a path and a shelf, because the
// player is shared: the music page and the DJ archive each resolve their own
// routes, and neither has to be known here for both to be heard. What that
// costs is that a queue cannot be rebuilt after its URLs stop working, which is
// one reason nothing tries to carry one across a reload.
interface PlayerTrack {
  /** Where the audio is, ready to hand to an <audio> element. */
  src: string
  title: string
  /** A line under the title — artist and album, or whatever the shelf knows. */
  subtitle?: string
  /** A picture, at the size it will be shown. */
  cover?: string
  /**
   * What the shelf calls this file.
   *
   * Only so a listing can mark the row that is playing. The player never reads
   * it: to the player a track is its URL, and two shelves may well address the
   * same file differently.
   */
  id?: string
}
