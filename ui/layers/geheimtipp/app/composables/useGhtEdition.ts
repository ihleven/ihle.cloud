import type { AusgabeDetail, EditionTipper } from '../types'

/**
 * The edition, fetched once and shared.
 *
 * Nearly every screen under the pool needs it — the matrix takes its column
 * headers from it, the ranking is it — and the pool's own site has the same
 * shape: the parent route loads it and the children read it. Keyed by code, so
 * moving between editions does not serve one from the other.
 */
export function useGhtEdition() {
  const code = useState<string>('ght:edition:code', () => '')
  const ausgabe = useState<AusgabeDetail | null>('ght:edition', () => null)
  const error = useState<unknown>('ght:edition:error', () => null)

  /**
   * The tippers of this edition by login, which is how every other payload
   * refers to them: `evals` and `odds.tipps` are keyed by login and carry no
   * name or avatar of their own.
   */
  const tipperByLogin = computed<Record<string, EditionTipper>>(() =>
    Object.fromEntries((ausgabe.value?.tipper ?? []).map(t => [t.ght, t])),
  )

  /**
   * The competition to read matchdays from. An edition has run exactly one so
   * far; taking it from the payload is still better than the hardcoded 'BL' the
   * pool's own frontend carries.
   */
  const comp = computed(() => ausgabe.value?.competitions?.[0] ?? 'BL')

  /**
   * `force` re-reads an edition already held.
   *
   * Needed because a tipper's name, picture and motto arrive twice — here in
   * `tipper[]`, and in the session as `registration` — so changing them leaves
   * this copy behind. Nothing else invalidates it: within the app this is loaded
   * once and kept, and only a full page load clears it.
   */
  async function load(wanted: string, force = false) {
    if (!wanted) return null
    if (ausgabe.value && code.value === wanted && !force) return ausgabe.value
    try {
      ausgabe.value = await ghtFetch<AusgabeDetail>(`/ausgaben/${encodeURIComponent(wanted)}`)
      code.value = wanted
      error.value = null
    }
    catch (e) {
      ausgabe.value = null
      error.value = e
    }
    return ausgabe.value
  }

  return { ausgabe, tipperByLogin, comp, error, load }
}
