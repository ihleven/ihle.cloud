import type { NavigationMenuItem } from '@nuxt/ui'

// The areas of the site, as navigation.
//
// Which areas exist is utils/modules; which of them an account may see is its
// entitlements. This turns the two into the shape the menu wants.
export function useAreas() {
  const route = useRoute()
  const { allowed } = useModules()

  /** The modules this account may see, in the order they are offered. */
  const areas = computed(() => allowed(navModules().map(m => ({ ...m, module: m.id }))))

  // Read through a computed rather than captured once, so the active marker
  // follows navigation as well.
  const items = computed<NavigationMenuItem[]>(() =>
    areas.value.map(({ icon, label, at }) => ({
      icon,
      label,
      to: at,
      active: route.path === at || route.path.startsWith(at + '/'),
    })),
  )

  return { items, areas }
}
