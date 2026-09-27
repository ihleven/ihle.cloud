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
   * What the area is about, and so which heading the navigation files it under.
   *
   * Eleven areas in one row is a list to read rather than a place to aim at.
   * Grouped, the bar carries three words and the areas are one step behind the
   * one they belong to.
   *
   * Without it an area stays where it is, at the top level: a concern of one is
   * a heading that says nothing. The groups appear in the order their first
   * member does, so this list stays the only place the order is decided.
   */
  concern?: string
  /**
   * The front page tile, where the module has one.
   *
   * Carries its own label because two of them disagree with the navigation on
   * purpose — the tile says CMS and Super 8 where the menu says Content and
   * Filme.
   *
   * photo is a picture of the thing the area holds, drawn behind the label.
   * Only the media areas have one, and that is the point: a photograph says
   * what an area is faster than its name does, and it only works where there is
   * something to photograph. Verwaltung and Familie keep their flat colours,
   * which is also what separates the two halves of the grid at a glance.
   *
   * The colour stays on a tile that has a photo. It is what shows while the
   * picture loads, and what shows instead of it if it never arrives.
   *
   * shapes limits what the front page's mosaic may make of it, where the tile
   * has something inside it with a shape of its own. Left out, it takes any of
   * them — most tiles are a word on a background and do not care.
   */
  tile?: { label: string, class: string, photo?: string, shapes?: TileShape[] }
}

/**
 * Every module, in the order they are offered.
 *
 * The order is the priority: the bottom bar shows the first few an account is
 * entitled to and the menu carries the rest, so what is used most belongs
 * nearest the top.
 */
export const appModules: Module[] = [
  { concern: 'Verwaltung', id: 'content', label: 'Content', at: '/content', icon: 'i-lucide-folder-tree', tile: { label: 'CMS', class: 'bg-blue-500/80' } },
  { concern: 'Medien', id: 'filme', label: 'Filme', at: '/filme', icon: 'i-lucide-clapperboard', tile: { label: 'Super 8', class: 'bg-rose-500/80', photo: '/schmalfilm.avif' } },
  { concern: 'Familie', id: 'familie', label: 'Familie', at: '/famihlie', icon: 'i-lucide-users', tile: { label: 'Familie', class: 'bg-violet-500/80' } },
  { concern: 'Familie', id: 'kalender', label: 'Kalender', at: '/kalender', icon: 'i-lucide-calendar', tile: { label: 'Kalender', class: 'tile-kalender', shapes: ['small', 'tall'] } },
  { concern: 'Verwaltung', id: 'hidrive', label: 'Dateien', at: '/hidrive', icon: 'i-lucide-hard-drive', tile: { label: 'Dateien', class: 'bg-amber-500/80' } },
  { concern: 'Medien', id: 'mediathek', label: 'Mediathek', at: '/mediathek', icon: 'i-lucide-library', tile: { label: 'Mediathek', class: 'bg-sky-500/80', photo: '/kino-rote-samtsitze.jpg' } },
  { concern: 'Medien', id: 'retro', label: 'Zeitschriften', at: '/retro', icon: 'i-lucide-newspaper', tile: { label: 'Zeitschriften', class: 'bg-cyan-500/80', photo: '/zeitschriftenstapel.avif' } },
  { concern: 'Medien', id: 'musik', label: 'Musik', at: '/musik', icon: 'i-lucide-music', tile: { label: 'Musik', class: 'bg-cyan-300/80', photo: '/schallplatten.jpg' } },
  { concern: 'Medien', id: 'djvet', label: 'DJ-Sets', at: '/djvet', icon: 'i-lucide-disc-3', tile: { label: 'DJ-Sets', class: 'tile-oscillate' } },
  { id: 'geheimtipp', label: 'Geheimtipp', at: '/geheimtipp', tile: { label: 'Geheimtipp', class: 'bg-sky-400/80' } },
  { id: 'search', label: 'Search', at: '/search', icon: 'i-feather-search' },
  { concern: 'Verwaltung', id: 'admin', label: 'Konten', at: '/admin', icon: 'i-lucide-shield' },

  // Unpublished work, which the CMS layer serves as a panel over the content
  // module. Same id as the editor above on purpose: unpublished work belongs to
  // whoever may edit, so it is covered by the same entitlement rather than
  // needing one of its own. No icon and no tile — it is reached from the count
  // in the module's own bar, when there is anything to reach.
  //
  // Under /content, so the guard would cover it even without this line. It
  // stays because the id it belongs to is not obvious from the path, and
  // because a reader looking for where drafts are guarded should find them
  // named.
  { id: 'content', label: 'Entwürfe', at: '/content/drafts' },

  // The art archive. No icon and no tile on purpose: this entry exists so the
  // route guard knows /werke belongs to somebody — without it the page was
  // reachable by any account, since the guard only gates prefixes named here.
  // Being absent from the navigation and the front page is then what keeps it
  // unadvertised while the archive is still being built out.
  { id: 'art', label: 'Kunst', at: '/werke' },
]

/**
 * Which module a path belongs to, if any.
 *
 * A path belongs to a module when it is the module's own or sits under it —
 * `/musik` and `/musik/anything`, but not `/musikschule`. Paths that belong to
 * no module (the front page, signing in) are common ground.
 */
/**
 * What a concern is drawn as where it stands for its areas.
 *
 * Only the navigation needs this, and only for the concerns — an area brings
 * its own icon. Kept beside the list rather than repeated on every module that
 * belongs to one, and a concern with no entry here simply shows its name.
 */
export const concernIcons: Record<string, string> = {
  Verwaltung: 'i-lucide-sliders-horizontal',
  Medien: 'i-lucide-play',
  Familie: 'i-lucide-users',
}

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
