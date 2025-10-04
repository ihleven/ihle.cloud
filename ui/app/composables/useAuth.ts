type Session = {
  name: string
  email: string
  exp: number
  expiry?: Date
  expires_in?: number
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

  function debugLog(message?: any, ...optionalParams: any[]) {
    if (options?.debug)
      console.log(message, ...optionalParams)
  }

  const session = useState<Session | undefined>('session')

  const intervalState = useState<number | undefined>('intervalstate')

  function init() {
    if (intervalState.value !== undefined) {
      debugLog('useAuth: already initialized')
      return
    }

    setInterval(() => {
      debugLog(' * session interval => ', session.value?.expires_in, '/', intervalState.value)
      if (!session.value)
        return

      const duration = sessionDuration(session.value)

      session.value.expires_in = duration
      if (duration <= 30 && (!intervalState.value || intervalState.value > 30)) {
        toast.add({ title: 'Session expiring soon', description: 'Your Session will expire in ' + duration + ' seconds' })
        intervalState.value = 30
      }
      if (duration <= 0 && (intervalState.value)) {
        console.log('expired ')
        intervalState.value = 0
        clearSession()
      }
    }, (options?.interval ?? 5) * 1000)

    intervalState.value = Number.MAX_SAFE_INTEGER
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
        console.log('loaded session => ', data)
      }
    }
    catch (e) {
      console.error('=>', e)
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

  // async function login2() {
  //   await navigateTo(`${config.public.apiBaseURL}/tool-api/auth/login?redirect=${window.location}`, { external: true })
  // }

  function login(target: string) {
    const redirect = typeof target === 'string' && target !== '' ? `?redirect=${target}` : ''
    navigateTo('/api/auth/login' + redirect, { external: true })
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
      return true
    }
    catch (e) {
      console.error('=>', e)
      // error.value = e.message
      return false
    }
  }

  return {
    session,
    loadSession,
    clearSession,
    logout,
    login,
    loginajax,
  }
}
