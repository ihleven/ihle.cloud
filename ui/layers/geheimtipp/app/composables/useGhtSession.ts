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

  /** The signed-in login, or "" — the pool sends "" for a caller it does not know. */
  const login = computed(() => aktuell.value?.authkey || '')
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

  async function signIn(username: string, password: string) {
    await ghtFetch('/login', {
      method: 'POST',
      body: new URLSearchParams({ username, password }),
    })
    await load(true)
  }

  async function signOut() {
    await ghtFetch('/logout', { method: 'POST' }).catch(() => {})
    await load(true)
  }

  return { aktuell, login, signedIn, edition, registration, account, failed, load, signIn, signOut }
}
