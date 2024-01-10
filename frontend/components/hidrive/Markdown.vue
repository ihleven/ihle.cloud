<template>
  <!-- <MarkdownDoc :doc="doc" /> -->
  <article class="bg-white">
    <!-- <header class="w-full bg-violet-200">
      <p class="mb-2 text-sm font-semibold leading-6 text-sky-500">{{ doc.title }}</p>
      <h1 class="inline-block text-2xl font-extrabold tracking-tight text-slate-900 sm:text-3xl">{{ doc?.title }}</h1>
      <p class="mt-2 text-lg text-slate-700 dark:text-slate-400">{{ doc?.description }}</p>
    </header> -->
    <main class="container-fluid-md font-raleway">
      <ContentRenderer :value="doc" class="prose max-w-none prose-p:text-justify" />
    </main>
  </article>
</template>

<script setup>
  import markdownParser from '@nuxt/content/transformers/markdown'
  import { provide } from 'vue'

  const props = defineProps({ path: String })

  function dirname(input) {
    const i = input.lastIndexOf('/')
    return input.slice(0, i)
  }

  const { data } = await useFetch(`/api/raw/${props.path}`)
  const text = await data.value.text()
  console.log('meta:', text)

  const doc = await markdownParser.parse(`<some-id>`, text)
  provide('basePath', `/api/raw`)
  provide('dirPath', dirname(props.path))
</script>
