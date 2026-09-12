<template>
  <article class="border border-accented">
    {{ entry }}
  </article>
</template>

<script setup lang="ts">
const props = defineProps<{
  entry: ExhibitionEntry
}>()
</script>
