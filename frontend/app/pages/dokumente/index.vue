<template>
  <div class="container mx-auto px-4 py-8">
    <MDC v-if="doc" :value="doc.content.markdown" class="prose prose-slate mt-8 max-w-none prose-headings:no-underline prose-a:font-semibold prose-a:underline prose-a:decoration-sky-300 prose-a:underline-offset-4 hover:prose-a:decoration-2 prose-img:rounded-xl" />
    <div v-else>
      Home not found
    </div>
  </div>
</template>

<script setup lang="ts">
const doc = await $fetch(`http://localhost:8000/foo`, {
  server: false,
  credentials: 'include',
  headers: { Accept: 'application/json' },
})
</script>
