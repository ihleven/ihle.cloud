<template>
  <article class="lg:flex lg:items-start">

    <section class="grow space-y-6 p-4">

      <!-- Scenes segment the film, so they are a sequence: only the start is
           edited, and each one runs until the next begins. -->
      <div>
        <header class="flex items-center justify-between border-b border-accented pb-1">
          <h3 class="font-medium">Szenen</h3>
          <UButton size="xs" variant="ghost" icon="i-lucide-plus" :label="`bei ${clock(now)}`" @click="addScene" />
        </header>
        <ul class="divide-y divide-accented">
          <li
            v-for="(scene, i) in scenes"
            :key="i"
            class="flex items-start gap-1 py-1"
            :class="{ 'bg-primary-50 dark:bg-primary-950': i === currentScene }"
          >
            <UButton
              size="xs" variant="ghost" class="w-14 shrink-0 cursor-pointer tabular-nums"
              :label="clock(scene.start)" :title="`Zur Position springen (${timecode(scene.start)})`"
              @click="seek(scene.start)"
            />
            <UButton
              size="xs" variant="ghost" icon="i-lucide-crosshair"
              title="Start auf die aktuelle Position setzen"
              @click="update(`content.scenes[${i}].start`, round(now))"
            />
            <div class="grow">
              <UInput
                :model-value="scene.title" variant="ghost" placeholder="Titel" class="w-full"
                @update:model-value="update(`content.scenes[${i}].title`, $event)"
              />
              <UTextarea
                :model-value="scene.description" variant="none" placeholder="Beschreibung" class="w-full"
                autoresize :rows="1"
                @update:model-value="update(`content.scenes[${i}].description`, $event)"
              />
            </div>
            <UButton size="xs" color="neutral" variant="ghost" icon="i-lucide-trash-2" @click="removeScene(i)" />
          </li>
        </ul>
      </div>

      <!-- Annotations are marks on the timeline, not parts of a scene: the list
           shows all of them so there is an overview, and the scene filter is a
           view rather than a constraint. -->
      <div>
        <header class="flex items-center justify-between border-b border-accented pb-1">
          <h3 class="font-medium">
            Annotationen
            <span class="text-muted">({{ shownAnnotations.length }}<template v-if="onlyThisScene">/{{ annotations.length }}</template>)</span>
          </h3>
          <div class="flex items-center gap-1">
            <UButton
              size="xs" :variant="onlyThisScene ? 'soft' : 'ghost'" icon="i-lucide-filter"
              :label="onlyThisScene ? 'nur diese Szene' : 'alle'"
              :disabled="currentScene < 0"
              @click="onlyThisScene = !onlyThisScene"
            />
            <UButton size="xs" variant="ghost" icon="i-lucide-plus" :label="`Person bei ${clock(now)}`" @click="addAnnotation" />
          </div>
        </header>
        <ul class="max-h-96 divide-y divide-accented overflow-y-auto">
          <li
            v-for="{ annotation, index } in shownAnnotations"
            :key="index"
            class="flex items-center gap-1 py-1"
            :class="{ 'bg-primary-50 dark:bg-primary-950': isNow(annotation) }"
          >
            <UButton
              size="xs" variant="ghost" color="neutral" icon="i-lucide-circle-play"
              class="shrink-0 cursor-pointer"
              :title="`Zur Position springen (${timecode(annotation.start)})`"
              @click="seek(annotation.start)"
            />
            <USelect
              :model-value="annotation.kind" :items="kinds" size="xs" class="w-24 shrink-0"
              @update:model-value="update(`content.annotations[${index}].kind`, $event)"
            />
            <UInput
              :model-value="annotation.label" variant="ghost" placeholder="Name" class="grow"
              @update:model-value="update(`content.annotations[${index}].label`, $event)"
            />

            <!-- Both ends jump and both can be set from the playhead. An end
                 that was never set shows the default it will behave as. -->
            <UButton
              size="xs" variant="ghost" class="cursor-pointer tabular-nums"
              :label="clock(annotation.start)" :title="`Zur Position springen (${timecode(annotation.start)})`"
              @click="seek(annotation.start)"
            />
            <UButton
              size="xs" variant="ghost" icon="i-lucide-crosshair" title="Anfang hierher"
              @click="update(`content.annotations[${index}].start`, round(now))"
            />
            <span class="text-muted">–</span>
            <UButton
              size="xs" variant="ghost" class="cursor-pointer tabular-nums"
              :class="{ 'italic opacity-60': !annotation.end }"
              :label="clock(endOf(annotation))"
              :title="annotation.end
                ? `Zum Ende springen (${timecode(annotation.end)})`
                : `Zum Ende springen (${timecode(endOf(annotation))}, Standard ${defaultDuration} s)`"
              @click="seek(endOf(annotation))"
            />
            <UButton
              size="xs" variant="ghost" icon="i-lucide-crosshair" title="Ende hierher"
              @click="update(`content.annotations[${index}].end`, round(now))"
            />
            <UButton
              v-if="annotation.end" size="xs" color="neutral" variant="ghost" icon="i-lucide-undo-2"
              :title="`Ende zurücksetzen (${defaultDuration} s)`"
              @click="update(`content.annotations[${index}].end`, null)"
            />

            <UButton
              size="xs" variant="ghost"
              :icon="annotation.at ? 'i-lucide-map-pin' : 'i-lucide-map-pin-off'"
              :color="placing === index ? 'primary' : 'neutral'"
              title="Im Bild platzieren"
              @click="placing = placing === index ? null : index"
            />
            <UButton
              v-if="annotation.at" size="xs" color="neutral" variant="ghost" icon="i-lucide-x"
              title="Position entfernen"
              @click="update(`content.annotations[${index}].at`, null)"
            />
            <UButton size="xs" color="neutral" variant="ghost" icon="i-lucide-trash-2" @click="removeAnnotation(index)" />
          </li>
          <li v-if="!shownAnnotations.length" class="py-2 text-center text-sm text-muted">
            Noch keine Annotationen.
          </li>
        </ul>
      </div>

      <div class="grid grid-cols-12 gap-4">
        <UFormField label="Titel:" class="col-span-6">
          <UInput :model-value="entry.name" class="w-full" @update:model-value="update('name', $event)" />
        </UFormField>
        <UFormField label="Format:" class="col-span-6">
          <UInput :model-value="entry.content.format" class="w-full" @update:model-value="update('content.format', $event)" />
        </UFormField>
        <UFormField label="Key:" class="col-span-12" help="Wo die Datei liegt. Keine URL.">
          <UInput :model-value="entry.content.key" class="w-full" @update:model-value="update('content.key', $event)" />
        </UFormField>
        <UFormField label="Poster:" class="col-span-6" help="Leer: ein Bild aus dem Film.">
          <UInput :model-value="entry.content.poster" class="w-full" @update:model-value="update('content.poster', $event)" />
        </UFormField>
        <UFormField label="Ort:" class="col-span-6">
          <UInput :model-value="entry.content.location" class="w-full" @update:model-value="update('content.location', $event)" />
        </UFormField>
        <UFormField label="Gefilmt von:" class="col-span-6">
          <UInput :model-value="entry.content.filmed_by" class="w-full" @update:model-value="update('content.filmed_by', $event)" />
        </UFormField>
        <UFormField label="Gefilmt am:" class="col-span-6">
          <UInput :model-value="entry.content.filmed_from" class="w-full" @update:model-value="update('content.filmed_from', $event)" />
        </UFormField>
        <UFormField label="Beschreibung:" class="col-span-12">
          <UTextarea :model-value="entry.content.description" autoresize :rows="2" class="w-full" @update:model-value="update('content.description', $event)" />
        </UFormField>
      </div>
    </section>

    <!-- The picture stays put while the lists scroll: every edit here is made
         while looking at a frame. -->
    <section class="top-12 shrink-0 lg:sticky lg:w-lg lg:border-l lg:border-accented">
      <div
        class="relative aspect-video w-full bg-black"
        :class="{ 'cursor-crosshair ring-2 ring-primary': placing !== null }"
        @click="place"
      >
        <video
          ref="video"
          playsinline
          controls
          crossorigin="use-credentials"
          class="h-full w-full"
          :src="src"
          @timeupdate="now = ($event.target as HTMLVideoElement).currentTime"
          @seeked="now = ($event.target as HTMLVideoElement).currentTime"
          @play="paused = false"
          @pause="paused = true"
          @loadedmetadata="duration = ($event.target as HTMLVideoElement).duration"
        />
        <span
          v-for="(annotation, i) in placed"
          :key="`${annotation.label}-${i}`"
          class="pointer-events-none absolute -translate-x-1/2 -translate-y-1/2 rounded bg-black/70 px-1.5 py-0.5 text-xs text-white ring-1 ring-white/30"
          :style="{ left: `${annotation.at!.x * 100}%`, top: `${annotation.at!.y * 100}%` }"
        >{{ annotation.label }}</span>
      </div>

      <!-- Moving the film while it is paused is what makes a boundary placeable:
           at 25fps a hundredth of a second is a quarter of a frame. -->
      <nav class="flex items-center justify-between gap-2 border-b border-accented p-2">
        <UButton
          size="lg" variant="ghost" color="neutral"
          :icon="paused ? 'i-lucide-play' : 'i-lucide-pause'"
          :title="paused ? 'Abspielen (Leertaste)' : 'Pause (Leertaste)'"
          @click="playPause"
        />
        <UFieldGroup size="sm" class="rounded">
          <UButton
            v-for="step in steps" :key="step"
            variant="ghost" color="neutral" :icon="stepIcon(step)"
            :title="`${step > 0 ? '+' : ''}${step}s`"
            @click="move(step)"
          />
        </UFieldGroup>
        <output class="w-24 shrink-0 text-right font-mono text-sm tabular-nums">{{ timecode(now) }}</output>
      </nav>
      <p v-if="placing !== null" class="p-2 text-center text-sm text-primary">
        Klick ins Bild, um «{{ annotations[placing]?.label || 'die Annotation' }}» zu platzieren.
      </p>
    </section>

  </article>
