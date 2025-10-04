<template>
  <main class="mx-auto min-h-screen max-w-screen-md bg-black text-white md:border-x">
    <NavigationBar target="musik" class="sticky top-0" />

    <section class="grid grid-cols-2">
      <img :src="media(`public/alben/${album}/cover.jpeg`)" class="aspect-square object-cover" />
      <aside>
        <figure v-if="selected" class="w-full">
          <audio controls :src="media(`public/alben/${album}/${selected.name}`)" class="rounded-none bg-gray-100">
            <a :href="media(`public/alben/${album}/${selected.name}`)"> Download audio </a>
          </audio>
        </figure>
      </aside>
    </section>

    <section>
      <ul class="divide-y divide-dashed divide-gray-500">
        <li v-for="track in tracks" :key="track.id" class="p-2">
          <div v-if="track.mime_type == 'audio/mpeg'" class="flex items-center justify-between">
            {{ decodeURI(track.name) }} 
            <button @click="play(track)">
              <svg
                xmlns="http://www.w3.org/2000/svg"
                fill="none"
                viewBox="0 0 24 24"
                stroke-width="1.5"
                stroke="currentColor"
                class="h-6 w-6"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  d="M5.25 5.653c0-.856.917-1.398 1.667-.986l11.54 6.348a1.125 1.125 0 010 1.971l-11.54 6.347a1.125 1.125 0 01-1.667-.985V5.653z"
                />
              </svg>
            </button>
          </div>
        </li>
      </ul>
    </section>
  </main>
</template>

<script setup>
const {media} = useHidrive()
  definePageMeta({    layout: 'default' })
  const { params } = useRoute()
  const album = params.slug.join("/")

  const selected = ref(null)

  const config = useRuntimeConfig()
  const { data: dir } = await useFetch(`${config.public.hidrive.api}/meta/public/alben/${album}/`, {
    server: false,
    credentials: 'include',
    headers: { Accept: 'application/json' },
  })

const tracks = computed(()=> dir.value? dir.value.members.filter(m => m.category === 'audio') : [])

  function play(member) {
    selected.value = member
  }

  
 
</script>
