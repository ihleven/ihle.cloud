// The feature areas the signed-in account may see.
//
// The server derives these from the account's permissions and puts them in the
// session, so the client never has to know that an entitlement is really a
// module.<id> permission. No session means no areas: an anonymous visitor is
// entitled to nothing.
//
// This is presentation, not protection. Hiding a link keeps an area out of
// someone's way; it does not stop them fetching the API by hand. Access itself
// is the server's business.
export function useModules() {
  const { session } = useAuth()

  const entitled = computed<string[]>(() => session.value?.modules ?? [])

  function may(module?: string): boolean {
    // Something with no area is common ground — the home page, sign-in.
    return !module || entitled.value.includes(module)
  }

  // Filters any list of things that name an area.
  function allowed<T extends { module?: string }>(items: T[]): T[] {
    return items.filter(item => may(item.module))
  }

  return { entitled, may, allowed }
}
