<template>
  <div>
    <h1 class="bg-red-100 px-2 py-1 text-sm text-neutral-700">Monat: {{ route.params.monat }}</h1>
    <NuxtPage />
  </div>
</template>

<script setup lang="ts">
// A month, under its year. A placeholder, as the year above it is.
const route = useRoute()
</script>
