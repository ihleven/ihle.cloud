// The parts of a date, as the calendar shows them.
//
// Separated from the components so the week number can be tested against the
// dates that break it — the turn of the year, where the ISO week belongs to a
// different year than the date does.

/** The month, 1-based, as a route carries it. */
export function monthOf(date: Date): number {
  return date.getMonth() + 1
}

/** The month's name, in German: September, not 9. */
export function monthName(date: Date): string {
  return date.toLocaleDateString('de-DE', { month: 'long' })
}

/** The weekday's name, in German. */
export function weekdayName(date: Date): string {
  return date.toLocaleDateString('de-DE', { weekday: 'long' })
}

/**
 * The ISO week, and the year that week belongs to.
 *
 * Two fields rather than one, because they disagree for a few days every year:
 * the 1st of January 2027 falls in week 53 of 2026, and a link to week 53 of
 * 2027 would be a link to a week that has not happened. ISO counts the week
 * containing the year's first Thursday as week 1, which is what the shift to
 * Thursday below is doing.
 */
export function isoWeek(date: Date): { year: number, week: number } {
  // Work in UTC so a summer-time boundary cannot move the date by a day.
  const thursday = new Date(Date.UTC(date.getFullYear(), date.getMonth(), date.getDate()))

  // getUTCDay has Sunday as 0; ISO weeks run Monday to Sunday, so Sunday is 7.
  const weekday = thursday.getUTCDay() || 7
  thursday.setUTCDate(thursday.getUTCDate() + 4 - weekday)

  const firstOfYear = new Date(Date.UTC(thursday.getUTCFullYear(), 0, 1))
  const days = (thursday.getTime() - firstOfYear.getTime()) / 86400000

  return { year: thursday.getUTCFullYear(), week: Math.ceil((days + 1) / 7) }
}

// Where each part of a date lives.
//
// Built here rather than written into the markup because two things address
// these routes — the date block's own links and the picker beside it — and the
// picker hands back months counted from one while JavaScript's Date counts them
// from zero. The old page mixed the two and every month link pointed at the
// month before.

/** The page for a year. */
export function yearPath(year: number): string {
  return `/kalender/${year}`
}

/** The page for a month, which is counted from one. */
export function monthPath(year: number, month: number): string {
  return `/kalender/${year}/${month}`
}

/** The page for a single day. */
export function dayPath(year: number, month: number, day: number): string {
  return `/kalender/${year}/${month}/${day}`
}

/**
 * The page for an ISO week.
 *
 * Lower-case `kw`, because that is what the route is: the old page linked to
 * `KW38` while the file behind it is `kw[kw].vue`, so even once the number was
 * worked out the link would not have matched.
 */
export function weekPath(year: number, week: number): string {
  return `/kalender/${year}/kw${week}`
}
