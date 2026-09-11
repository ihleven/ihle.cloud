<template>
  <section class="p-4">
    <p class="text-gray-500">
      <NuxtLink :to="entrylink(entry.path)" class="group max-w-md">
        {{ entry.content.key }}
      </NuxtLink>
    </p>
  </section>
</template>

<script setup lang="ts">
defineProps<{
  entry: Entry
}>()
</script>
