<template>
  <main class="min-h-screen">
    <NavigationBar target="mediathek" />

    <section class="sticky top-0 w-screen max-w-[100vh] mx-auto">
      <video ref="videoElem" autoplay playsinline muted controls width="720" height="548" class="w-screen" crossorigin="anonymous">
        <source :src="video.src" type="video/mp4" />

        <!-- <track v-if="video.chapters" default kind="chapters" label="weihnachten75" :src="video.chapters" srclang="de" /> -->
      </video>
    </section>

    <section class="max-w-screen-md mx-auto pt-8 pb-16">
      <!-- <h2 class="px-4 py-1 text-md font-bold text-neutral-600">Kapitel:</h2> -->

      <ul class="bg-neutral-700 text-neutral-300 border-y border-neutral-300">
        <li v-for="file in dir.members" :key="file.id">
          <nuxt-link
            v-if="file.name.endsWith('.Creator2160p60.mp4')"
            :to="`/mediathek/${params.slug[0]}/${file.name.replace('.Creator2160p60.mp4', '')}`"
            class="flex justify-between items-center cursor-pointer px-2 py-1"
          >
            <h3 class="text-sm font-semibold">
              {{ file.name.replace(".Creator2160p60.mp4", "") }}
              <!-- <small class="text-neutral-300 text-xs font-medium">{{ cue.startTime }} - {{ cue.endTime }}</small>
            <p class="text-neutral-500 text-sm font-light">{{ cue.text }}</p> -->
            </h3>
            <svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
            </svg>
          </nuxt-link>
        </li>
      </ul>
    </section>
  </main>
</template>

<script setup>
const { params } = useRoute();
console.log("params:", params);
const prefix = "http://localhost:10815/proxy/public/mediathek/";
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

const video = computed(() => {
  return { src: prefix + params.slug.join("/") + ".Creator2160p60.mp4" };
});

const { data: dir, refresh } = await useFetch(`/api/hidrive/public/mediathek/${params.slug[0]}`, { server: false });
console.log("dir", dir);
</script>
