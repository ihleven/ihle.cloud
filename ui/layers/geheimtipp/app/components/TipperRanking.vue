<template>
  <table class="w-full table-auto border-collapse md:w-fit">
    <caption class="py-1 text-sm whitespace-nowrap">
      <button
        class="px-1" :class="gesamt ? 'text-gray-500' : 'font-semibold text-gray-900'"
        @click="gesamt = false"
      >Tageswertung</button>
      /
      <button
        class="px-1" :class="gesamt ? 'font-semibold text-gray-900' : 'text-gray-500'"
        @click="gesamt = true"
      >Gesamtwertung</button>
    </caption>

    <thead class="text-xs">
      <tr>
        <td class="text-center">Platz</td>
        <td class="text-center">Tipper</td>
        <td class="text-center">Tipps</td>
        <td class="text-center">Punkte</td>
      </tr>
    </thead>

    <tbody class="divide-y divide-gray-300 border-y border-gray-300 bg-white text-sm sm:text-base">
      <tr v-for="ght in ranked" :key="ght" class="hover:bg-sky-200">
        <td class="px-1 text-right">{{ round.evals[ght]?.trikots?.GELB?.platz }}.</td>

        <td>
          <div class="flex items-center">
            <img
              v-if="tipper[ght]?.avatar"
              class="mx-1 my-0.5 h-7 w-7 rounded"
              :src="ghtAvatar(tipper[ght]?.avatar, 32)"
              :alt="ght"
            >
            <small class="text-lg font-medium">{{ tipper[ght]?.name || ght }}</small>
          </div>
        </td>

        <td class="text-center"><small>{{ round.evals[ght]?.ntipps }}</small></td>

        <!-- Both numbers are always shown; the toggle decides which is the
             headline. Reading one against the other is the point of the table. -->
        <td class="text-center">
          <span :class="gesamt ? 'text-sm font-light' : 'font-semibold'">{{ round.evals[ght]?.spieltag?.punkte }}</span>
          /
          <span :class="gesamt ? 'font-bold' : 'text-sm font-light'">{{ round.evals[ght]?.trikots?.GELB?.punkte }}</span>
        </td>
      </tr>
    </tbody>
  </table>
</template>

<script setup lang="ts">
import type { CompRound, EditionTipper } from '../types'

const props = defineProps<{
  round: CompRound
  tipper: Record<string, EditionTipper>
}>()

const gesamt = ref(true)

const ranked = computed(() => {
  const logins = Object.keys(props.round.evals)
  if (gesamt.value) {
    const place = (t: string) => props.round.evals[t]?.trikots?.GELB?.platz ?? Number.MAX_SAFE_INTEGER
    return logins.sort((a, b) => place(a) - place(b))
  }
  const points = (t: string) => props.round.evals[t]?.spieltag?.punkte ?? -1
  return logins.sort((a, b) => points(b) - points(a))
})
</script>
