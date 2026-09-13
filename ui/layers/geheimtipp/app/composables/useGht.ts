// Reaching the pool's backend.
//
// One place holds the base and the credentials, so no page repeats them and
// moving to a shared origin later is still the one config value it was meant to
// be.

function base(): string {
  return useRuntimeConfig().public.geheimtippBase as string
}

/** A read, for use during setup: gives back pending and error to render. */
export function useGht<T>(path: MaybeRefOrGetter<string>) {
  return useFetch<T>(() => toValue(path), { baseURL: base(), credentials: 'include' })
}

/** A read, imperatively — outside setup, or when the result is being cached. */
export function ghtFetch<T>(path: string, options: Record<string, unknown> = {}) {
  return $fetch<T>(path, { baseURL: base(), credentials: 'include', ...options })
}

/**
 * An avatar's URL. `avatar` is the id the payloads carry, and the pixel size is
 * a path segment rather than a query — the pool renders one image per size.
 */
export function ghtAvatar(avatar: number | null | undefined, size: number): string {
  return avatar ? `${base()}/media/avatars/${avatar}/${size}` : ''
}

/**
 * A club crest. Shipped as files rather than read from the payload: the pool's
 * `teams[].logo` is always empty, and its own frontend inlines every crest into
 * one component.
 *
 * Served from /assets rather than /ght, because the proxy owns every path
 * under its mount point — and not from /geheimtipp either, where a directory of
 * files would shadow the SPA route of the same name.
 */
export function ghtCrest(team: string): string {
  return `/assets/geheimtipp/clubs/${team}.svg`
}
