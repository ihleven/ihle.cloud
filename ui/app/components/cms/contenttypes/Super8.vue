<template>
  <article class="lg:flex">

    <section class="grow">
      <div class="grid grid-cols-12 justify-stretch gap-4 p-4">
        <ul class="col-span-12 divide-y divide-accented bg-elevated">
          <li v-for="(caption, i) in entry.content.captions" :key="caption.id">

            <header class="flex items-center" :class="{ 'bg-primary-100': currentTime ? timingstr2num(caption.ts[0]) < currentTime && currentTime < timingstr2num(caption.ts[1]) :null }">
              <UButton size="xs" variant="ghost" :label="`${caption.ts[0]}`" class="" @click="jump(caption.ts[0])" />
              <UIcon name="i-lucide-arrow-right" class="mx-4" />

              <UPopover
                mode="hover" :open-delay="200" :close-delay="100"
                :content="{
                  align: 'center',
                  side: 'bottom',
                  sideOffset: 2,
                }"
              >
                <UButton size="xs" variant="ghost" :label="`${caption.ts[1]}`" class="" @click="jump(caption.ts[1])" />

                <template #content>
                  <UButton size="xs" variant="ghost" label="Set" class="" @click="caption.ts[1]=num2timingstr(currentTime)" />
                </template>
              </UPopover>
              <UInput v-model="caption.id" variant="ghost" />
            </header>
            <UTextarea
              :value="caption.txt" variant="none" class="w-full"
              autoresize :rows="1"
              @update:model-value="update(`content.captions[${i}].txt`, $event)"
            />
          </li>

        </ul>
        <UFormField label="Titel:" class="col-span-6">
          <UInput
            :model-value="entry.name"
            placeholder="Titel" class="w-full"
            @update:model-value="update('name', $event)"
          />
        </UFormField>

        <UFormField label="Thumbnail:" class="col-span-6">
          <UInput
            :model-value="entry.content.thumbnail"
            placeholder="Thumbnail" class="w-full"
            @update:model-value="update('content.thumbnail', $event)"
          />
        </UFormField>

        <UFormField
          label="Source:"
          class="col-span-12"
        >
          <UInput
            placeholder="Source"
            :model-value="entry.content.src"
            class="w-full"
            @update:model-value="update('content.src', $event)"
          />
        </UFormField>

        <UFormField
          label="category:"
          class="col-span-3"
        >
          <USelect
            :model-value="entry.content.category"
            class="w-full"
            @update:model-value="update('content.category', $event)"
          />
        </UFormField>

        <UFormField
          label="Jahr:"
          class="col-span-4"
        >
          <UInputNumber
            placeholder="Jahr ..."
            :model-value="entry.content.year"
            :min="1970"
            :max="2030"
            :format-options="{
              minimumIntegerDigits: 4,
              maximumFractionDigits: 0,
              useGrouping: false,
            }"
            class="w-full"
            @update:model-value="update('content.year', $event)"
          />
        </UFormField>

        <UFormField
          label="VTT:"
          class="col-span-12"
        >
          <UTextarea
            autoresize
            placeholder="VTT"
            :model-value="entry.content.vtt"
            class="w-full"
            @update:model-value="update('content.vtt', $event)"
          />
        </UFormField>

        <UFormField
          label="Kommentar:"
          class="col-span-12"
        >
          <UTextarea
            placeholder="Kommentar"
            :model-value="entry.content.comments"
            class="w-full"
            @update:model-value="update('content.comments', $event)"
          />
        </UFormField>
      </div>

    </section>
    <section class="sticky top-12 z-10 w-screen lg:w-lg lg:border-l lg:border-accented">
      <video
        ref="video"
        autoplay
        playsinline
        muted
        controls
        width="720"
        height="548"
        class="w-screen"
        crossorigin="use-credentials"
      >
        <source :src="entry.content.src" type="video/mp4">

      <!-- <track
        v-if="video.chapters" default kind="chapters"
        label="weihnachten75" :src="video.chapters"
        srclang="de"
      > -->
      </video>
      <nav class="flex items-center justify-between space-x-4 border-b border-accented bg-white p-2">
        <!-- <nuxt-link class="flex h-12 w-12 items-center" to="/videos">
          <UIcon name="i-lucide-chevron-left" class="size-5" />
        </nuxt-link> -->
        <UButton
          size="xl" variant="ghost" color="neutral"
          :icon="paused ? 'i-lucide-play' : 'i-lucide-pause'"
          @click="playPause()"
        />
        <UFieldGroup size="xl" class="rounded">
          <UButton
            variant="ghost" color="neutral" :icon="'i-lucide-skip-back'"
            class=""
            @click="move(-0.5)"
          />

          <UButton
            variant="ghost" color="neutral" :icon="'i-lucide-rewind'"
            class=""
            @click="move(-0.1)"
          />
          <UButton
            variant="ghost" color="neutral" :icon="'i-lucide-step-back'"
            class=""
            @click="move(-0.01)"
          />
          <UButton
            variant="ghost" color="neutral" :icon="'i-lucide-step-forward'"
            class=""
            @click="move(0.01)"
          />
          <UButton
            variant="ghost" color="neutral" :icon="'i-lucide-fast-forward'"
            class=""
            @click="move(0.1)"
          />
          <UButton
            variant="ghost" color="neutral" :icon="'i-lucide-skip-forward'"
            class=""
            @click="move(0.5)"
          />
        </UFieldGroup>
        <USlider color="neutral" :default-value="50" tooltip />
      </nav>
      {{ currentTime }}
    </section>

  </article>
