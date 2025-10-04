<template>
  <main class="min-h-screen bg-neutral-100">
    <section class="min-w-screen bg-black">
      <video
        v-if="video"
        ref="videoElem"
        autoplay
        playsinline
        muted
        controls
        class="mx-auto"
        :width="video.width"
        :height="video.height"
        :class="{ 'w-screen': screen }"
      >
        <source
          :src="video.src"
          type="video/mp4"
        >
      </video>
    </section>

    <nav class="flex items-center justify-between border-y border-gray-300">
      <nuxt-link
        to="/mediathek"
        class="flex items-center justify-start py-2 focus:outline-none"
      >
        <svg
          data-v-e29e7744=""
          viewBox="0 0 24 24"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
          class="h-8 w-8 fill-none stroke-current stroke-2"
          data-v-inspector="frontend/components/NavigationBar.vue:4:7"
        >
          <polyline
            data-v-e29e7744=""
            points="15 18 9 12 15 6"
            data-v-inspector="frontend/components/NavigationBar.vue:11:9"
          />
        </svg>
        <span class="text-md -m-1 font-medium tracking-tighter">mediathek</span>
      </nuxt-link>
      <h2 class="text-md px-4 py-1 font-semibold text-neutral-600">
        {{ dir?.name }}
      </h2>
      <button @click="toggleScreen()">
        screen
      </button>
    </nav>

    <section class="flex items-start border-0 border-neutral-300">
      <aside class="flex w-1/2 justify-end">
        <img
          :src="`http://localhost:8000/hi/media/public/mediathek/${params.slug[0]}/cover.jpg`"
          class="max-h-[50vh]"
        >
      </aside>

      <ul class="w-1/2 divide-y divide-gray-300 border-b border-neutral-300 bg-white px-4 text-neutral-700">
        <li
          v-for="f in files"
          :key="f.id"
          class="flex w-full cursor-pointer items-center justify-between p-1"
          @click="play(f)"
        >
          <h3 class="text-sm font-medium">
            {{ decodeURIComponent(f.name) }}
            <!-- <small class="text-xs font-medium text-neutral-300">{{ f.name }}</small> -->
            <p class="text-xs font-light text-neutral-500">
              {{ f.image.width }}x{{ f.image.height }} {{ f.mime_type }}, {{ bytes(f.size) }}
            </p>
          </h3>
        </li>
      </ul>
    </section>
  </main>
</template>

<script setup>
definePageMeta({
  layout: 'token',
  middleware: [
    function (to) {
      console.log('to', to)
    },
  ],
})

const { bytes } = useHelpers()

const { params, query } = useRoute()

console.log('params:', params, query)

const { data: dir } = await useFetch(`http://localhost:8000/hi/meta/public/mediathek/${params.slug[0]}/`, {
  server: false,
  credentials: 'include',
  headers: { Accept: 'application/json' },
})

const files = computed(() => (dir.value ? dir.value.members.filter(f => f.name.endsWith('.mp4')) : []))

const file = ref(files.value.length ? files.value[0] : null)

const prefix = 'http://localhost:8000/hi/media/public/mediathek/'

const video = computed(() => {
  return file.value
    ? {
        src: prefix + params.slug[0] + '/' + file.value.name,
        width: file.value.image.width,
        height: file.value.image.height,
      }
    : null
})

const videoElem = ref({})

async function play(f) {
  console.log(videoElem.value.src)
  file.value = f
  videoElem.value.load()
  await navigateTo({ query: { name: f.name } })
}

const screen = ref(false)
function toggleScreen() {
  screen.value = !screen.value
}
</script>
