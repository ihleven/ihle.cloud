<template>
  <div>
    <h1 class="bg-cyan-100 px-2 py-1 text-sm text-neutral-700">Jahr: {{ route.params.jahr }}</h1>
    <NuxtPage />
  </div>
</template>

<script setup lang="ts">
// A year. A placeholder from the unfinished migration, kept so the route it
// marks stays reachable.
const route = useRoute()
</script>
