<template>
  <div class="container mx-auto px-4 py-8 max-w-screen-md">
    <ContentRenderer v-if="doc" :value="doc" class="prose prose-slate mt-8 max-w-none prose-headings:no-underline prose-a:font-semibold prose-a:underline prose-a:decoration-sky-300 prose-a:underline-offset-4 hover:prose-a:decoration-2 prose-img:rounded-xl" />
    <div v-else>
      {{ route.path.replace('/dokumente', '') }} not found
    </div>
  </div>
</template>

<script setup lang="ts">
const route = useRoute()
console.log('route', route.path)
const { data: doc } = await useAsyncData(route.path, () => {
  return queryCollection('content').path(route.path.replace('/dokumente', '')).first()
})

// const { data: home } = await useAsyncData(() => queryCollection('content').first())

useSeoMeta({
  title: doc.value?.title,
  description: doc.value?.description,
})
</script>
