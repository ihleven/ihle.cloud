<template>
  <div class="ght-tipp-rows mx-auto max-w-screen-md divide-y divide-neutral-300 border-b border-gray-300 bg-white">
    <template v-for="tipp in sorted" :key="tipp.login">
      <div class="flex items-center">
        <img class="h-16 w-16 rounded" :src="ghtAvatar(tipper[tipp.login]?.avatar, 64)" :alt="tipp.login">
      </div>

      <div class="flex flex-col justify-center overflow-hidden text-lg font-medium text-ellipsis whitespace-nowrap">
        {{ tipper[tipp.login]?.name || tipp.login }}
        <span class="text-sm text-gray-400">@{{ tipp.login }}</span>
      </div>

      <div class="flex items-stretch justify-start p-px">
        <span class="ght-chip">{{ tipp.Result ? `${tipp.Result[0]}:${tipp.Result[1]}` : '-:-' }}</span>
      </div>

      <div
        class="ght-outline flex flex-col items-center justify-center bg-[url('/assets/geheimtipp/trikots/trikot-gelb.svg')] bg-contain bg-center bg-no-repeat"
      >
        <strong class="text-3xl leading-8 font-extrabold">{{ tipp.Punkte }}</strong>
        <small class="text-xs leading-3">Pkt.</small>
      </div>

      <div />
    </template>
  </div>
</template>

<script setup lang="ts">
import type { EditionTipper, Tipp } from '../types'

const props = defineProps<{
  tipps: Record<string, Tipp> | null
  tipper: Record<string, EditionTipper>
}>()

// Best tip first. This is the one place the pool ranks by the tip itself rather
// than by standing, because the question the page answers is who got it right.
const sorted = computed(() => Object.values(props.tipps ?? {}).sort((a, b) => (b.Punkte ?? 0) - (a.Punkte ?? 0)))
</script>
