import type { Aktuell, EditionTipper, Geheimtipper } from '../types'

/**
 * Who is signed in to the pool, and which matchdays it is offering.
 *
 * Both answers come from GET /aktuell, so they are one request and one piece of
 * state. That is also why identity is read from the server rather than from the
 * cookie: the pool's own frontend decoded the JWT in the browser to learn the
 * login name, which meant the credential had to be readable by any script on the
 * page. Here the token stays HttpOnly and this call says who it belongs to.
 *
 * The identity is advisory either way — it decides which column of the tip
 * matrix is yours. The backend checks the token on every write.
 */
export function useGhtSession() {
  const aktuell = useState<Aktuell | null>('ght:aktuell', () => null)
  const failed = useState<boolean>('ght:aktuell:failed', () => false)

  /**
   * The signed-in login, or "".
   *
   * This app's session comes first. Whoever is signed in here is who the pool is
   * told about — the proxy mints its credential from the session — so the name
   * is already known without asking, and a pool that is briefly unreachable no
   * longer reads as being signed out. The pool's own answer is the fallback,
   * for a visitor with no account here at all.
   */
  const { session } = useAuth()
  const login = computed(() => {
    if (session.value?.modules.includes('geheimtipp')) return session.value.sub
    return aktuell.value?.authkey || ''
  })
  const signedIn = computed(() => login.value !== '')

  /** The current edition's code, e.g. "2027". Not the `aktuell` flag in /ausgaben, which is set on every edition. */
  const edition = computed(() => aktuell.value?.edition ?? '')

  /** The signed-in person's entry in the current edition, if they have one. Tipping needs it; reading does not. */
  const registration = computed<EditionTipper | null>(() => aktuell.value?.registration ?? null)

  /**
   * The account behind the session — real name, email — as against the
   * registration, which is the name they play an edition under. Their own
   * record, because /aktuell only ever answers about the caller.
   */
  const account = computed<Geheimtipper | null>(() => aktuell.value?.geheimtipper ?? null)

  async function load(force = false) {
    if (aktuell.value && !force) return aktuell.value
    try {
      aktuell.value = await ghtFetch<Aktuell>('/aktuell')
      failed.value = false
    }
    catch {
      // Unreachable is not the same as signed out, and the gate has to tell
      // them apart: one is a login form, the other is an error.
      aktuell.value = null
      failed.value = true
    }
    return aktuell.value
  }

  /**
   * Signing out of the pool is signing out of this app.
   *
   * There is no separate pool session left to end: the credential is minted per
   * forwarded request from this app's session and never given to the browser,
   * so ending the session is what makes it stop being minted. The pool's own
   * /logout is gone with the cookie it used to clear.
   */
  async function signOut() {
    await useAuth().logout()
    aktuell.value = null
  }

  return { aktuell, login, signedIn, edition, registration, account, failed, load, signOut }
}
