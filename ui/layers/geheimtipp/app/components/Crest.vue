<template>
  <img
    v-if="!missing"
    :src="ghtCrest(team)"
    :alt="team"
    :class="box"
    class="max-w-none object-contain"
    @error="missing = true"
  >
  <!-- A club with no crest on file — a promotion, usually. The code is what the
       narrow layout shows anyway, so the row stays readable. -->
  <span v-else :class="box" class="block text-center text-xs font-semibold text-gray-400">{{ team }}</span>
</template>

<script setup lang="ts">
const props = withDefaults(defineProps<{
  team: string
  /** 8 in the matrix, 6 in the league table, 20 on a match splash. */
  size?: '6' | '8' | '20'
}>(), { size: '8' })

// Written out rather than interpolated: Tailwind reads the source for class
// names, and `h-${size}` is not a name it can find.
//
// max-w-none goes with them: the preflight rule `img { max-width: 100% }` would
// otherwise resolve against a table cell that has no width of its own yet, and
// collapse the crest to nothing. The pool's own component inlines an <svg>,
// which that rule does not match, so it never had to say this.
const box = computed(() => ({ 6: 'h-6 w-6', 8: 'h-8 w-8', 20: 'h-20 w-20' }[props.size]))

const missing = ref(false)
</script>
