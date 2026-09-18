<template>
  <main class="min-h-[60vh] bg-cyan-100 px-2 py-1">
    <h1 class="text-sm text-neutral-700">Tag: {{ route.params.tag }}</h1>
  </main>
</template>

<script setup lang="ts">
// A day, the deepest the unfinished migration reached.
const route = useRoute()
</script>
