// What is playing, for the whole app.
//
// The queue lives here rather than on a page because the sound has to outlive
// the page. An <audio> element written into a route's template is unmounted the
// moment you navigate, and an unmounted element is a silent one — so the single
// element is mounted by Player.vue, outside the router's reach, and this is
// what it plays.
//
// It is a queue rather than a track for the same reason. Advancing meant
// searching the page's loaded shelf for the file that had just finished, which
// works exactly as long as that page is on screen and still holds that album. A
// queue carries its own succession and needs nobody.
//
// Nothing exists until something is played: the state is an empty array and the
// component draws nothing, so somebody who never presses play never has an
// audio element, a decoder or a request.
export function usePlayer() {
  const queue = useState<PlayerTrack[]>('player:queue', () => [])
  const index = useState<number>('player:index', () => 0)

  const current = computed(() => queue.value[index.value])
  const hasNext = computed(() => index.value < queue.value.length - 1)
  const hasPrevious = computed(() => index.value > 0)

  /**
   * Play a list, starting at one of them.
   *
   * The whole list rather than the one track chosen: what follows it is part of
   * the choice. Handing over a single track is what tied advancing to the page
   * that happened to know the rest.
   */
  function play(tracks: PlayerTrack[], at = 0) {
    if (!tracks.length) return

    queue.value = tracks
    index.value = Math.min(Math.max(at, 0), tracks.length - 1)
  }

  /**
   * The next one, or silence.
   *
   * Running off the end closes the player rather than leaving it up: a bar that
   * is still on screen says something is still playing, and at the end of an
   * album nothing is.
   */
  function next() {
    if (hasNext.value) {
      index.value++

      return
    }
    stop()
  }

  function previous() {
    if (hasPrevious.value) index.value--
  }

  /** Stop, and leave nothing behind — the element goes with the queue. */
  function stop() {
    queue.value = []
    index.value = 0
  }

  return { queue, index, current, hasNext, hasPrevious, play, next, previous, stop }
}
