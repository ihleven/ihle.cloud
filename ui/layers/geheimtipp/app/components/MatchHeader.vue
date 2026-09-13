<template>
  <div>
    <nav class="flex h-12 justify-between border-b border-dashed bg-white">
      <NuxtLink
        class="flex items-center pr-2 hover:text-sky-600"
        :to="`/geheimtipp/spieltag-${match.round.season}-${match.round.num}`"
      >
        <UIcon name="i-heroicons-chevron-left" class="m-2 h-8 w-8" />
        Spieltag {{ match.round.code }}
      </NuxtLink>
    </nav>

    <section class="ght-splash gap-4">
      <NuxtLink
        class="grid h-full w-16 place-items-center [grid-area:previous]"
        :class="{ 'pointer-events-none opacity-40': !previous }"
        :to="previous ? `/geheimtipp/matches/${previous}` : ''"
      >
        <UIcon name="i-heroicons-chevron-left" class="h-10 w-10 rounded-full bg-black/40 p-1 text-white hover:bg-black/20" />
      </NuxtLink>

      <div class="text-center [grid-area:match]">
        <h1 class="ght-outline pt-2 text-2xl">Spiel {{ match.no }}</h1>
      </div>

      <NuxtLink
        class="grid h-full w-16 place-items-center [grid-area:next]"
        :class="{ 'pointer-events-none opacity-40': !next }"
        :to="next ? `/geheimtipp/matches/${next}` : ''"
      >
        <UIcon name="i-heroicons-chevron-right" class="h-10 w-10 rounded-full bg-black/25 p-1 text-white hover:bg-black/20" />
      </NuxtLink>

      <div class="text-center [grid-area:intro]">
        <strong class="ght-outline">{{ ghtDate(match.kickoff) }}, {{ ghtTime(match.kickoff) }}</strong>
      </div>

      <div class="flex flex-col items-end text-xl [grid-area:hteam]">
        <GhtCrest :team="match.teams[0]" size="20" />
        <h1 class="ght-outline -translate-y-1/2 bg-black/25 p-1 text-xl leading-none">{{ match.teams[0] }}</h1>
      </div>

      <div class="flex items-end justify-center text-xl [grid-area:sep]">
        <h2 class="ght-outline text-2xl">{{ match.result || '-:-' }}</h2>
      </div>

      <div class="flex flex-col items-start text-xl [grid-area:ateam]">
        <GhtCrest :team="match.teams[1]" size="20" />
        <h1 class="ght-outline -translate-y-1/2 bg-black/20 p-1 text-xl leading-none">{{ match.teams[1] }}</h1>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import type { MatchDetail } from '../types'

const props = defineProps<{ match: MatchDetail }>()

// A match id is season, round and number run together — 20270301 — so stepping
// within a round is arithmetic on the last two digits. The neighbours are not in
// this payload, and the round says how many there are.
const number = computed(() => props.match.id % 100)
const previous = computed(() => number.value > 1 ? props.match.id - 1 : null)
const next = computed(() => number.value < props.match.round.numMatches ? props.match.id + 1 : null)
</script>
