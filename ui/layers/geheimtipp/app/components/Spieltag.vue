<template>
  <table class="w-full table-auto border-collapse md:mx-auto md:w-min">
    <thead>
      <tr>
        <td class="hidden md:table-cell" />
        <td colspan="5" />
        <td
          v-for="t in columns" :key="t"
          class="w-10 text-center"
          :class="{ 'hidden md:table-cell': t !== login }"
        >
          <img
            v-if="tipper[t]?.avatar"
            class="mx-auto h-8 w-8 rounded"
            :src="ghtAvatar(tipper[t]?.avatar, 32)"
            :alt="t"
          >
          <small class="font-semibold text-gray-500">{{ t }}</small>
        </td>
        <td class="md:hidden" />
      </tr>
    </thead>

    <tbody class="divide-y divide-zinc-300 border-y border-zinc-400">
      <tr v-for="(m, i) in round.matches" :key="m.id" class="group bg-white hover:bg-sky-200">
        <td class="hidden px-2 text-sm text-gray-500 md:table-cell">
          {{ ghtDateIfChanged(m.kickoff, i > 0 ? round.matches[i - 1]!.kickoff : null) }}
        </td>

        <td class="px-2 text-right font-semibold whitespace-nowrap">
          <NuxtLink :to="`/geheimtipp/matches/${m.id}`">
            <span class="hidden md:inline">{{ round.teams[m.teams[0]]?.name || m.teams[0] }}</span>
            <span class="md:hidden">{{ m.teams[0] }}</span>
          </NuxtLink>
        </td>

        <td class="h-10 w-8">
          <GhtCrest :team="m.teams[0]" class="group-hover:grayscale-0 md:grayscale" />
        </td>

        <td class="w-16 px-1 text-center">
          <input
            v-if="results"
            v-model="resultDraft[m.id]"
            class="mx-3 my-1 w-10 rounded border bg-gray-200 px-1 py-0.5 text-center"
          >
          <!-- "-:-" is how the backend says "not played", and the kick-off time
               is more use there than a placeholder. -->
          <div v-else-if="m.result === '-:-'" class="px-1 py-0.5 text-xs font-light text-zinc-400">
            <small>{{ ghtDateIfChanged(m.kickoff) }}</small><br>{{ ghtTime(m.kickoff) }}
          </div>
          <div v-else class="m-1 rounded border border-transparent bg-white px-1 py-0.5 font-bold hover:border-black">{{ m.result }}</div>
        </td>

        <td class="w-8">
          <GhtCrest :team="m.teams[1]" class="grayscale group-hover:grayscale-0" />
        </td>

        <td class="px-2 font-semibold whitespace-nowrap">
          <NuxtLink :to="`/geheimtipp/matches/${m.id}`">
            <span class="hidden md:inline">{{ round.teams[m.teams[1]]?.name || m.teams[1] }}</span>
            <span class="md:hidden">{{ m.teams[1] }}</span>
          </NuxtLink>
        </td>

        <td
          v-for="t in columns" :key="t"
          :class="{ 'hidden md:table-cell': t !== login }"
        >
          <GhtTipp
            :tipp="m.odds?.tipps?.[t]"
            :deadline="m.kickoff"
            :editable="mayTip(login, t, m.kickoff)"
            @update:tipp="draft[m.id] = $event"
          />
        </td>

        <td class="pl-2">
          <NuxtLink
            class="flex items-center justify-end text-gray-400 group-hover:text-sky-600"
            :to="`/geheimtipp/matches/${m.id}`"
          >
            <UIcon name="i-heroicons-chevron-right" class="h-6 w-6" />
          </NuxtLink>
        </td>
      </tr>
    </tbody>

    <tfoot class="border-t border-gray-700">
      <tr v-if="open || results">
        <td class="hidden md:table-cell" />
        <td colspan="5" class="text-center">
          <button
            v-if="results"
            class="rounded bg-blue-500 px-2 py-1 text-sm text-white hover:bg-blue-600"
            @click="$emit('results', { ...resultDraft })"
          >
            Ergebnisse speichern
          </button>
        </td>
        <td
          v-for="t in columns" :key="t"
          :class="{ 'hidden md:table-cell': t !== login }"
        >
          <button
            v-if="t === login && open"
            class="rounded border border-sky-300 bg-white px-2 py-1 text-sm text-sky-400 hover:border-zinc-100 hover:bg-sky-300 hover:text-zinc-50 disabled:opacity-50"
            :disabled="!Object.keys(draft).length"
            @click="submit"
          >
            Abschicken
          </button>
        </td>
        <td />
      </tr>

      <tr class="border-t border-gray-300">
        <td class="hidden md:table-cell" />
        <td colspan="5" class="text-center">Tageswertung:</td>
        <td
          v-for="t in columns" :key="t"
          class="text-center"
          :class="{ 'hidden md:table-cell': t !== login }"
        >
          <div v-if="round.evals[t]?.spieltag?.platz" class="text-xs">
            {{ round.evals[t]?.spieltag?.platz }}.
          </div>
          <div
            class="ght-outline mx-auto flex h-10 w-12 flex-col items-center justify-center bg-[url('/assets/geheimtipp/trikots/trikot-gelb.svg')] bg-contain bg-center bg-no-repeat"
          >
            <strong class="mt-1 text-xl leading-none font-bold">{{ round.evals[t]?.spieltag?.punkte }}</strong>
            <!-- The <small> sits inside the sized div rather than carrying the
                 size itself, so its own shrink applies on top of it: "Pkt."
                 comes out a step below the label size. -->
            <div class="-mt-0.5 text-xs leading-none">
              <small>Pkt.</small>
            </div>
          </div>
        </td>
        <td />
      </tr>
    </tfoot>
  </table>
