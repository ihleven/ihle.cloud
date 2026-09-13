<template>
  <form class="ght-tipp-grid mx-auto w-min gap-2" @submit.prevent="submit">
    <div v-if="saved" class="text-center text-sm [grid-area:title]">
      Mein Tipp: {{ saved }}
      <p class="text-gray-500">(gespeichert am {{ savedAt }})</p>
    </div>

    <button
      type="button"
      class="grid place-items-center rounded-tl-lg border border-sky-300 bg-white shadow-lg [grid-area:hu] active:shadow-none"
      @click="step('home', 1)"
    >
      <UIcon name="i-heroicons-chevron-up" class="h-8 w-8 text-sky-500" />
    </button>
    <button
      type="button"
      class="grid place-items-center rounded-tr-lg border border-sky-300 bg-white shadow-lg [grid-area:gu] active:shadow-none"
      @click="step('away', 1)"
    >
      <UIcon name="i-heroicons-chevron-up" class="h-8 w-8 text-sky-500" />
    </button>
    <button
      type="button"
      class="grid place-items-center rounded-bl-lg border border-sky-300 bg-white shadow-lg [grid-area:hd] active:shadow-none"
      @click="step('home', -1)"
    >
      <UIcon name="i-heroicons-chevron-down" class="h-8 w-8 text-sky-500" />
    </button>
    <button
      type="button"
      class="grid place-items-center rounded-br-lg border border-sky-300 bg-white shadow-lg [grid-area:gd] active:shadow-none"
      @click="step('away', -1)"
    >
      <UIcon name="i-heroicons-chevron-down" class="h-8 w-8 text-sky-500" />
    </button>

    <div class="col-span-2 col-start-3 row-span-2 row-start-3 grid place-items-stretch shadow-lg active:shadow-none">
      <input
        :value="ergebnis" readonly
        class="w-full rounded-md border-2 border-gray-800 bg-white px-3 py-2 text-center text-lg font-bold focus:outline-none"
      >
    </div>

    <div class="grid place-items-center [grid-area:submit]">
      <button
        type="submit"
        :disabled="!ergebnis || state === 'loading'"
        class="flex h-12 w-40 items-center justify-center gap-2 rounded border border-transparent bg-sky-300 text-gray-100 hover:bg-sky-400 hover:text-white disabled:opacity-50 disabled:hover:bg-sky-300"
      >
        Tipp speichern
        <UIcon v-if="state === 'loading'" name="i-heroicons-arrow-path" class="h-6 w-6 animate-spin" />
        <UIcon v-else-if="state === 'success'" name="i-heroicons-check" class="h-6 w-6" />
        <UIcon v-else-if="state === 'error'" name="i-heroicons-exclamation-triangle" class="h-6 w-6" />
        <UIcon v-else name="i-heroicons-cloud-arrow-up" class="h-6 w-6" />
      </button>

      <p v-if="refusal" class="m-4 rounded-lg border-2 border-red-500 bg-red-100 p-4 text-red-600">{{ refusal }}</p>
    </div>
  </form>
</template>

<script setup lang="ts">
import type { Tipp } from '../types'
import type { SaveState } from '../composables/useGhtTipps'

const props = defineProps<{
  /** The tip already on record, if there is one. */
  tipp?: Tipp | null
  state: SaveState
  /** What the backend said if it would not take the tip. */
  refusal?: string
}>()

const emit = defineEmits<{ save: [ergebnis: string] }>()

// -1 rather than 0 for "nothing chosen yet", because 0:0 is a tip someone might
// actually mean. The first press of any arrow sets 0:0 and goes from there.
const home = ref(props.tipp?.Result?.[0] ?? -1)
const away = ref(props.tipp?.Result?.[1] ?? -1)

watch(() => props.tipp, (t) => {
  home.value = t?.Result?.[0] ?? -1
  away.value = t?.Result?.[1] ?? -1
})

const ergebnis = computed(() => home.value >= 0 && away.value >= 0 ? `${home.value}:${away.value}` : '')
const saved = computed(() => props.tipp?.Result ? `${props.tipp.Result[0]}:${props.tipp.Result[1]}` : '')
const savedAt = computed(() => {
  const t = props.tipp?.placed
  return t ? `${ghtDate(t)}, ${ghtTime(t)}` : ''
})

function step(side: 'home' | 'away', by: number) {
  const side_ = side === 'home' ? home : away
  if (home.value < 0 || away.value < 0) {
    home.value = 0
    away.value = 0
    return
  }
  side_.value = Math.max(0, side_.value + by)
}

function submit() {
  if (ergebnis.value) emit('save', ergebnis.value)
}
</script>
