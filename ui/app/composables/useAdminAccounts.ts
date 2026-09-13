// Account administration.
//
// The same operations `ihlvn account` performs from a terminal. Every call goes
// to an endpoint gated on the admin entitlement server-side, so this composable
// carries no access logic of its own — hiding the section in the navigation is
// presentation, and the API is what actually refuses.

export type AdminAccount = {
  name: string
  display_name: string
  email: string
  disabled: boolean
  groups: string[]
  permissions: string[]
  // Permissions this build does not define. Stored, and look like grants, but a
  // scope drops them — worth showing rather than leaving as a silent discrepancy.
  unregistered_permissions: string[]
  // What the permissions add up to: the areas this account may see.
  modules: string[]
  has_password: boolean
  passkeys: number
  created_at: string
}

export type AdminPasskey = {
  id: string
  name: string
  syncable: boolean
  created_at: string
  last_used_at: string | null
}

export type PasswordAdvice = {
  length: number
  too_short: boolean
  breaches: number
  unchecked?: string
}

// A password is set unless the screening had something to say and the caller has
// not yet said to use it anyway.
export type PasswordResult = { set: boolean, advice: PasswordAdvice }

export type Enrollment = { url: string, expires_at: string }

export type AccountEdit = {
  display_name: string
  email: string
  groups: string[]
  permissions: string[]
  disabled: boolean
}

export function useAdminAccounts() {
  // The admin endpoints hang off this app's API prefix, not the bare origin the
  // /auth routes use.
  const base = useRuntimeConfig().public.apiBaseURL as string

  function call<T>(path: string, options: Record<string, unknown> = {}): Promise<T> {
    return $fetch<T>(`/admin/accounts${path}`, { baseURL: base, credentials: 'include', ...options })
  }

  const list = () => call<AdminAccount[]>('')

  // The areas this build defines, which is what may be granted. Read from the
  // server rather than listed here, so the picker cannot offer a name that
  // resolves to nothing.
  const areas = () => $fetch<string[]>('/admin/modules', { baseURL: base, credentials: 'include' })
  const get = (name: string) => call<AdminAccount>(`/${encodeURIComponent(name)}`)

  const create = (body: { name: string, display_name: string, email: string }) =>
    call<AdminAccount>('', { method: 'POST', body })

  const update = (name: string, body: AccountEdit) =>
    call<AdminAccount>(`/${encodeURIComponent(name)}`, { method: 'PUT', body })

  // confirm carries the decision to use a password the screening objected to.
  // Without it the server answers with the advice and sets nothing, which is
  // what lets the form show the objection before anything is written.
  const setPassword = (name: string, password: string, confirm = false) =>
    call<PasswordResult>(`/${encodeURIComponent(name)}/password`, {
      method: 'POST',
      body: { password, confirm },
    })

  const enroll = (name: string) =>
    call<Enrollment>(`/${encodeURIComponent(name)}/enroll`, { method: 'POST' })

  const passkeys = (name: string) =>
    call<AdminPasskey[]>(`/${encodeURIComponent(name)}/passkeys`)

  const removePasskey = (name: string, id: string) =>
    call<void>(`/${encodeURIComponent(name)}/passkeys/${encodeURIComponent(id)}`, { method: 'DELETE' })

  // Everything the account can sign in with, and everywhere it is signed in.
  const revoke = (name: string) =>
    call<{ passkeys: number, sessions: number }>(`/${encodeURIComponent(name)}/passkeys`, { method: 'DELETE' })

  const signOutEverywhere = (name: string) =>
    call<{ sessions: number }>(`/${encodeURIComponent(name)}/sessions`, { method: 'DELETE' })

  // The server's own message, which says what was refused and why — "an
  // administrator cannot remove their own admin entitlement" is more use than
  // "Request failed".
  //
  // The error middleware encodes an error as a bare JSON string rather than an
  // object, so that is what arrives as `data`.
  function describe(e: unknown): string {
    const err = e as { data?: unknown, message?: string }
    if (typeof err?.data === 'string' && err.data) return err.data
    const data = err?.data as { error?: string, message?: string } | undefined
    return data?.error || data?.message || err?.message || 'Something went wrong.'
  }

  return {
    list, areas, get, create, update,
    setPassword, enroll,
    passkeys, removePasskey, revoke, signOutEverywhere,
    describe,
  }
}
