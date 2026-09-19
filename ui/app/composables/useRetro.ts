// Reaching the magazine archive.
//
// The same arrangement as the Mediathek, and for the same reasons: one fixed
// shelf on the family storage, the same for everybody entitled to it, so the
// answers can be cached by address. What differs is what is on the shelf —
// scanned magazines, one PDF per issue, each of them fifty to a hundred and
// twenty megabytes — which is why a listing shows covers the storage renders
// rather than anything this app would have to download to look at.
//
// What the names and the folders mean is read in utils/retro, away from here,
// so it can be tested against the real filenames without a browser.

function base(): string {
  return useRuntimeConfig().public.apiBaseURL as string
}

/** A path as the API wants it in a route: each segment encoded, slashes kept. */
function encodePath(path: string): string {
  return path.split('/').filter(Boolean).map(encodeURIComponent).join('/')
}

export function useRetro() {
  /** A directory in the archive: what it holds. */
  function meta(path: string) {
    return $fetch<DriveMeta>(`${base()}/retro/meta/${encodePath(path)}/`, { credentials: 'include' })
  }

  /**
   * The scan itself.
   *
   * The media route rather than a streaming one: a reader opening a large
   * document asks for several pieces of it at once, and only the storage can
   * answer that.
   */
  function issue(path: string): string {
    return `${base()}/retro/media/${encodePath(path)}`
  }

  /** A cover, at the width a tile shows it. */
  function cover(path: string, width = 400): string {
    return `${base()}/retro/thumb/${encodePath(path)}?width=${width}`
  }

  /**
   * The archive: every magazine on the shelf, wherever it sits.
   *
   * Walked rather than listed one level deep, because the shelf is not arranged
   * the way a tidy one would be: Power Play lives *inside* the Happy Computer
   * folder, forty-three issues of it, and reading only the top level left them
   * invisible. A magazine is any folder that holds scans; where it sits says
   * nothing about whether it is one.
   */
  async function shelf(): Promise<Magazine[]> {
    return walk('')
  }

  async function walk(dir: string): Promise<Magazine[]> {
    const listing = await meta(dir)
    const members = listing.members ?? []

    const issues = members
      .filter(m => /\.pdf$/i.test(driveName(m)))
      .map((m) => {
        const name = driveName(m)
        const path = dir ? `${dir}/${name}` : name
        const title = issueTitle(dir, name)
        const { year, month, label } = dated(title)

        return { name, path, title, year, month, label, size: m.size, cover: cover(path), href: issue(path) }
      })

    const below = await Promise.all(
      members
        .filter(m => m.category === 'directory')
        .map(async (m) => {
          const under = dir ? `${dir}/${driveName(m)}` : driveName(m)
          try {
            return await walk(under)
          }
          catch (err) {
            // One unreadable folder should cost what is in it, not the shelf.
            // Without this a single refusal anywhere in the tree would leave a
            // visitor looking at an error instead of the hundred and thirty-six
            // issues that read perfectly well.
            console.warn('skipping a folder of the archive:', under, err)

            return []
          }
        }),
    )

    // A folder holding scans is a magazine; one holding only folders is just a
    // shelf. The root usually is the latter, and has no name worth showing.
    const here = issues.length
      ? [{ dir, title: magazineTitle(dir) || 'Archiv', groups: groupIssues(issues) }]
      : []

    return [...here, ...below.flat()]
  }

  return { shelf, meta, issue, cover, encodePath }
}
