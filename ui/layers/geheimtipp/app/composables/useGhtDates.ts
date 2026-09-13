// German date formatting, as the pool renders it.
//
// The formatters are built once at module scope rather than per call: Intl
// formatter construction is the expensive part, and the tip matrix asks for a
// date on every one of its rows.

const dmy = new Intl.DateTimeFormat('de-DE', { year: 'numeric', month: 'numeric', day: 'numeric' }).format
const dm = new Intl.DateTimeFormat('de-DE', { month: 'numeric', day: 'numeric' }).format
const d = new Intl.DateTimeFormat('de-DE', { day: 'numeric' }).format
const hm = new Intl.DateTimeFormat('de-DE', { hour: '2-digit', minute: '2-digit' }).format

export function ghtDate(value: string): string {
  return dmy(new Date(value))
}

export function ghtTime(value: string): string {
  return hm(new Date(value))
}

/**
 * The date, but only when it differs from the previous row's.
 *
 * The matrix lists a matchday's fixtures in kick-off order, so repeating
 * "13.9.2026" down a column of eight rows says nothing; the date is printed
 * where it changes and left blank where it does not.
 */
export function ghtDateIfChanged(value: string, previous?: string | null): string {
  if (!previous) return ghtDate(value)
  const a = new Date(value)
  const b = new Date(previous)
  const sameDay = a.getFullYear() === b.getFullYear() && a.getMonth() === b.getMonth() && a.getDate() === b.getDate()
  return sameDay ? '' : ghtDate(value)
}

/**
 * A matchday's window, with the left side shortened to whatever the right side
 * does not already say: "11. – 13.9.2026", "30.8. – 1.9.2026".
 *
 * Empty when the round has no dates yet. The backend sends Go's zero time for
 * those, which formats as "1.1.1" — the pool's own site prints it, and the
 * second half of a season's fixture list is full of them.
 */
export function ghtDateRange(from: string, to: string): string {
  const a = new Date(from)
  const b = new Date(to)
  if (isNaN(a.getTime()) || isNaN(b.getTime())) return ''
  if (a.getFullYear() < 1900 || b.getFullYear() < 1900) return ''

  const left = a.getFullYear() !== b.getFullYear()
    ? dmy(a)
    : a.getMonth() !== b.getMonth()
      ? dm(a) + '.'
      : d(a) + '.'

  return `${left} – ${dmy(b)}`
}
