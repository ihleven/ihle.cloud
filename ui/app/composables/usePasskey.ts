// Passkey ceremonies.
//
// WebAuthn cannot be driven declaratively: it needs navigator.credentials, and
// the browser API wants ArrayBuffers where the server speaks base64url — that is
// how the WebAuthn JSON encoding represents binary. So every ceremony is a
// decode on the way in and an encode on the way out.
//
// The begin endpoints take the account's password as a form field, because
// adding or removing a credential is a step-up: holding a session is not enough.
// The finish endpoints take JSON, because the server parses the credential
// itself out of the body.

type PasskeyInfo = {
  id: string
  name: string
  syncable: boolean
  created_at: string
  last_used_at: string | null
}

function fromBase64url(value: string): Uint8Array {
  const padded = value.replace(/-/g, '+').replace(/_/g, '/')
  const raw = atob(padded + '='.repeat((4 - (padded.length % 4)) % 4))
  const bytes = new Uint8Array(raw.length)
  for (let i = 0; i < raw.length; i++) bytes[i] = raw.charCodeAt(i)
  return bytes
}

function toBase64url(buffer: ArrayBuffer): string {
  const bytes = new Uint8Array(buffer)
  let s = ''
  for (let i = 0; i < bytes.length; i++) s += String.fromCharCode(bytes[i]!)
  return btoa(s).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '')
}

// Serialised by hand rather than with credential.toJSON(), which is not
// available in every browser yet.
function encodeAttestation(c: PublicKeyCredential) {
  const response = c.response as AuthenticatorAttestationResponse
  return {
    id: c.id,
    rawId: toBase64url(c.rawId),
    type: c.type,
    authenticatorAttachment: c.authenticatorAttachment || undefined,
    clientExtensionResults: c.getClientExtensionResults(),
    response: {
      clientDataJSON: toBase64url(response.clientDataJSON),
      attestationObject: toBase64url(response.attestationObject),
    },
  }
}

function encodeAssertion(c: PublicKeyCredential) {
  const response = c.response as AuthenticatorAssertionResponse
  return {
    id: c.id,
    rawId: toBase64url(c.rawId),
    type: c.type,
    authenticatorAttachment: c.authenticatorAttachment || undefined,
    clientExtensionResults: c.getClientExtensionResults(),
    response: {
      clientDataJSON: toBase64url(response.clientDataJSON),
      authenticatorData: toBase64url(response.authenticatorData),
      signature: toBase64url(response.signature),
      userHandle: response.userHandle ? toBase64url(response.userHandle) : null,
    },
  }
}