</template>

<script setup lang="ts">
const props = defineProps<{
  entry: FilmEntry
  variant?: 'standalone' | 'embed' | 'modal' | 'listitem'
}>()

const emit = defineEmits<{
  'update:entry': [FilmEntry]
}>()

function update(key: string, value: unknown) {
  emit('update:entry', cloneSetPath(props.entry, key, value as object) as FilmEntry)
}

const kinds = ['person', 'location', 'object', 'note']

// How long an annotation lasts when no end was set. A person is visible for a
// moment, not for a whole scene, so the default is short and the end is there
// to be adjusted rather than to be filled in every time.
const defaultDuration = 5

// Coarse, medium and fine. The finest is under a frame, which is what a scene
// boundary has to be placed to.
const steps = [-0.5, -0.1, -0.01, 0.01, 0.1, 0.5]

function stepIcon(step: number) {
  const size = Math.abs(step)
  if (size >= 0.5) return step < 0 ? 'i-lucide-skip-back' : 'i-lucide-skip-forward'
  if (size >= 0.1) return step < 0 ? 'i-lucide-rewind' : 'i-lucide-fast-forward'
  return step < 0 ? 'i-lucide-step-back' : 'i-lucide-step-forward'
}

const video = ref<HTMLVideoElement | null>(null)
const now = ref(0)
const paused = ref(true)
const duration = ref(0)
// The annotation waiting for a click on the frame, by index, or null.
const placing = ref<number | null>(null)
const onlyThisScene = ref(false)

