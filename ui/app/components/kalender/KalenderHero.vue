<template>
  <div :class="s.root">
    <div :class="s.frame">
      <div>
        <!-- The month and the year stacked against the day, which is the shape
             the old page had: a narrow right-aligned column of the smaller
             parts, and the day itself large beside it. -->
        <div class="flex items-start text-outline text-white">
          <span class="flex flex-col items-end justify-between" :class="s.column">
            <NuxtLink :to="monthPath(year, month)" :class="s.month" class="font-medium hover:underline">
              {{ monthLabel }}
            </NuxtLink>
            <NuxtLink :to="weekPath(week.year, week.week)" :class="s.week" class="hover:underline">
              KW{{ week.week }}
            </NuxtLink>
            <NuxtLink :to="yearPath(year)" :class="s.year" class="font-black hover:underline">
              {{ year }}
            </NuxtLink>
          </span>

          <NuxtLink :to="dayPath(year, month, day)" :class="s.day" class="font-black hover:underline">
            {{ day }}
          </NuxtLink>
        </div>

        <div :class="[s.weekday, s.weekdayTone]" class="font-semibold">{{ weekdayLabel }}</div>
      </div>

      <!-- The picker the old page reached for and never got: it referenced
           v-calendar, which was not installed, so the right half of the band was
           empty. This is Nuxt UI's own, which is already here.

           On a card rather than bare on the green: the calendar is drawn in the
           app's surface colours and would otherwise sit on a background it knows
           nothing about. -->
      <UCalendar
        v-if="variant === 'page'"
        v-model="picked"
        v-model:placeholder="showing"
        :week-starts-on="1"
        class="rounded-md bg-default p-2 shadow-sm"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { CalendarDate } from '@internationalized/date'
import type { DateValue } from '@internationalized/date'

// Today, as the calendar writes it: the month, the week, the year, the day and
// the weekday, each of them a way into the page that shows it — and, on the
// page itself, a picker for reaching any other day.
//
// One component for two places. The front page shows the same block as the
// calendar itself, so the two cannot drift into saying today differently, and
// the only thing the tile changes is how large it is — hence a variant rather
// than a copy. The picker is the exception: a tile is two hundred pixels across
// and has no room for a month.
//
// `text-white` is stated next to `text-outline` rather than left to it: the
// utility is declared twice in main.css and only one of the two sets a colour,
// so relying on it would make the text's colour depend on which declaration
// wins.
const { variant = 'page' } = defineProps<{ variant?: 'page' | 'tile' }>()

const now = ref(new Date())

const day = computed(() => now.value.getDate())
const year = computed(() => now.value.getFullYear())
const month = computed(() => monthOf(now.value))
const monthLabel = computed(() => monthName(now.value))
const weekdayLabel = computed(() => weekdayName(now.value))
const week = computed(() => isoWeek(now.value))

// What the picker has selected, and which month it is showing. Two values
// because they move independently: paging to March without choosing a day
// changes the second and not the first.
const picked = shallowRef<DateValue>(new CalendarDate(year.value, month.value, day.value))
const showing = shallowRef<DateValue>(new CalendarDate(year.value, month.value, day.value))

// Choosing a day goes to that day. The old page did the same, and had to add one
// to the month on the way because it was reading a JavaScript Date; this value
// counts months from one already, which is what dayPath expects.
watch(picked, (value) => {
  if (value) navigateTo(dayPath(value.year, value.month, value.day))
})

// Paging to another month goes to that month — but only when the month has
// actually changed. Choosing a day moves this value too, and without the guard
// every click would navigate twice, the second one landing on the month and
// undoing the first.
watch(showing, (value, before) => {
  if (!value) return
  if (before && value.year === before.year && value.month === before.month) return

  navigateTo(monthPath(value.year, value.month))
})

// The tile is a square on the front page and has to hold the same five parts in
// a fraction of the width — half a phone's screen at its smallest. So every
// size is named here rather than scaled by a transform, which would blur the
// outline the letters are drawn with.
const sizes = {
  page: {
    root: '',
    // Beside the date where there is room, under it where there is not.
    // Hiding it on a phone would be the easy way to fit it and would leave the
    // one screen that most wants a picker without one.
    frame: 'flex flex-col items-start gap-4 sm:flex-row sm:justify-between',
    column: 'p-2',
    month: 'text-2xl',
    week: 'text-base',
    year: 'text-3xl',
    day: 'text-8xl leading-[6rem]',
    weekday: 'text-3xl',
    // Dark on the band's one flat green, as the old page had it.
    weekdayTone: '',
  },
  tile: {
    // Centred rather than stacked from the top: the tile is a square and the
    // date is five short lines, so anchoring it to the top leaves the bottom
    // half of every tile empty.
    root: 'flex h-full flex-col justify-center',
    frame: '',
    column: 'p-1',
    month: 'text-xs sm:text-sm',
    week: 'text-[10px] sm:text-xs',
    year: 'text-xl sm:text-2xl',
    day: 'text-6xl leading-none sm:text-7xl',
    weekday: 'text-sm sm:text-base',
    // Outlined here, unlike on the page: the tile's background runs the whole
    // green ramp, and dark text disappears into the foot of it.
    weekdayTone: 'text-outline text-white',
  },
} as const

const s = computed(() => sizes[variant])
</script>
