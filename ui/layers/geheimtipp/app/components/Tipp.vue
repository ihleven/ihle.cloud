<template>
  <div class="relative mx-auto w-min">
    <input
      v-if="editable"
      :value="draft"
      class="w-10 rounded border-0 bg-sky-100 px-1 py-0.5 text-center text-sky-800 ring-1 ring-sky-300 focus:ring focus:ring-sky-500 focus:outline-none"
      @input="onInput"
    >
    <div v-else class="px-1 text-center" :class="classes">{{ result }}</div>

    <!-- The points badge, half outside the corner. Hidden at zero: a wrong tip
         is already shown by being grey, and a column of noughts would read as
         scores. -->
    <small
      v-if="tipp?.Punkte"
      class="absolute top-0 -right-3 flex h-4 w-4 -translate-y-1/2 items-center justify-center rounded-full border border-black bg-white text-xs text-black ring-2 ring-white"
      :class="{
        'border-blue-700 font-bold text-blue-700': tipp.Punkte === 2,
        'border-blue-500 font-bold text-blue-500': tipp.Punkte === 1,
      }"
    >
      {{ tipp.Punkte }}
    </small>
  </div>
</template>

<script setup lang="ts">
import type { Tipp } from '../types'

const props = defineProps<{
  /** Absent when this tipper did not tip this match. */
  tipp?: Tipp | null
  /** Kick-off: what decides whether a missing tip is still to come or was missed. */
  deadline: string
  /** Whether this cell belongs to the signed-in tipper and is still open. */
  editable?: boolean
}>()

const emit = defineEmits<{ 'update:tipp': [ergebnis: string] }>()

const result = computed(() => {
  const r = props.tipp?.Result
  if (r) return `${r[0]}:${r[1]}`
  // Before kick-off a blank cell means "not yet"; after it, the pool writes the
  // miss out, because a blank there would read as a rendering fault.
  return new Date(props.deadline) > new Date() ? '' : '-:-'
})

// What is in the field, which is not what is on record until it is sent. Seeded
// from the saved tip and re-seeded when that changes, so a save elsewhere on the
// page does not leave a stale draft sitting in the box.
const draft = ref(result.value)
watch(result, r => draft.value = r)

function onInput(event: Event) {
  draft.value = (event.target as HTMLInputElement).value
  emit('update:tipp', draft.value)
}

// Four states, and the styling is the whole of how the matrix is read at a
// glance: exact tips carry a doubled rule, tendencies are plain blue, misses
// recede, and a cell with no tip at all is fainter still.
const classes = computed(() => {
  if (!props.tipp) return 'text-slate-300'
  switch (props.tipp.Punkte) {
    case 2: return 'border-r-4 border-double border-blue-700 font-bold text-blue-700'
    case 1: return 'text-blue-600'
    case 0: return 'font-light text-slate-400'
    default: return 'text-blue-500'
  }
})
</script>
