type Session = {
  iss: string
  sub: string
  aud: string | string[]
  exp: number
  nbf: number
  iat: number
  permissions: Record<string, object>
  // The feature areas this account may see, derived by the server from its
  // permissions. See useModules.
  modules: string[]
  name: string
  email: string
  expiry?: Date
  expires_in?: number
  last_notification: number | undefined
}

type UseAuthOptions = {
  init?: boolean
  interval?: number
  debug?: boolean
}

function sessionDuration(session: Session): number {
  if (!session?.exp)
    return 0
  const now = new Date().getTime()
  const remaining = session.exp * 1000 - now
  return remaining > 0 ? Math.floor(remaining / 1000) : 0
}

export const useAuth = (options?: UseAuthOptions) => {
  const config = useRuntimeConfig()
  const toast = useToast()

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  function debugLog(message?: any, ...optionalParams: any[]) {
    if (options?.debug)
      console.log(message, ...optionalParams)
  }

  const session = useState<Session | undefined>('session')

  // const intervalState = useState<number | undefined>('intervalstate')

  function init() {
    debugLog('init', session.value?.last_notification)
    if (session.value?.last_notification !== undefined) {
      debugLog('useAuth: already initialized')
      return
    }

    setInterval(() => {
      debugLog(' * session interval => ', session.value?.expires_in, '/', session.value?.last_notification)
      if (!session.value)
        return

      session.value.expires_in = sessionDuration(session.value)

      if (session.value.expires_in <= 0) {
        clearSession()
        return
      }

      for (const i of [10, 60, 300, 600, 3600]) {
        if (session.value.expires_in <= i && (session.value.last_notification === undefined || session.value.last_notification > i)) {
          toast.add({ title: 'Session expiring soon', description: 'Your Session will expire in ' + formatDuration(i) })
          session.value.last_notification = i
          return
        }
      }
    }, (options?.interval ?? 5) * 1000)
  }
  if (options?.init) {
    console.log('useAuth init')
    init()
  }

  async function loadSession() {
    try {
      const data = await $fetch<Session | null>(`/auth/session`, {
        baseURL: config.public.api.base,
        credentials: 'include',
      })
      session.value = data ?? undefined

      if (session.value) {
        session.value.expiry = new Date(session.value.exp * 1000)
        session.value.expires_in = sessionDuration(session.value)
        session.value.last_notification = undefined
        debugLog('loaded session => ', data)
      }
    }
    catch (e) {
      // if (e.status !== 401) {
      console.error(' *** session error => ', e.data)
      // }
      session.value = undefined
    }
  }

  function clearSession() {
    session.value = undefined
  }

  async function logout() {
    await $fetch(`/auth/logout`, {
      baseURL: config.public.api.base,
      method: 'POST',
      credentials: 'include',
    })
    clearSession()
  }

  async function loginajax(username: string, password: string): Promise<boolean> {
    try {
      const formData = new FormData()
      formData.append('username', username)
      formData.append('password', password)
      const data = await $fetch<Session>('/auth/login', { baseURL: config.public.api.base, method: 'POST', body: formData, credentials: 'include' })
      if (data) {
        session.value = data
      }
      if (session.value) {
        session.value.expiry = new Date(session.value.exp * 1000)
        session.value.expires_in = sessionDuration(session.value)
      }
      console.log(' logged in for session = ', session.value)
      await land()
      return true
    }
    catch (e) {
      console.error('=>', e)
      // error.value = e.message
      return false
    }
  }

  // land sends an account that exists only for the pool to the pool.
  //
  // Signing in does not navigate — the form is a dialog raised over whatever
  // page was asked for, and closing it reveals that page. For an account with
  // one area that would mean landing on a front page holding one link. It is
  // called from every way of signing in rather than from the password form
  // alone, so a passkey does not behave differently.
  //
  // Already being inside the pool counts as arrived: this must not throw
  // someone back to the entrance on every page load.
  async function land() {
    const modules = session.value?.modules
    if (modules?.length !== 1 || modules[0] !== 'geheimtipp') {
      return
    }
    if (!useRoute().path.startsWith('/geheimtipp')) {
      await navigateTo('/geheimtipp')
    }
  }

  return {
    session,
    loadSession,
    clearSession,
    logout,
    loginajax,
    land,
  }
}