const toast = useToast()
const { session } = useAuth()

const src = computed(() => `${useRuntimeConfig().public.api.base}/api/v1/films/${encodeURIComponent(props.entry.id)}`)

const scenes = computed<Scene[]>(() => props.entry.content.scenes ?? [])
const annotations = computed<Annotation[]>(() => props.entry.content.annotations ?? [])

const currentScene = computed(() => scenes.value.findLastIndex(scene => scene.start <= now.value))

// A scene ends where the next one starts; the last runs to the end of the film.
function sceneEnd(i: number): number {
  return scenes.value[i + 1]?.start ?? duration.value ?? Number.MAX_SAFE_INTEGER
}

function isNow(annotation: Annotation) {
  return annotation.start <= now.value && now.value < endOf(annotation)
}

// An annotation without an explicit end runs for the default duration.
function endOf(annotation: Annotation) {
  return annotation.end ?? annotation.start + defaultDuration
}

// Every annotation, in time order, so there is an overview. Keeping the index
// makes each edit a targeted patch rather than a rewrite of the whole array.
// The scene filter narrows the view when a film has many; it is not what an
// annotation belongs to.
const shownAnnotations = computed(() => {
  const all = annotations.value
    .map((annotation, index) => ({ annotation, index }))
    .sort((a, b) => a.annotation.start - b.annotation.start)

  const i = currentScene.value
  if (!onlyThisScene.value || i < 0) return all

  const from = scenes.value[i]!.start
  const to = sceneEnd(i)
  return all.filter(({ annotation }) => annotation.start < to && endOf(annotation) > from)
})

