import { computed, ref } from 'vue'
import { beforeEach, describe, expect, it } from 'vitest'

// The composable reaches for Nuxt's auto-imports. Standing them up here rather
// than mocking the composable means these tests run the code that ships: a
// keyed store, so two callers of usePlayer() share one queue exactly as they do
// in the app.
const states = new Map<string, ReturnType<typeof ref>>()

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const g = globalThis as any
g.useState = <T>(key: string, init: () => T) => {
  if (!states.has(key)) states.set(key, ref(init()))

  return states.get(key)
}
g.computed = computed

const { usePlayer } = await import('../app/composables/usePlayer')

function track(title: string): PlayerTrack {
  return { src: `https://example.test/${title}.mp3`, title, id: `dir/${title}.mp3` }
}

const album = [track('one'), track('two'), track('three')]

beforeEach(() => states.clear())

describe('usePlayer', () => {
  it('plays nothing until something is queued', () => {
    const player = usePlayer()
    expect(player.current.value).toBeUndefined()
    expect(player.queue.value).toEqual([])
  })

  it('takes a whole list and starts where it is told', () => {
    const player = usePlayer()
    player.play(album, 1)
    expect(player.current.value?.title).toBe('two')
    expect(player.queue.value).toHaveLength(3)
  })

  it('starts at the beginning when told nothing', () => {
    const player = usePlayer()
    player.play(album)
    expect(player.current.value?.title).toBe('one')
  })

  it('refuses an empty list rather than opening on nothing', () => {
    const player = usePlayer()
    player.play([])
    expect(player.current.value).toBeUndefined()
  })

  it('clamps a position the list does not have', () => {
    const player = usePlayer()
    player.play(album, 99)
    expect(player.current.value?.title).toBe('three')

    player.play(album, -4)
    expect(player.current.value?.title).toBe('one')
  })

  it('advances through the list', () => {
    const player = usePlayer()
    player.play(album)
    player.next()
    expect(player.current.value?.title).toBe('two')
    player.next()
    expect(player.current.value?.title).toBe('three')
  })

  // The end of an album is silence, not a bar left on screen claiming
  // otherwise.
  it('stops at the end instead of staying open', () => {
    const player = usePlayer()
    player.play(album, 2)
    expect(player.hasNext.value).toBe(false)

    player.next()
    expect(player.current.value).toBeUndefined()
    expect(player.queue.value).toEqual([])
  })

  it('goes back, and not past the start', () => {
    const player = usePlayer()
    player.play(album, 1)
    player.previous()
    expect(player.current.value?.title).toBe('one')

    player.previous()
    expect(player.current.value?.title).toBe('one')
    expect(player.hasPrevious.value).toBe(false)
  })

  it('leaves nothing behind when stopped', () => {
    const player = usePlayer()
    player.play(album)
    player.stop()
    expect(player.current.value).toBeUndefined()
    expect(player.queue.value).toEqual([])
  })

  // The reason the state is keyed rather than held in the module: the page that
  // queues an album and the component that plays it are different callers, and
  // the layout reserving space at the bottom is a third.
  it('is one queue however many callers ask for it', () => {
    const page = usePlayer()
    const component = usePlayer()

    page.play(album, 1)
    expect(component.current.value?.title).toBe('two')

    component.next()
    expect(page.current.value?.title).toBe('three')
  })
})
