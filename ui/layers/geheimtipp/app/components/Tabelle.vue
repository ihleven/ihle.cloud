<template>
  <table class="w-full table-auto border-collapse md:w-fit">
    <caption class="py-1 text-sm text-gray-600">Tabelle</caption>
    <thead class="text-xs">
      <tr>
        <td colspan="2" class="px-1 text-center">Platz</td>
        <td class="px-1 text-center">Verein</td>
        <td class="px-1 text-center">Spiele</td>
        <td class="px-1 text-center">S-U-N</td>
        <td colspan="2" class="px-1 text-center">Differenz</td>
        <td class="px-1 text-center">Punkte</td>
      </tr>
    </thead>
    <tbody class="divide-y divide-gray-300 border-y border-gray-300 bg-white text-sm sm:text-base">
      <tr v-for="t in round.tabelle" :key="t.team" class="group">
        <td class="text-right">{{ t.POS }}.</td>

        <!-- Movement against the previous matchday. PD, not a comparison with
             PrevPos: the backend has already worked out what counts as a move. -->
        <td class="text-center text-gray-500">
          <UIcon v-if="t.PD > 0" name="i-heroicons-chevron-up" class="h-4 w-4" />
          <UIcon v-else-if="t.PD < 0" name="i-heroicons-chevron-down" class="h-4 w-4" />
          <span v-else>-</span>
        </td>

        <td class="font-semibold">
          <div class="flex items-center">
            <GhtCrest :team="t.team" size="6" class="mx-2 my-1" />
            {{ round.teams[t.team]?.name || t.team }}
          </div>
        </td>

        <td class="px-1 text-right">{{ t.PLD }}</td>
        <td class="px-2 text-center whitespace-nowrap">
          <small>{{ t.W }}-{{ t.D }}-{{ t.L }}</small>
        </td>
        <td class="px-1 text-center font-medium">
          <small>{{ t.GF }}:{{ t.GA }}</small>
        </td>

        <!-- Goal quotient and two-point counting are how older seasons were
             reckoned; the season says which, so a historic table is not
             silently restated in today's rules. -->
        <td v-if="round.compSeason.quot" class="text-right">{{ t.GQ }}</td>
        <td v-else class="text-right">
          <span v-if="t.GD === 0">&plusmn;0</span>
          <span v-else-if="t.GD > 0">&plus;{{ t.GD }}</span>
          <span v-else>{{ t.GD }}</span>
        </td>

        <td v-if="round.compSeason.pkt2" class="pr-2 pl-1 text-right font-bold">{{ t.PP }}:{{ t.MP }}</td>
        <td v-else class="pr-2 pl-1 text-right font-bold">{{ t.PTS }}</td>
      </tr>
    </tbody>
  </table>
</template>

<script setup lang="ts">
import type { CompRound } from '../types'

defineProps<{ round: CompRound }>()
</script>
