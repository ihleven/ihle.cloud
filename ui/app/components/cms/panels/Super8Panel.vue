<template>
  <section class="p-4">
    <img :src="entry.content.thumbnail" class="bg-black text-gray-500">
    {{ entry.content }}
  </section>
</template>

<script setup lang="ts">
defineProps<{
  entry: Entry
}>()
</script>
