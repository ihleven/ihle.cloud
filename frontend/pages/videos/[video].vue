<template>
  <main class="min-h-screen bg-neutral-200 text-black">
    <NavigationBar target="videos" />

    <section class="sticky top-0 w-screen">
      <video
        ref="videoElem"
        autoplay
        playsinline
        muted
        controls
        width="720"
        height="548"
        class="w-screen"
        crossorigin="anonymous"
      >
        <source :src="video.src" type="video/mp4" />

        <track v-if="video.chapters" default kind="chapters" label="weihnachten75" :src="video.chapters" srclang="de" />
      </video>

      <nav v-if="false" class="flex items-center justify-between p-4 backdrop-blur focus:outline-none">
        <nuxt-link class="flex h-12 w-12 items-center text-white" to="/videos">
          <Micon name="chevron-left" />
        </nuxt-link>
        <button class="text-white" @click="playPause()">
          <!-- <Icon name="fe:play" />
          <Icon name="fe:pause" /> -->
          <Micon name="play" class="h-12 w-12" filled />
        </button>
        <button class="text-white" @click="move(-0.05)">
          <Micon name="skip-back" />
        </button>
        <button class="text-white" @click="move(0.05)">
          <Micon name="skip-forward" />
        </button>
        <button class="text-white" @click="move(-0.01)">
          <Micon name="rewind" />
        </button>
        <button class="text-white" @click="move(0.01)">
          <Micon name="fast-forward" />
        </button>
      </nav>
    </section>

    <section class="mx-auto max-w-screen-md bg-neutral-200 pb-16 pt-8">
      <h2 class="text-md px-4 py-1 font-bold text-neutral-600">Kapitel:</h2>
      <ul class="divide-y divide-dashed divide-neutral-300 border-y border-neutral-300 bg-neutral-100">
        <li
          v-for="cue in chapters"
          :key="cue.id"
          class="flex cursor-pointer items-center justify-between px-2 py-1"
          :class="{ 'bg-sky-300': cue.id === active.id }"
          @click="jump(cue.startTime)"
        >
          <h3 class="text-sm font-semibold text-neutral-700">
            {{ cue.id }}
            <small class="text-xs font-medium text-neutral-300">{{ cue.startTime }} - {{ cue.endTime }}</small>
            <p class="text-sm font-light text-neutral-500">{{ cue.text }}</p>
          </h3>
          <svg class="h-6 w-6 text-neutral-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
          </svg>
        </li>
      </ul>
    </section>
  </main>
</template>

<script setup>
  const { params } = useRoute()
  const prefix = 'https://ihle.cloud/media/videos/'
  const videoElem = ref({})
  const time = ref(0)

  const chapters = ref([])
  const active = ref({})
  onMounted(() => {
    const track = videoElem.value.textTracks[0]
    if (track) {
      videoElem.value.textTracks[0].oncuechange = event => {
        if (track.cues) {
          chapters.value = track.cues
        }
        active.value = track.activeCues.length ? track.activeCues[0] : {}
      }
    }
    videoElem.onplaying = function () {
      console.log('Video is now loaded and playing')
    }
  })

  function playPause() {
    if (videoElem.value.paused) videoElem.value.play()
    else videoElem.value.pause()
  }

  function jump(ts) {
    time.value = ts
    videoElem.value.currentTime = ts
  }

  function move(ts) {
    time.value = videoElem.value.currentTime + ts
    videoElem.value.currentTime = time.value
  }

  function dur(seconds) {
    const date = new Date(0)
    date.setSeconds(seconds) // specify value for SECONDS here
    const minutes = date.getMinutes()
    const sec = date.getSeconds()
    const millis = date.getMilliseconds()
    return `${minutes}:${sec}.${millis}}`
  }

  const videos = {
    'goldene-hochzeit': {
      src: 'https://ihle.cloud/proxy/public/Super%208/Goldene%20Hochzeit%2018.10.74.mp4',
      chapters: null,
    },
  }

  const video = computed(() => {
    if (videos[params.video]) {
      return videos[params.video]
    }
    return {
      src: prefix + params.video + '/' + params.video + '.mp4',
      chapters: prefix + params.video + '/chapters.vtt',
    }
  })
</script>
