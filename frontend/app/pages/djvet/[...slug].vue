<template>
  <main class="bg-gray-50">
    <header class="sticky top-0 border-b border-gray-300 bg-gray-100">
      <Breadcrumbs :path="route.path" />
    </header>

    <section class="border-b border-gray-300" />

    <section
      class="mx-auto flex max-w-screen-md flex-col items-stretch border-x border-white bg-white shadow-lg md:flex-row"
    >
      <ul role="list" class="w-full divide-y divide-gray-300 border-r border-gray-300 md:w-2/4">
        <li v-for="f in tracks" :key="f.id" class="p-1">
          <nuxt-link :to="`/hidrive${f.path}`" class="flex">
            <div class="mr-4 flex h-16 w-16 flex-shrink-0 items-center justify-center self-center" />
            <div>
              <h4 class="text-md font-bold">{{ decodeURI(f.name) }}</h4>
              <p class="mt-1 text-xs font-medium text-gray-400">
                {{ f }}
              </p>
            </div>
          </nuxt-link>
        </li>
      </ul>

      <div class="w-full md:w-1/2">
        <div class="h-full md:w-[50vw]">
          <img
            :src="media(`public/djvet/${albumpath}/cover.jpeg`)"
            class="aspect-square w-full object-cover object-center"
          >
          <img
            :src="media(`public/djvet/${albumpath}/tracks.png`)"
            class="aspect-square w-full object-cover object-center"
          >
        </div>
      </div>
    </section>
  </main>
</template>

<script setup>
const route = useRoute()

const { bytes } = useHelpers()
const { media } = useHidrive()
const albumpath = typeof route.params.slug === 'string' ? route.params.slug : route.params.slug.join('/')

const { data: meta } = await useFetch(`http://localhost:8000/hi/meta/public/djvet/${albumpath}`, {
  headers: { Accept: 'application/json' },
  credentials: 'include',
  server: false,
})

const tracks = computed(() => meta.value?.members.filter(m => m.name.endsWith('.mp3')))

meta.name = decodeURI(meta.name)
console.log('albumpath', meta.value)
const { data: metadata } = await useFetch(`http://localhost:8000/hi/tags/public/djvet/${albumpath}`, {
  headers: { Accept: 'application/json' },
  credentials: 'include',
  server: false,
})
</script>
