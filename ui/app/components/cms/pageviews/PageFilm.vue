<template>
  <article class="w-full bg-elevated">
    <section class="sticky top-0 w-full bg-black">
      <!-- The overlay is positioned against the frame, so the wrapper has to be
           exactly the size of the picture rather than of the player. -->
      <div class="relative mx-auto aspect-video max-w-3xl">
        <video
          ref="video"
          controls
          playsinline
          crossorigin="use-credentials"
          class="h-full w-full bg-black"
          :poster="poster"
          :muted="silent"
          :autoplay="silent"
          @timeupdate="now = ($event.target as HTMLVideoElement).currentTime"
        >
          <source :src="src" type="video/mp4">
          <!-- Parsed by the browser rather than built by hand: the server
               renders the scenes as WebVTT, which is what this element expects. -->
          <track
            kind="chapters"
            srclang="de"
            label="Szenen"
            :src="chapters"
            default
          >
        </video>

        <span
          v-for="(annotation, i) in placed"
          :key="`${annotation.label}-${i}`"
          class="pointer-events-none absolute -translate-x-1/2 -translate-y-1/2 rounded bg-black/70 px-1.5 py-0.5 text-xs text-white ring-1 ring-white/30"
          :style="{ left: `${annotation.at!.x * 100}%`, top: `${annotation.at!.y * 100}%` }"
        >
          {{ annotation.label }}
        </span>
      </div>

      <!-- Frame stepping, for landing on an exact moment — which is what a
           scene's start time has to be. Only where the format's frame rate is
           known; see utils/film. -->
      <div v-if="frame" class="mx-auto flex max-w-3xl items-center justify-center gap-1 py-1 text-white">
        <UButton
          size="xs"
          variant="ghost"
          color="neutral"
          icon="i-lucide-chevrons-left"
          aria-label="Eine Sekunde zurück"
          title="Eine Sekunde zurück"
          @click="nudge(-1)"
        />
        <UButton
          size="xs"
          variant="ghost"
          color="neutral"
          icon="i-lucide-chevron-left"
          aria-label="Ein Bild zurück"
          title="Ein Bild zurück"
          @click="nudge(-frame)"
        />
        <!-- To the frame, not to the second: eighteen presses of the frame
             button would otherwise leave this reading unchanged. -->
        <span class="w-20 text-center text-xs text-white/70 tabular-nums">{{ timecode(now, fps!) }}</span>
        <UButton
          size="xs"
          variant="ghost"
          color="neutral"
          icon="i-lucide-chevron-right"
          aria-label="Ein Bild vor"
          title="Ein Bild vor"
          @click="nudge(frame)"
        />
        <UButton
          size="xs"
          variant="ghost"
          color="neutral"
          icon="i-lucide-chevrons-right"
          aria-label="Eine Sekunde vor"
          title="Eine Sekunde vor"
          @click="nudge(1)"
        />
      </div>
    </section>

    <div class="mx-auto max-w-3xl space-y-6 p-4">
      <header>
        <h1 class="text-xl font-semibold">{{ entry.name }}</h1>
        <p class="text-sm text-muted">
          {{ [entry.content.filmed_from, entry.content.location, entry.content.filmed_by].filter(Boolean).join(' · ') }}
        </p>
      </header>

      <!-- Whoever is on screen but has no position on the frame. -->
      <p v-if="unplaced.length" class="text-sm">
        <span class="text-muted">Zu sehen:</span> {{ unplaced.map(a => a.label).join(', ') }}
      </p>

      <ol v-if="scenes.length" class="divide-y divide-accented border-y border-accented">
        <li v-for="(scene, i) in scenes" :key="i">
          <button
            type="button"
            class="flex w-full gap-3 p-2 text-left hover:bg-default"
            :class="{ 'bg-default font-medium': i === currentScene }"
            @click="seek(scene.start)"
          >
            <span class="w-14 shrink-0 text-muted tabular-nums">{{ clock(scene.start) }}</span>
            <span>
              {{ scene.title }}
              <span v-if="scene.description" class="block text-sm text-muted">{{ scene.description }}</span>
            </span>
          </button>
        </li>
      </ol>
    </div>
  </article>
</template>

<script setup lang="ts">
const props = defineProps<{ entry: FilmEntry }>()

const video = ref<HTMLVideoElement | null>(null)
const now = ref(0)

const base = useRuntimeConfig().public.api.base as string

// Addressed by the entry's id. The storage key is read off the entry on the
// server, so where the bytes live never reaches the browser.
const id = computed(() => encodeURIComponent(props.entry.id))
const src = computed(() => `${base}/api/v1/films/${id.value}`)
const chapters = computed(() => `${base}/api/v1/films/${id.value}/chapters.vtt`)
const poster = computed(() => `${base}/api/v1/films/${id.value}/poster.jpg`)

/**
 * Whether the film can start on its own.
 *
 * Super 8 was shot on silent stock, so there is nothing to mute and nothing a
 * visitor can be startled by — which is also what makes starting it allowed at
 * all, since browsers only autoplay what makes no sound. The old pages played
 * every film this way and could, for the same reason: Super 8 was all there
 * was.
 *
 * Keyed off the format rather than assumed, so a format with sound — the reason
 * the field exists — waits to be asked rather than inheriting a decision that
 * was only ever true of silent film. The value matches films.FormatSuper8.
 */
const silent = computed(() => props.entry.content.format === 'super8')

/**
 * How fast this film ran, and how long one of its frames lasts.
 *
 * The transport appears only where the rate is known: a step of the wrong
 * length lands between frames, which looks like the film refusing to move.
 *
 * The rate is carried rather than recovered from the frame length — one divided
 * by one eighteenth is 18.000000000000004, and that is not a number to count
 * frames with.
 */
const fps = computed(() => frameRates[props.entry.content.format])
const frame = computed(() => (fps.value ? 1 / fps.value : undefined))

/**
 * Move by a fixed number of seconds.
 *
 * Pauses first. Stepping a frame at a time while the film is running is not
 * stepping — it is nudging something that is already moving away from you, and
 * the frame you chose is gone before you have looked at it.
 */
function nudge(by: number) {
  const el = video.value
  if (!el) return

  el.pause()
  el.currentTime = stepTo(el.currentTime, by, el.duration)
}

const scenes = computed<Scene[]>(() => props.entry.content.scenes ?? [])

// A scene runs until the next one starts, which is also how the chapter track
// is rendered — so the highlight and the browser's own chapter agree.
const currentScene = computed(() => {
  const t = now.value
  return scenes.value.findLastIndex(scene => scene.start <= t)
})

// How long an annotation lasts when no end was set. Must match the editor, or a
// name would linger here that the person placing it saw disappear.
const defaultDuration = 5

const active = computed(() =>
  (props.entry.content.annotations ?? []).filter(
    a => a.start <= now.value && now.value < (a.end ?? a.start + defaultDuration),
  ),
)
const placed = computed(() => active.value.filter(a => a.at))
const unplaced = computed(() => active.value.filter(a => !a.at))

function seek(seconds: number) {
  if (video.value) video.value.currentTime = seconds
}

function clock(seconds: number) {
  const total = Math.floor(seconds)
  return `${Math.floor(total / 60)}:${String(total % 60).padStart(2, '0')}`
}
</script>