const placed = computed(() => annotations.value.filter(a => a.at && isNow(a)))

function round(seconds: number) {
  return Math.round(seconds * 1000) / 1000
}

// The readout is written here rather than left to timeupdate: a paused video
// reports a seek late, and a readout that lags is useless for the one job it
// has — showing what you are nudging.
function seek(seconds: number) {
  if (!video.value) return
  video.value.currentTime = seconds
  now.value = video.value.currentTime
}

function playPause() {
  if (!video.value) return
  if (video.value.paused) video.value.play()
  else video.value.pause()
}

// Nudging past either end is not an error, it is the end of the film.
function move(delta: number) {
  if (!video.value) return
  const max = duration.value || video.value.duration || 0
  video.value.currentTime = Math.min(Math.max(video.value.currentTime + delta, 0), max)
  now.value = video.value.currentTime
}

// M:SS for a list, where the point is to read it.
function clock(seconds: number) {
  const total = Math.floor(seconds)
  return `${Math.floor(total / 60)}:${String(total % 60).padStart(2, '0')}`
}

// M:SS.mmm for the readout, where the point is to see what you are nudging.
function timecode(seconds: number) {
  const ms = Math.round(seconds * 1000)
  const m = Math.floor(ms / 60000)
  const s = Math.floor(ms / 1000) % 60
  return `${m}:${String(s).padStart(2, '0')}.${String(ms % 1000).padStart(3, '0')}`
}

// Scenes must start strictly after one another, which the server enforces; the
// editor keeps them sorted so that is never the reason a save is refused.
function addScene() {
  const start = round(now.value)
  if (scenes.value.some(scene => Math.abs(scene.start - start) < 0.05)) {
    toast.add({ title: 'Hier beginnt schon eine Szene', color: 'warning' })
    return
  }
  update('content.scenes', [...scenes.value, { title: 'Neue Szene', start }].sort((a, b) => a.start - b.start))
}

function removeScene(i: number) {
  update('content.scenes', scenes.value.filter((_, index) => index !== i))
}

// A new annotation starts at the playhead and runs for the default duration —
// it marks the moment you are looking at, not the scene you are in. Kept in time
// order so the stored file reads the way the list does.
function addAnnotation() {
  update('content.annotations', [...annotations.value, {
    kind: 'person',
    label: '',
    start: round(now.value),
    author: session.value?.sub,
  }].sort((a, b) => a.start - b.start))
}

function removeAnnotation(index: number) {
  if (placing.value === index) placing.value = null
  update('content.annotations', annotations.value.filter((_, i) => i !== index))
}

// Fractions of the frame, measured the same way the viewer draws them, so what
// is placed here lands there.
function place(event: MouseEvent) {
  if (placing.value === null) return
  const box = (event.currentTarget as HTMLElement).getBoundingClientRect()
  update(`content.annotations[${placing.value}].at`, {
    x: round((event.clientX - box.left) / box.width),
    y: round((event.clientY - box.top) / box.height),
  })
  placing.value = null
}

// Watch, pause, nudge, type a name, repeat — reaching for the buttons is the
// slow part, so the transport is on the keyboard too. Typing into a field is
// still typing: the shortcuts stay out of the way there.
function onKey(event: KeyboardEvent) {
  if (!video.value || event.metaKey || event.ctrlKey) return

  const target = event.target as HTMLElement | null
  if (target && (target.isContentEditable || ['INPUT', 'TEXTAREA', 'SELECT'].includes(target.tagName))) return

  if (event.code === 'Space') {
    event.preventDefault()
    playPause()
    return
  }
  if (event.key !== 'ArrowLeft' && event.key !== 'ArrowRight') return

  event.preventDefault()
  const size = event.shiftKey ? 0.5 : event.altKey ? 0.01 : 0.1
  move(event.key === 'ArrowLeft' ? -size : size)
}

onMounted(() => window.addEventListener('keydown', onKey))
onBeforeUnmount(() => window.removeEventListener('keydown', onKey))
</script>