</template>

<script setup lang="ts">
import type { CompRound, EditionTipper } from '../types'

// Below md only the signed-in tipper's own column is shown, which is what makes
// the matrix usable on a phone — twelve columns of three-character scores do not
// fit, and shrinking them until they do makes none of them legible.
const props = defineProps<{
  round: CompRound
  tipper: Record<string, EditionTipper>
  /** The signed-in login, or "" — decides which column is theirs. */
  login: string
  /** Whether to offer the result fields. One person keeps them; see keepsResults. */
  results?: boolean
}>()

const emit = defineEmits<{
  tipps: [tipps: Record<string, string>]
  results: [results: Record<string, string>]
}>()

/**
 * The columns, in the pool's order: your own first, then by overall standing.
 *
 * The list comes from the matchday's evals rather than from the edition's
 * tippers: someone who registered but has not tipped this round has no column
 * to show, and the two lists are not the same.
 */
/** Whether anything on this matchday is still open to the signed-in tipper. */
const open = computed(() => props.round.matches.some(m => mayTip(props.login, props.login, m.kickoff)))

// What has been typed but not sent, keyed by match id. Cleared by the parent
// replacing the round after a save, which re-seeds every cell from the record.
const draft = reactive<Record<string, string>>({})

// The results are seeded from the record rather than left blank, because the
// form sends the whole matchday: an empty field would read as "clear this one".
const resultDraft = reactive<Record<string, string>>({})
watchEffect(() => {
  for (const m of props.round.matches) resultDraft[m.id] = m.result ?? ''
})

function submit() {
  // Only cells that are still open go up. A draft can outlive its kick-off if
  // the page has been sitting open, and sending it would earn a refusal for
  // something the person did type in time.
  const tipps: Record<string, string> = {}
  for (const m of props.round.matches) {
    if (draft[m.id] && mayTip(props.login, props.login, m.kickoff)) tipps[m.id] = draft[m.id]!
  }
  emit('tipps', tipps)
}

const columns = computed(() => {
  const logins = Object.keys(props.round.evals)
  const place = (t: string) => props.round.evals[t]?.trikots?.GELB?.platz ?? Number.MAX_SAFE_INTEGER
  return logins.sort((a, b) => {
    if (a === props.login) return -1
    if (b === props.login) return 1
    return place(a) - place(b)
  })
})
</script>
