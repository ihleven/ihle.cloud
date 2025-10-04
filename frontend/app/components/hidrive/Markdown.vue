<template>
  <!-- <MarkdownDoc :doc="doc" /> -->
  <article class="bg-white">
    <!-- <main class="container-fluid-md font-raleway">
      <ContentRenderer
        :value="doc"
        class="prose max-w-none prose-p:text-justify"
      />
    </main> -->
    <!-- <MDCRenderer
      :body="ast.body"
      :data="ast.data"
    /> -->
    <MDC

      :value="text"
      tag="article"
    />
    {{ text }}
  </article>
</template>

<script setup lang="ts">
// import markdownParser from '@nuxt/content/transformers/markdown'
// import { provide } from 'vue'

const props = defineProps({ path: String })

// const { data: ast } = await useAsyncData('markdown', () => parseMarkdown('::alert\nMissing markdown input\n::'))

// function dirname(input) {
//   const i = input.lastIndexOf('/')
//   return input.slice(0, i)
// }

const { data } = await useFetch('http://localhost:8000/hi/media/' + props.path, {
  server: false,
  credentials: 'include',
})
const text = await data.value.text()
console.log('text:', text)

// const doc = await markdownParser.parse(`<some-id>`, text)
// provide('basePath', `/api/raw`)
// provide('dirPath', dirname(props.path))
</script>
