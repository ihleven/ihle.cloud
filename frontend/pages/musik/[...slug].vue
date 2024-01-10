<template>
  <main class="mx-auto min-h-screen max-w-screen-md bg-black text-white md:border-x">
    <NavigationBar target="musik" class="sticky top-0" />

    <section class="grid grid-cols-2">
      <img :src="`/api/raw/public/alben/${album}/cover.jpeg`" class="aspect-square object-cover" />
      <aside>
        <figure v-if="track" class="w-full">
          <!-- <figcaption>Listen:</figcaption> -->
          <audio controls :src="`/api/raw/public/alben/${album}/${track.name}`" class="rounded-none bg-gray-100">
            <a href="/media/cc0-audio/t-rex-roar.mp3"> Download audio </a>
          </audio>
        </figure>
      </aside>
    </section>

    <section>
      <ul v-if="dir" class="divide-y divide-dashed divide-gray-500">
        <li v-for="m in dir.members" :key="m.id" class="p-2">
          <div v-if="m.mime_type == 'audio/mpeg'" class="flex items-center justify-between">
            {{ decodeURI(m.name) }} {{ m.category }}
            <button @click="play(m)">
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
  definePageMeta({
    layout: 'default',
  })
  const { params } = useRoute()
  const album = params.slug

  const track = ref(null)
  // const time = ref(0)

  // const chapters = ref([])
  // const active = ref({})
  // onMounted(() => {
  //   const track = videoElem.value.textTracks[0]
  //   if (track) {
  //     videoElem.value.textTracks[0].oncuechange = event => {
  //       if (track.cues) {
  //         chapters.value = track.cues
  //       }
  //       active.value = track.activeCues.length ? track.activeCues[0] : {}
  //     }
  //   }
  //   videoElem.onplaying = function () {
  //     console.log('Video is now loaded and playing')
  //   }
  // })

  function play(member) {
    track.value = member
  }

  const { data: dir } = await useFetch(`/api/meta/public/alben/${params.slug[0]}`, {
    server: false,
    headers: { Accept: 'application/json' },
  })
  console.log('params:', params.slug, dir.value)
</script>
