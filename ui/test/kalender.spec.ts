import { describe, expect, it } from 'vitest'
import { dayPath, isoWeek, monthOf, monthPath, weekPath, yearPath } from '../app/utils/kalender'

describe('monthOf', () => {
  // The legacy page linked getMonth() straight into the route, so every month
  // link pointed at the month before it — and in January at month 0.
  it('is the month somebody would write down, not the index', () => {
    expect(monthOf(new Date(2026, 0, 15))).toBe(1)
    expect(monthOf(new Date(2026, 8, 16))).toBe(9)
    expect(monthOf(new Date(2026, 11, 31))).toBe(12)
  })
})

describe('isoWeek', () => {
  it('counts from the week holding the first Thursday', () => {
    expect(isoWeek(new Date(2026, 0, 1))).toEqual({ year: 2026, week: 1 })
    expect(isoWeek(new Date(2026, 8, 16))).toEqual({ year: 2026, week: 38 })
  })

  // The days either side of New Year are the whole reason the year is returned
  // alongside the week.
  it('gives January days to the previous year where ISO does', () => {
    // 1 Jan 2027 is a Friday, so it belongs to week 53 of 2026.
    expect(isoWeek(new Date(2027, 0, 1))).toEqual({ year: 2026, week: 53 })
  })

  it('gives late December days to the next year where ISO does', () => {
    // 31 Dec 2024 is a Tuesday, in the week whose Thursday is 2 Jan 2025.
    expect(isoWeek(new Date(2024, 11, 31))).toEqual({ year: 2025, week: 1 })
  })

  it('knows a year with fifty-three weeks', () => {
    // 2020 began on a Wednesday and was a leap year, so it has 53.
    expect(isoWeek(new Date(2020, 11, 31))).toEqual({ year: 2020, week: 53 })
  })

  it('handles a Sunday, which is day 7 and not day 0', () => {
    // 20 Sep 2026 is a Sunday, the last day of week 38.
    expect(isoWeek(new Date(2026, 8, 20))).toEqual({ year: 2026, week: 38 })
    // The Monday after starts week 39.
    expect(isoWeek(new Date(2026, 8, 21))).toEqual({ year: 2026, week: 39 })
  })
})

describe('paths', () => {
  it('addresses a year, a month and a day', () => {
    expect(yearPath(2026)).toBe('/kalender/2026')
    expect(monthPath(2026, 9)).toBe('/kalender/2026/9')
    expect(dayPath(2026, 9, 18)).toBe('/kalender/2026/9/18')
  })

  // The route file is kw[kw].vue, so the link has to be lower-case to match it.
  it('spells the week route the way the route is spelled', () => {
    expect(weekPath(2026, 38)).toBe('/kalender/2026/kw38')
  })

  // The picker hands back months counted from one, the same as monthOf, so the
  // two agree and neither needs adjusting at the call site.
  it('agrees with the month the date block shows', () => {
    const date = new Date(2026, 8, 18)
    expect(monthPath(date.getFullYear(), monthOf(date))).toBe('/kalender/2026/9')
  })
})
