// Reading an archive of scanned magazines out of its filenames.
//
// Nobody wrote down what a file in there is called or how the folders are
// arranged; it is thirty years of somebody filing PDFs. Every rule here is a
// reading of what the shelf actually looks like, and each is wrong somewhere —
// which is why they live apart from the fetching, where they can be tested
// against the real names without a browser.

/** A magazine: a folder that holds scans, wherever it sits. */
export interface Magazine {
  /** The path to it, which is the path to every issue in it. */
  dir: string
  /** What to call it: the last part of the path, spaced out. */
  title: string
  groups: Group[]
}

/** A run of issues that belong together — a year, or the specials. */
export interface Group {
  title: string
  issues: Issue[]
}

export interface Issue {
  /** The file, as it is stored. */
  name: string
  /** Where it sits in the archive, which is what the reader route carries. */
  path: string
  /** What to call it: the filename, less the magazine's own name and the .pdf. */
  title: string
  /** The year it appeared, where the filename says one. */
  year?: string
  /** The month within that year, for putting a year's row in order. */
  month?: number
  /** What a tile shows: the month, or the special's name. Short by necessity. */
  label: string
  size?: number
  /** The cover, rendered by the storage from the first page. */
  cover: string
  /** The scan itself. */
  href: string
}

/**
 * The issues of a magazine, gathered into the runs a reader thinks in.
 *
 * By year, because that is what the filenames carry and what somebody looking
 * for an issue remembers. What has no year is a special edition, and those go
 * together at the end: there are a dozen of them and no ordering that would
 * mean anything.
 *
 * A year reads January to December. The filenames do not give that for free:
 * one 1983 issue is called Hobby Computer rather than Happy Computer, and
 * sorting by name alone puts December before November.
 */
export function groupIssues(issues: Issue[]): Group[] {
  const years = new Map<string, Issue[]>()
  const specials: Issue[] = []

  for (const issue of issues) {
    if (!issue.year) {
      specials.push(issue)
      continue
    }
    const run = years.get(issue.year)
    if (run) run.push(issue)
    else years.set(issue.year, [issue])
  }

  const groups = [...years.entries()]
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([title, run]) => ({
      title,
      issues: run.sort((a, b) => (a.month ?? 0) - (b.month ?? 0) || a.name.localeCompare(b.name)),
    }))

  if (specials.length) {
    groups.push({ title: 'Sonderhefte', issues: specials.sort((a, b) => a.name.localeCompare(b.name)) })
  }

  return groups
}

/**
 * A filename read as an issue.
 *
 * "happy-computer-1984.03-Cartman.pdf" in a folder called HappyComputer says
 * its magazine twice, and the folder has already said it, so what is left is
 * which issue this is. The two spell it differently — one runs the words
 * together, the other separates them — so the comparison ignores everything
 * that is not a letter or a digit, and the count of those decides how much of
 * the original name to drop.
 */
export function issueTitle(dir: string, name: string): string {
  const stem = name.replace(/\.pdf$/i, '')
  const wanted = letters(lastSegment(dir))
  if (!wanted || !letters(stem).startsWith(wanted)) {
    return words(stem)
  }

  // Walk the name until as many letters and digits have passed as the
  // magazine's own name holds; what follows is the issue.
  let seen = 0
  let i = 0
  for (; i < stem.length && seen < wanted.length; i++) {
    if (/[a-z0-9]/i.test(stem[i]!)) seen++
  }

  return words(stem.slice(i).replace(/^[^a-z0-9]+/i, '')) || words(stem)
}

/**
 * The year and the month an issue title carries.
 *
 * Regular issues are numbered and dated — "N05.1984.03 Cartman" is the fifth,
 * from March 1984, scanned by Cartman — so the year groups them and the month
 * is all a tile needs to say. A special edition carries neither and keeps its
 * own name, which is the only thing that distinguishes one from another.
 */
export function dated(title: string): { year?: string, month?: number, label: string } {
  const match = /(?:^|[^0-9])((?:19|20)\d{2})[.\-/ ]([01]?\d)(?![0-9])/.exec(title)
  if (!match) {
    return { label: title }
  }
  const month = Number(match[2])

  return { year: match[1], month, label: months[month] ?? match[2]! }
}

/**
 * What to call a magazine, from the folder holding it.
 *
 * Run-together names are separated — "HappyComputer" is two words and was only
 * ever one because a folder name cannot hold a space comfortably. A name that
 * is genuinely one word, like Powerplay, is left as it is.
 */
export function magazineTitle(dir: string): string {
  return words(lastSegment(dir).replace(/([a-z0-9])([A-Z])/g, '$1 $2'))
}

/** The last part of a path: the folder itself, without what it sits in. */
export function lastSegment(path: string): string {
  const parts = path.split('/').filter(Boolean)

  return parts[parts.length - 1] ?? ''
}

/** Written out, because a tile has room for a word and "03" says less. */
const months = [
  '', 'Januar', 'Februar', 'März', 'April', 'Mai', 'Juni',
  'Juli', 'August', 'September', 'Oktober', 'November', 'Dezember',
]

/** A name reduced to what two spellings of it have in common. */
function letters(s: string): string {
  return s.replace(/[^a-z0-9]/gi, '').toLowerCase()
}

function words(slug: string): string {
  return slug
    .replace(/[-_]+/g, ' ')
    .replace(/\s+/g, ' ')
    .trim()
}
