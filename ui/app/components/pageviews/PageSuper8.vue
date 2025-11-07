<template>
  <article class="h-full w-full bg-elevated">
    <header class="h-12 w-full bg-green-500">-</header>
    <section class="sticky top-0 w-full bg-black">
      <video
        ref="video"
        autoplay
        playsinline
        muted
        controls
        width="720"
        height="548"
        class="mx-auto aspect-video bg-black"
        crossorigin="use-credentials"
      >
        <source :src="src" type="video/mp4">
      </video>
    </section>
    <div>src:{{ src }}</div>
    <!-- {{ entry }} -->
  </article>
</template>

<script setup lang="ts">
const props = defineProps<{ entry: Super8Entry }>()

const video = ref<HTMLVideoElement | null>(null)
// const track = ref(null)

const config = useRuntimeConfig()
const src = computed(() => `${config.public.api.base}/api/v1/super8/${props.entry.content.src}`)

onMounted(() => {
  if (!video.value) return
  const track = video.value.addTextTrack('captions', 'Captions', 'de')
  track.mode = 'showing'

  props.entry.content.captions?.forEach((cue) => {
    const start = new Date('1970-01-01T' + cue.ts[0] + 'Z').getTime() / 1000
    const end = new Date('1970-01-01T' + cue.ts[1] + 'Z').getTime() / 1000
    console.log('cue', cue.ts[0], start)
    const new_cue = new VTTCue(start, end, cue.txt)
    new_cue.id = cue.id
    console.log('cue', new_cue)
    track.addCue(new_cue)
  })
})
</script>
