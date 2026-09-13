import type { CompRound } from '../types'

/**
 * Entering match results.
 *
 * Separate from placing tips: a different endpoint, a different person, and a
 * different meaning — a tip is a guess, a result is the fact everything else is
 * scored against. Saving one rescores the whole matchday, which is why the call
 * answers with the round rather than with the match.
 *
 * The whole round goes up, as the pool's own form does, including fixtures that
 * have not been played: those carry "-:-", which is how the backend is already
 * being told "no result" by its own frontend today. Sending only what changed
 * would be tidier and is exactly the kind of guess not worth making against a
 * productive database.
 */
export function useGhtResults() {
  const state = ref<'idle' | 'loading' | 'success' | 'error'>('idle')

  async function save(round: CompRound, results: Record<string, string>) {
    state.value = 'loading'
    try {
      const body = new URLSearchParams(
        Object.fromEntries(round.matches.map(m => [String(m.id), results[m.id] ?? m.result ?? ''])),
      )
      await ghtFetch(`/comprounds/${round.comp}/${round.season}/${round.round}`, { method: 'PUT', body })
      state.value = 'success'
    }
    catch (e) {
      state.value = 'error'
      throw e
    }
  }

  return { state, save }
}

/**
 * Whether to offer the result fields at all.
 *
 * A presentation decision, not a permission: the backend refuses the write from
 * anyone else. Someone who edits this out of the page still cannot save.
 */
export function keepsResults(login: string): boolean {
  return login !== '' && login === (useRuntimeConfig().public.geheimtippAdmin as string)
}