</template>

<script setup lang="ts">
const props = defineProps<{
  entry: Super8Entry
  variant?: 'standalone' | 'embed' | 'modal' | 'listitem'
}>()

const emit = defineEmits<{
  'update:entry': [Super8Entry]
  'patch:entry': [key: string, value: string | number | boolean | object]
}>()

function update(key: string, value: string) { emit('update:entry', cloneSetPath(props.entry, key, value) as ArtworkEntry) }
function patch(key: string, value: string) { emit('patch:entry', key, value) }

const video = ref<HTMLVideoElement | null>(null)
const paused = ref<boolean | undefined>(undefined)
const currentTime = ref<number | undefined>(undefined)

onMounted(() => {
  if (!video.value) return

  video.value.onplay = () => {
    paused.value = false
  }
  video.value.onpause = () => {
    paused.value = true
  }
  video.value.addEventListener('timeupdate', (event) => {
    currentTime.value = event.target?.currentTime
  })
})

const time = ref(0)

function playPause() {
  if (video.value?.paused) video.value.play()
  else video.value.pause()
}
function timingstr2num(ts: string): number {
  return new Date('1970-01-01T' + ts + 'Z').getTime() / 1000
}

function num2timingstr(ts: number): string {
  const date = new Date(ts * 1000)
  const hours = date.getUTCHours()
  const minutes = date.getUTCMinutes()
  const seconds = date.getUTCSeconds()
  const ms = date.getUTCMilliseconds()
  const timingstr = String(hours).padStart(2, '0') + ':' + String(minutes).padStart(2, '0') + ':' + String(seconds).padStart(2, '0') + ':' + String(ms).padStart(3, '0')

  return timingstr
}

function jump(ts: string) {
  const time = new Date('1970-01-01T' + ts + 'Z').getTime() / 1000
  console.log(ts, time)
  // time.value = time
  video.value.currentTime = time
}

function move(ts) {
  console.log(video.value.currentTime)
  time.value = video.value.currentTime + ts
  video.value.currentTime = time.value
  console.log(video.value.currentTime)
}
function fastSeek(ts) {
  console.log(video.value.currentTime)
  video.value.fastSeek(20)
  console.log(video.value.currentTime)
}
</script>
