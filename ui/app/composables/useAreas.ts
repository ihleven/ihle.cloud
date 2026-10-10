import type { NavigationMenuItem } from '@nuxt/ui'

// The areas of the site, as navigation.
//
// Which areas exist is utils/modules; which of them an account may see is its
// entitlements; what each is about is its concern. This turns the three into
// the shapes the two navigations want — and they want different ones, because
// a row across the top and a column in a panel have different room.
export function useAreas() {
  const route = useRoute()
  const { allowed } = useModules()

  /** The modules this account may see, in the order they are offered. */
  const areas = computed(() => allowed(navModules().map(m => ({ ...m, module: m.id }))))

  // Read through a computed rather than captured once, so the active marker
  // follows navigation as well.
  function at(path: string) {
    return route.path === path || route.path.startsWith(path + '/')
  }

  function link(area: Module): NavigationMenuItem {
    return { icon: area.icon, label: area.label, to: area.at, active: at(area.at) }
  }

  /**
   * The areas under their concerns, in the order the registry offers them.
   *
   * A concern appears where its first area does, so the registry stays the only
   * place the order is decided. Areas with no concern keep their place in the
   * sequence as themselves.
   */
  const grouped = computed(() => {
    const groups: { concern: string, areas: Module[] }[] = []
    for (const area of areas.value) {
      const concern = area.concern ?? ''
      const existing = concern ? groups.find(g => g.concern === concern) : undefined
      if (existing) {
        existing.areas.push(area)

        continue
      }
      groups.push({ concern, areas: [area] })
    }

    return groups
  })

  /** Flat, for the footer's own column of areas. */
  const items = computed<NavigationMenuItem[]>(() => areas.value.map(link))

  return { areas, grouped, items }
}
