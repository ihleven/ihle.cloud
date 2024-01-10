<template>
  <main class="bg-gray-50">
    <header class="sticky top-0 border-b border-gray-300 bg-gray-100">
      <Breadcrumbs :path="route.path" />
    </header>

    <section class="border-b border-gray-300"></section>

    <section
      class="mx-auto flex max-w-screen-md flex-col items-stretch border-x border-white bg-white shadow-lg md:flex-row"
    >
      <ul role="list" class="w-full divide-y divide-gray-300 border-r border-gray-300 md:w-2/4">
        <li v-for="f in tracks" :key="f.id" class="p-1">
          <nuxt-link :to="`/hidrive${f.path}`" class="flex">
            <div class="mr-4 flex h-16 w-16 flex-shrink-0 items-center justify-center self-center">
              <img
                v-if="f.category == 'image'"
                :src="`/api/thumbs?path=${f.path}&width=100`"
                class="h-16 w-16 rounded object-cover object-center"
                aria-hidden="true"
              />
              <Micon v-else name="folder" class="h-8 w-8 stroke-1" filled />
            </div>
            <div>
              <h4 class="text-md font-bold">{{ decodeURI(f.name) }}</h4>
              <p class="mt-1 text-xs font-medium text-gray-400">
                <template v-if="f.type == 'dir'"
                  >{{ f.category }}, {{ f.nmembers }} files, {{ bytes(f.size) }}
                </template>
                <template v-if="f.category == 'image'">
                  {{ f.mime_type }}, {{ f.image.width }}x{{ f.image.height }}px, {{ bytes(f.size) }}
                </template>
                <template v-if="f.category == 'code'"> {{ f.category }}, {{ bytes(f.size) }} </template>

                <template v-else></template>
                <br />
                {{ metadata[decodeURI(f.name)] }} geändert: {{ new Date(f.mtime * 1000).toLocaleString() }}
                {{ f.readable ? 'R' : '' }}{{ f.writable ? 'W' : '' }} - {{ f.type }}, {{ f.category }}
              </p>
            </div>
          </nuxt-link>
        </li>
      </ul>

      <div class="w-full md:w-1/2">
        <div class="h-full md:w-[50vw]">
          <img
            :src="`/api/raw/public/djvet/${path}/cover.jpeg`"
            class="aspect-square w-full object-cover object-center"
          />
          <img
            :src="`/api/raw/public/djvet/${path}/tracks.png`"
            class="aspect-square w-full object-cover object-center"
          />
        </div>
      </div>
    </section>
  </main>
</template>

<script setup>
  const route = useRoute()
  console.log(typeof route.params.slug)
  const { bytes } = useHelpers()
  const path = typeof route.params.slug === 'string' ? route.params.slug : route.params.slug.join('/')

  const { data: meta } = await useFetch(`/api/meta/public/djvet/${path}`, {
    headers: { Accept: 'application/json' },
  })
  meta.name = decodeURI(meta.name)

  const { data: metadata } = await useFetch(`/api/tag/public/djvet/${path}`, {
    headers: { Accept: 'application/json' },
  })

  const tracks = ref(meta.value.members.filter(m => m.name.endsWith('.mp3')))
</script>