export function usePasskey() {
  const config = useRuntimeConfig()
  const base = config.public.api.base as string

  // unavailable says why passkeys cannot be used here, or '' when they can.
  //
  // It reports a reason rather than just a boolean because the usual cause is
  // not the browser at all: WebAuthn is only exposed in a secure context, so
  // reaching the app over plain http on anything but localhost hides it. A
  // button that silently disappears gives no way to work that out.
  function unavailable(): string {
    if (typeof window === 'undefined') return 'Not available yet.'
    if (!window.isSecureContext) {
      return 'Passkeys need a secure context: https, or localhost in development.'
    }
    if (!window.PublicKeyCredential || !navigator.credentials) {
      return 'This browser does not support passkeys.'
    }
    return ''
  }

  function supported(): boolean {
    return unavailable() === ''
  }

  // The ceremony a begin opened. Sent back on the matching finish, because a
  // browser may have two open at once — one offered in the username field's
  // autofill, one started by the button — and a single cookie cannot name both.
  const ceremonyHeader = 'X-Ceremony'

  function post<T>(path: string, body?: unknown, form?: Record<string, string>, ceremony?: string): Promise<T> {
    return $fetch<T>(path, {
      baseURL: base,
      method: 'POST',
      credentials: 'include',
      headers: ceremony ? { [ceremonyHeader]: ceremony } : undefined,
      body: form ? new URLSearchParams(form) : body,
    })
  }

  // Whether the browser can offer a passkey inside the username field's autofill.
  //
  // There is deliberately no way to ask "does this person have a passkey for
  // this site" — that would let any page enumerate accounts. Conditional
  // mediation is the answer to the same need: the browser offers a credential if
  // there is one and stays silent if there is not, so nothing has to be chosen
  // up front.
  async function conditionalAvailable(): Promise<boolean> {
    if (!supported()) return false
    const pkc = window.PublicKeyCredential as unknown as {
      isConditionalMediationAvailable?: () => Promise<boolean>
    }
    if (!pkc.isConditionalMediationAvailable) return false
    try {
      return await pkc.isConditionalMediationAvailable()
    }
    catch {
      return false
    }
  }

  // Sign in without a username: the authenticator says which credential it
  // holds, and the server resolves the account from its handle.
  //
  // The new session is not returned: the caller refreshes it through useAuth, so
  // there is one path that owns session state rather than two.
  //
  // With mediation 'conditional' the browser attaches the offer to a field
  // marked autocomplete="webauthn" instead of opening a dialog, and the promise
  // stays pending until the person picks one — so it is started in the
  // background and aborted if they do something else.
  async function signIn(mediation?: CredentialMediationRequirement, signal?: AbortSignal): Promise<void> {
    const options = await post<{ publicKey: PublicKeyCredentialRequestOptions, ceremony: string }>('/auth/passkey/login/begin')

    const publicKey = options.publicKey as unknown as Record<string, unknown>
    publicKey.challenge = fromBase64url(publicKey.challenge as unknown as string)
    if (Array.isArray(publicKey.allowCredentials)) {
      publicKey.allowCredentials = publicKey.allowCredentials.map(
        (c: { id: string }) => ({ ...c, id: fromBase64url(c.id) }),
      )
    }

    const credential = await navigator.credentials.get({
      publicKey: publicKey as unknown as PublicKeyCredentialRequestOptions,
      mediation,
      signal,
    }) as PublicKeyCredential | null
    if (!credential) throw new Error('no credential was returned')

    await post('/auth/passkey/login/finish', encodeAssertion(credential), undefined, options.ceremony)
  }

  // Enrol a credential. Authorised either by being signed in or by holding an
  // enrollment link, and in both cases by the account's password.
  async function enrol(password: string, name: string): Promise<void> {
    const options = await post<{ publicKey: PublicKeyCredentialCreationOptions, ceremony: string }>(
      '/auth/passkey/register/begin', undefined, { password },
    )

    const publicKey = options.publicKey as unknown as Record<string, unknown>
    publicKey.challenge = fromBase64url(publicKey.challenge as unknown as string)
    const user = publicKey.user as { id: string }
    user.id = fromBase64url(user.id) as unknown as string
    if (Array.isArray(publicKey.excludeCredentials)) {
      publicKey.excludeCredentials = publicKey.excludeCredentials.map(
        (c: { id: string }) => ({ ...c, id: fromBase64url(c.id) }),
      )
    }

    const credential = await navigator.credentials.create({
      publicKey: publicKey as unknown as PublicKeyCredentialCreationOptions,
    }) as PublicKeyCredential | null
    if (!credential) throw new Error('no credential was created')

    await $fetch(`/auth/passkey/register/finish?name=${encodeURIComponent(name)}`, {
      baseURL: base,
      method: 'POST',
      credentials: 'include',
      headers: { [ceremonyHeader]: options.ceremony },
      body: encodeAttestation(credential),
    })
  }

  function list(): Promise<PasskeyInfo[]> {
    return $fetch<PasskeyInfo[]>('/auth/passkey', { baseURL: base, credentials: 'include' })
  }

  function remove(id: string, password: string): Promise<void> {
    return $fetch(`/auth/passkey/${id}`, {
      baseURL: base,
      method: 'DELETE',
      credentials: 'include',
      body: new URLSearchParams({ password }),
    })
  }

  // The browser reports an already-registered authenticator this way rather than
  // as a plain failure, and it is worth saying so instead of "registration
  // failed".
  // An aborted conditional attempt is routine — the person typed a password
  // instead, or navigated away — and must not surface as a failure.
  function aborted(e: unknown): boolean {
    return (e as { name?: string })?.name === 'AbortError'
  }

  function describe(e: unknown): string {
    const err = e as { name?: string, data?: { message?: string }, message?: string }
    if (err?.name === 'InvalidStateError') return 'This device is already registered on your account.'
    if (err?.name === 'NotAllowedError') return 'Cancelled, or the device did not respond in time.'
    return err?.data?.message || err?.message || 'Something went wrong.'
  }

  return { supported, unavailable, conditionalAvailable, signIn, enrol, list, remove, aborted, describe }
}
