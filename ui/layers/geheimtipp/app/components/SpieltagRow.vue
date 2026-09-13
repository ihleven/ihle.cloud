<template>
  <NuxtLink
    :to="`/geheimtipp/spieltag-${season}-${round.round}`"
    class="flex items-center justify-between border-b border-gray-300 bg-white py-2 pl-2 hover:bg-sky-50"
  >
    <span>{{ label }} {{ round.round }}</span>
    <small class="ml-4 text-sm text-gray-500">{{ ghtDateRange(round.from, round.to) }}</small>
    <span class="ml-auto flex items-center pr-1 text-gray-400">
      {{ round.round }}
      <UIcon name="i-heroicons-chevron-right" class="h-6 w-6" />
    </span>
  </NuxtLink>
</template>

<script setup lang="ts">
import type { RoundWindow } from '../types'

defineProps<{
  label: string
  round: RoundWindow
  /** The edition code. The pool's own frontend hardcodes it and edits it each
   *  August; here it comes from /aktuell. */
  season: string
}>()
</script>
