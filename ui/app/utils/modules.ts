// The areas this app is made of, in one place.
//
// The same handful of facts about a module were written down three times — the
// navigation list, the route guard's prefix map, and the front page's tiles —
// and had to be kept in agreement by hand. They were not: /retro was missing
// from the guard for months. Everything that needs to know about an area now
// derives from here.
//
// What a module gets by way of chrome is deliberately *not* here. That is the
// layout's business: a page with no layout of its own gets the default one, and
// a module that needs something else brings its own. See layers/geheimtipp.

export interface Module {
  /** The entitlement, spelled as the session reports it. */
  id: string
  /** What it is called. */
  label: string
  /** Where it begins — and the prefix the route guard matches on. */
  at: string
  /**
   * The navigation icon.
   *
   * Its presence is what puts the module in the navigation: geheimtipp has no
   * icon because it is reached from the front page and then keeps its own
   * chrome, not this app's.
   */
  icon?: string
  /**
   * The front page tile, where the module has one.
   *
   * Carries its own label because two of them disagree with the navigation on
   * purpose — the tile says CMS and Super 8 where the menu says Content and
   * Filme.
   */
  tile?: { label: string, class: string }
}

/**
 * Every module, in the order they are offered.
 *
 * The order is the priority: the bottom bar shows the first few an account is
 * entitled to and the menu carries the rest, so what is used most belongs
 * nearest the top.
 */
export const appModules: Module[] = [
  { id: 'content', label: 'Content', at: '/entries', icon: 'i-lucide-folder-tree', tile: { label: 'CMS', class: 'bg-blue-500/80' } },
  { id: 'filme', label: 'Filme', at: '/filme', icon: 'i-lucide-clapperboard', tile: { label: 'Super 8', class: 'bg-rose-500/80' } },
  { id: 'familie', label: 'Familie', at: '/famihlie', icon: 'i-lucide-users', tile: { label: 'Familie', class: 'bg-violet-500/80' } },
  { id: 'kalender', label: 'Kalender', at: '/kalender', icon: 'i-lucide-calendar', tile: { label: 'Kalender', class: 'tile-kalender' } },
  { id: 'hidrive', label: 'Dateien', at: '/hidrive', icon: 'i-lucide-hard-drive', tile: { label: 'Dateien', class: 'bg-amber-500/80' } },
  { id: 'mediathek', label: 'Mediathek', at: '/mediathek', icon: 'i-lucide-library', tile: { label: 'Mediathek', class: 'bg-sky-500/80' } },
  { id: 'retro', label: 'Zeitschriften', at: '/retro', icon: 'i-lucide-newspaper', tile: { label: 'Zeitschriften', class: 'bg-cyan-500/80' } },
  { id: 'musik', label: 'Musik', at: '/musik', icon: 'i-lucide-music', tile: { label: 'Musik', class: 'bg-cyan-300/80' } },
  { id: 'djvet', label: 'DJ-Sets', at: '/djvet', icon: 'i-lucide-disc-3', tile: { label: 'DJ-Sets', class: 'tile-oscillate' } },
  { id: 'geheimtipp', label: 'Geheimtipp', at: '/geheimtipp', tile: { label: 'Geheimtipp', class: 'bg-sky-400/80' } },
  { id: 'search', label: 'Search', at: '/search', icon: 'i-feather-search' },
  { id: 'admin', label: 'Konten', at: '/admin', icon: 'i-lucide-shield' },
]

/**
 * Which module a path belongs to, if any.
 *
 * A path belongs to a module when it is the module's own or sits under it —
 * `/musik` and `/musik/anything`, but not `/musikschule`. Paths that belong to
 * no module (the front page, signing in) are common ground.
 */
export function moduleAt(path: string): Module | undefined {
  return appModules.find(m => path === m.at || path.startsWith(m.at + '/'))
}

/** The modules the navigation offers, which is those with an icon. */
export function navModules(): Module[] {
  return appModules.filter(m => m.icon)
}

/** The modules the front page shows, which is those with a tile. */
export function tileModules(): Module[] {
  return appModules.filter(m => m.tile)
}

/**
 * How the bottom bar divides what an account may see.
 *
 * A tab bar holds four and a menu button; everything past that lives in the
 * menu. Given fewer than there is room for, nothing is held back — but the
 * menu button stays regardless, because it carries signing out and the way to
 * anything the bar is too small for.
 */
export function splitForBar(entitled: Module[], room = 4): { shown: Module[], rest: Module[] } {
  return { shown: entitled.slice(0, room), rest: entitled.slice(room) }
}
