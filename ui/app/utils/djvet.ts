// Reading the DJ archive out of its folders.
//
// The arrangement is deliberate, unlike the record shelf's: a series holds a run
// of CDs, a CD is a folder of mp3s with its sleeve and a photograph of its
// tracklist beside them. So the folders are trusted here, and what the files say
// about themselves is read from their tags rather than from their names — the
// filename is the fallback, not the source.

import type { Track } from './musik'

/** A run of CDs that belong together. */
export interface DjSeries {
  /** The folder, relative to the archive — and the route's first segment. */
  slug: string
  title: string
  /** What it is. Provisional; see seriesDescription. */
  description: string
  cds: Cd[]
}

/** One CD of a series. */
export interface Cd {
  /** Where it is, relative to the archive: `<series>/<cd>`. */
  dir: string
  /** The route's second segment. */
  slug: string
  title: string
  /** The sleeve, at the width it is shown. */
  cover: string
}

/** A CD with what is on it, which costs two more requests to know. */
export interface CdDetail extends Cd {
  series: string
  /** The tracklist, photographed rather than written out. */
  tracklist: string
  tracks: DjTrack[]
}

/**
 * A track, named by its tags where it has them.
 *
 * Structurally a musik Track with the tag's answers added, so the shelf's own
 * sorting works on it unchanged.
 */
export interface DjTrack extends Track {
  artist?: string
  /** How long it runs, already written out. */
  length?: string
}

/**
 * What a series is called, and what it is.
 *
 * Hard-coded, and meant to move: the text belongs beside the music or in the
 * CMS. Pages call the two functions below and never this map, so when the text
 * moves only they change.
 *
 * The names are here because they are not derivable from the folders — `rxc` is
 * RXC and not Rxc, `ambient-meets-acid` is Ambient meets Acid and not Ambient
 * Meets Acid. A series that is not listed gets its slug spaced out and says
 * nothing about itself, which is better than being confidently mis-titled.
 */
const known: Record<string, { name: string, description: string }> = {
  'pool-party': { name: 'Pool Party', description: 'Beschreibung folgt.' },
  'lunar-suite': { name: 'Lunar Suite', description: 'Beschreibung folgt.' },
  'rxc': { name: 'RXC', description: 'Beschreibung folgt.' },
  'lounge-club': { name: 'Lounge Club', description: 'Beschreibung folgt.' },
  'rudi-bar': { name: 'Rudi Bar', description: 'Beschreibung folgt.' },
  'ambient-meets-acid': { name: 'Ambient meets Acid', description: 'Beschreibung folgt.' },
}

/** What to call a series. */
export function seriesTitle(slug: string): string {
  return known[slug]?.name ?? spaced(slug)
}

/**
 * What a series is.
 *
 * The one place a page asks, so that the answer can come from somewhere else
 * later without any page knowing it changed.
 */
export function seriesDescription(slug: string): string {
  return known[slug]?.description ?? ''
}

/** What to call a CD, which is its folder spaced out. */
export function cdTitle(slug: string): string {
  return spaced(slug)
}

/** A slug as a title: separators become spaces, each word starts upper-case. */
function spaced(slug: string): string {
  return slug
    .split(/[-_]+/)
    .filter(Boolean)
    .map(word => word.charAt(0).toUpperCase() + word.slice(1))
    .join(' ')
}

/**
 * A length, written out.
 *
 * Named for what it is rather than `clock`, which several components already
 * define privately; auto-importing a fourth spelling of the same idea under
 * that name would make which one is in scope a question.
 */
export function duration(seconds?: number): string {
  if (!seconds || seconds < 0) return ''

  const whole = Math.round(seconds)
  const minutes = Math.floor(whole / 60)
  const rest = whole % 60

  return `${minutes}:${String(rest).padStart(2, '0')}`
}

/**
 * What one file is called, preferring what it says about itself.
 *
 * The tags are the source and the filename the fallback, which is the archive's
 * one real difference from the record shelf: these files were written with
 * their titles in them, and the filenames are frequently just a number.
 */
export function named(name: string, tag?: MusikTags): { title: string, number?: number } {
  const fallback = trackTitle(name)

  return {
    title: tag?.title || fallback.title,
    number: tag?.track || fallback.number,
  }
}
