<template>
  <main class="bg-black min-h-screen">
    <NavigationBar target="videos" />
    <section class="sticky top-0 w-screen">
      <video
        ref="video"
        autoplay
        playsinline
        muted
        controls
        width="720"
        height="548"
        class="w-screen"
        crossorigin="anonymous"
      >
        <source
          :src="prefix + params.video + '/' + params.video + '.mp4'"
          type="video/mp4"
        />
        <track
          default
          kind="chapters"
          label="weihnachten75"
          :src="prefix + params.video + '/chapters.vtt'"
          srclang="de"
        />
      </video>

      <nav
        v-if="false"
        class="backdrop-blur p-4 flex items-center justify-between focus:outline-none"
      >
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
      <ul
        class="bg-neutral-100 border-y border-neutral-300 divide-y divide-neutral-300 divide-dashed"
      >
        <li
          v-for="cue in chapters"
          :key="cue.id"
          @click="jump(cue.startTime)"
          class="flex justify-between items-center cursor-pointer px-2 py-1"
          :class="{ 'bg-sky-300': cue.id === active.id }"
        >
          <h3 class="text-neutral-700 text-sm font-semibold">
            {{ cue.id }}
            <small class="text-neutral-300 text-xs font-medium"
              >{{ cue.startTime }} - {{ cue.endTime }}</small
            >
            <p class="text-neutral-500 text-sm font-light">{{ cue.text }}</p>
          </h3>
          <svg
            class="h-6 w-6 text-neutral-400"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M9 5l7 7-7 7"
            />
          </svg>
        </li>
      </ul>
    </section>
  </main>
</template>

<script setup>
const { params } = useRoute();
const prefix = "https://ihle.cloud/media/videos/";
const video = ref({});
const time = ref(0);

const chapters = ref([]);
const active = ref({});
onMounted(() => {
  const track = video.value.textTracks[0];
  video.value.textTracks[0].oncuechange = (event) => {
    if (track.cues) {
      chapters.value = track.cues;
    }
    active.value = track.activeCues.length ? track.activeCues[0] : {};
  };
  video.onplaying = function () {
    console.log("Video is now loaded and playing");
  };
});

function playPause() {
  if (video.value.paused) video.value.play();
  else video.value.pause();
}

function jump(ts) {
  time.value = ts;
  video.value.currentTime = ts;
}

function move(ts) {
  time.value = video.value.currentTime + ts;
  video.value.currentTime = time.value;
}

function dur(seconds) {
  let date = new Date(0);
  date.setSeconds(seconds); // specify value for SECONDS here
  const minutes = date.getMinutes();
  const sec = date.getSeconds();
  const millis = date.getMilliseconds();
  return `${minutes}:${sec}.${millis}}`;
}
</script>
