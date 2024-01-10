<template>
  <main class="min-h-screen bg-white py-12">
    <button
      type="button"
      class="focus-visible:ring-sky text-sm font-medium ring-0 focus:outline-none focus-visible:ring-2 focus-visible:ring-opacity-75"
      @click="openModal"
    >
      Open dialog
    </button>
    <MarkdownDoc :doc="doc" />

    <Modal :open="isOpen" @close="closeModal">
      <textarea
        v-model="data"
        class="h-rull w-full rounded border-0 p-2 font-mono text-sm focus:outline-none focus-visible:ring focus-visible:ring-sky-300"
      ></textarea>
    </Modal>
  </main>
</template>

<script setup>
  import markdownParser from '@nuxt/content/transformers/markdown'

  import { provide } from 'vue'

  function dirname(input) {
    const i = input.lastIndexOf('/')
    return input.slice(0, i)
  }

  const route = useRoute()
  const path = typeof route.params.slug === 'string' ? route.params.slug : route.params.slug.join('/')
  const dir = dirname(path)

  const { data } = await useFetch(`/api/home/${path}`)
  console.log('meta:', data.value)

  const doc = await markdownParser.parse(`<some-id>`, data.value)
  // Extract the filename:
  provide('basePath', `/api/home`)
  provide('dirPath', dir)

  const isOpen = ref(false)

  function closeModal() {
    isOpen.value = false
  }
  function openModal() {
    isOpen.value = true
  }
</script>
