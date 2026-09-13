import type { NavigationMenuItem } from '@nuxt/ui'

// The areas of the site, as navigation.
//
// Each entry names the area it belongs to; the account's entitlements decide
// which are offered. Shared rather than repeated, because the bar, the menu and
// the footer all show the same list and would otherwise drift apart.
const areas = [
  { module: 'content', icon: 'i-lucide-folder-tree', label: 'Content', to: '/entries' },
  { module: 'filme', icon: 'i-lucide-clapperboard', label: 'Filme', to: '/filme' },
  { module: 'familie', icon: 'i-lucide-users', label: 'Familie', to: '/famihlie' },
  { module: 'kalender', icon: 'i-lucide-calendar', label: 'Kalender', to: '/kalender' },
  { module: 'mediathek', icon: 'i-lucide-library', label: 'Mediathek', to: '/mediathek' },
  { module: 'musik', icon: 'i-lucide-music', label: 'Musik', to: '/musik' },
  { module: 'search', icon: 'i-feather-search', label: 'Search', to: '/search' },
  { module: 'admin', icon: 'i-lucide-shield', label: 'Konten', to: '/admin' },
]

export function useAreas() {
  const route = useRoute()
  const { allowed } = useModules()

  // Read through a computed rather than captured once, so the active marker
  // follows navigation as well.
  const items = computed<NavigationMenuItem[]>(() =>
    allowed(areas).map(({ icon, label, to }) => ({
      icon,
      label,
      to,
      active: route.path.startsWith(to),
    })),
  )

  return { items }
}
