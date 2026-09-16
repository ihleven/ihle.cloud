// The archive's filenames, read as they actually are.
//
// These names are thirty years of somebody filing scans: the same magazine
// spelled three ways, issues that carry a date and specials that carry nothing,
// and one magazine filed inside another. Every rule that reads them is a guess
// about a pattern, and a guess is worth pinning to the names it was made from —
// the tiles said "h 2cpc" and ".1984.03 Cartman" until one of these caught it.

import { describe, expect, it } from 'vitest'

import { dated, groupIssues, issueTitle, lastSegment, magazineTitle } from '../app/utils/retro'
import type { Issue } from '../app/utils/retro'

describe('issueTitle', () => {
  it('drops the magazine the folder has already named', () => {
    expect(issueTitle('HappyComputer', 'Happy.Computer.N05.1984.03-Cartman.pdf')).toBe('N05.1984.03 Cartman')
    expect(issueTitle('HappyComputer', 'happy-computer-sh-2cpc.pdf')).toBe('sh 2cpc')
    expect(issueTitle('HappyComputer', 'happy-computer-sonderheft1-sincliar.pdf')).toBe('sonderheft1 sincliar')
  })

  it('reads the folder it is in, not the path above it', () => {
    expect(issueTitle('HappyComputer/Powerplay', 'powerplay-1988-01.pdf')).toBe('1988 01')
    expect(issueTitle('HappyComputer/Powerplay', 'Power.Play.N007.1988.10-DURiAN_400dpi.pdf'))
      .toBe('N007.1988.10 DURiAN 400dpi')
  })

  it('keeps a name that does not start with the magazine', () => {
    // The one 1983 issue that is called Hobby Computer rather than Happy.
    expect(issueTitle('HappyComputer', 'Hobby.Computer.N01.1983.11-AndreBetz.pdf'))
      .toBe('Hobby.Computer.N01.1983.11 AndreBetz')
  })
})

describe('dated', () => {
  it('finds the year and month a regular issue carries', () => {
    expect(dated('N05.1984.03 Cartman')).toEqual({ year: '1984', month: 3, label: 'März' })
    expect(dated('N02.1983.12 AndreBetz')).toEqual({ year: '1983', month: 12, label: 'Dezember' })
    expect(dated('1988 01')).toEqual({ year: '1988', month: 1, label: 'Januar' })
  })

  it('leaves a special edition alone', () => {
    expect(dated('sh 25 dfü')).toEqual({ label: 'sh 25 dfü' })
    expect(dated('sonderheft 03 86')).toEqual({ label: 'sonderheft 03 86' })
  })
})

describe('groupIssues', () => {
  const issue = (name: string, year?: string, month?: number): Issue =>
    ({ name, title: name, year, month, label: name, cover: '', href: '' })

  it('puts a year in order even when the filenames do not', () => {
    const groups = groupIssues([
      issue('Happy.Computer.N02.1983.12', '1983', 12),
      issue('Hobby.Computer.N01.1983.11', '1983', 11),
    ])

    expect(groups).toHaveLength(1)
    expect(groups[0]!.title).toBe('1983')
    expect(groups[0]!.issues.map(i => i.month)).toEqual([11, 12])
  })

  it('sorts the years and keeps the specials last', () => {
    const groups = groupIssues([
      issue('a', '1986', 1),
      issue('sonderheft'),
      issue('b', '1984', 1),
    ])

    expect(groups.map(g => g.title)).toEqual(['1984', '1986', 'Sonderhefte'])
  })

  it('has no group at all when nothing is dated', () => {
    expect(groupIssues([issue('sh 9')]).map(g => g.title)).toEqual(['Sonderhefte'])
  })
})

describe('magazineTitle', () => {
  it('separates a run-together name', () => {
    expect(magazineTitle('HappyComputer')).toBe('Happy Computer')
  })

  it('names the folder, not the path to it', () => {
    expect(magazineTitle('HappyComputer/Powerplay')).toBe('Powerplay')
  })

  it('leaves a name that is genuinely one word', () => {
    expect(magazineTitle('Powerplay')).toBe('Powerplay')
    expect(lastSegment('a/b/c')).toBe('c')
  })
})
