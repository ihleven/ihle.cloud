import type { TippResponse, TippStatus } from '../types'

/** Where a save got to, for whatever the button shows. */
export type SaveState = 'idle' | 'loading' | 'success' | 'error'

/**
 * Placing tips.
 *
 * One call carries a whole matchday — the matrix sends every changed cell at
 * once, the match page sends one — because the backend takes them that way and
 * answers per match. A tip can be refused on its own (past the kick-off, or by
 * someone not registered for the edition) while the others are taken, so the
 * refusal arrives as an `Error` inside `Status[id]` and not as a failed request.
 * Anything that reports per-item has to be read per item.
 */
export function useGhtTipps() {
  const state = ref<SaveState>('idle')
  /** Keyed by match id — what the backend said about the ones it would not take. */
  const refusals = ref<Record<string, string>>({})

  /**
   * `tipps` maps match id to a result as the pool writes it: "3:1".
   *
   * Answers with the whole round as it now stands, so a caller can replace what
   * it is showing rather than patching its own copy and hoping it matches.
   */
  async function place(comp: string, season: number | string, round: number, tipps: Record<string, string>) {
    state.value = 'loading'
    refusals.value = {}
    try {
      const answer = await ghtFetch<TippResponse>(
        `/comprounds/${comp}/${season}/${round}/tipps`,
        { method: 'PUT', body: new URLSearchParams(tipps) },
      )

      for (const [id, status] of Object.entries(answer.Status ?? {})) {
        const cause = (status as TippStatus)?.Error?.cause
        if (cause) refusals.value[id] = cause
      }

      state.value = Object.keys(refusals.value).length ? 'error' : 'success'
      return answer
    }
    catch (e) {
      // The request itself failed — not signed in, or the pool is unreachable.
      // No tip was taken, so nothing is keyed by match here.
      state.value = 'error'
      throw e
    }
  }

  return { state, refusals, place }
}

/**
 * Whether this tip is still open to whoever is signed in.
 *
 * Both halves matter and neither is enough on its own: your own column stops
 * being editable at kick-off, and someone else's never was. The backend checks
 * the same thing — this only decides whether to offer the field.
 */
export function mayTip(login: string, owner: string, kickoff: string): boolean {
  return login !== '' && login === owner && new Date(kickoff) > new Date()
}
