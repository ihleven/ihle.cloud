<template>
  <main class="bg-neutral-200 text-black min-h-screen">
    <NavigationBar target="videos" />

    <section class="sticky top-0 w-screen">
      <video ref="videoElem" autoplay playsinline muted controls width="720" height="548" class="w-screen" crossorigin="anonymous">
        <source :src="video.src" type="video/mp4" />

        <track v-if="video.chapters" default kind="chapters" label="weihnachten75" :src="video.chapters" srclang="de" />
      </video>

      <nav v-if="false" class="backdrop-blur p-4 flex items-center justify-between focus:outline-none">
        <nuxt-link class="h-12 w-12 text-white flex items-center" to="/videos">
          <Micon name="chevron-left" />
        </nuxt-link>
        <button @click="playPause()" class="text-white">
          <!-- <Icon name="fe:play" />
          <Icon name="fe:pause" /> -->
          <Micon name="play" class="h-12 w-12" filled />
        </button>
        <button @click="move(-0.05)" class="text-white">
          <Micon name="skip-back" />
        </button>
        <button @click="move(0.05)" class="text-white">
          <Micon name="skip-forward" />
        </button>
        <button @click="move(-0.01)" class="text-white">
          <Micon name="rewind" />
        </button>
        <button @click="move(0.01)" class="text-white">
          <Micon name="fast-forward" />
        </button>
      </nav>
    </section>

    <section class="bg-neutral-200 max-w-screen-md mx-auto pt-8 pb-16">
      <h2 class="px-4 py-1 text-md font-bold text-neutral-600">Kapitel:</h2>
      <ul class="bg-neutral-100 border-y border-neutral-300 divide-y divide-neutral-300 divide-dashed">
        <li
          v-for="cue in chapters"
          :key="cue.id"
          @click="jump(cue.startTime)"
          class="flex justify-between items-center cursor-pointer px-2 py-1"
          :class="{ 'bg-sky-300': cue.id === active.id }"
        >
          <h3 class="text-neutral-700 text-sm font-semibold">
            {{ cue.id }}
            <small class="text-neutral-300 text-xs font-medium">{{ cue.startTime }} - {{ cue.endTime }}</small>
            <p class="text-neutral-500 text-sm font-light">{{ cue.text }}</p>
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
const { params } = useRoute();
const prefix = "https://ihle.cloud/media/videos/";
const videoElem = ref({});
const time = ref(0);

const chapters = ref([]);
const active = ref({});
onMounted(() => {
  const track = videoElem.value.textTracks[0];
  if (track) {
    videoElem.value.textTracks[0].oncuechange = (event) => {
      if (track.cues) {
        chapters.value = track.cues;
      }
      active.value = track.activeCues.length ? track.activeCues[0] : {};
    };
  }
  videoElem.onplaying = function () {
    console.log("Video is now loaded and playing");
  };
});

function playPause() {
  if (videoElem.value.paused) videoElem.value.play();
  else videoElem.value.pause();
}

function jump(ts) {
  time.value = ts;
  videoElem.value.currentTime = ts;
}

function move(ts) {
  time.value = videoElem.value.currentTime + ts;
  videoElem.value.currentTime = time.value;
}

function dur(seconds) {
  let date = new Date(0);
  date.setSeconds(seconds); // specify value for SECONDS here
  const minutes = date.getMinutes();
  const sec = date.getSeconds();
  const millis = date.getMilliseconds();
  return `${minutes}:${sec}.${millis}}`;
}

const videos = {
  "goldene-hochzeit": {
    src: "https://ihle.cloud/proxy/public/Super%208/Goldene%20Hochzeit%2018.10.74.mp4",
    chapters: null,
  },
};

const video = computed(() => {
  if (videos[params.video]) {
    return videos[params.video];
  }
  return { src: prefix + params.video + "/" + params.video + ".mp4", chapters: prefix + params.video + "/chapters.vtt" };
});
</script>
