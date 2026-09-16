<template>
  <article class="min-h-full w-full bg-elevated">
    <p v-if="pending" class="px-3 py-8 text-sm text-muted">Wird geladen…</p>

    <div v-else-if="error" class="px-3 py-8 text-sm">
      <p class="text-error">Diese Serie ist nicht erreichbar.</p>
    </div>

    <p v-else-if="!episodes.length" class="px-3 py-8 text-sm text-muted">
      Hier liegt nichts, was sich abspielen lässt.
    </p>

    <template v-else>
      <!-- The player and the bar that drives it travel together and stay at the
           top, so that reaching the fortieth episode of a series does not mean
           scrolling the thing you are watching off the screen. -->
      <div class="sticky top-0 z-10">
        <section class="bg-black">
          <video
            v-if="current"
            ref="player"
            :key="source"
            controls
            playsinline
            preload="metadata"
            class="mx-auto block h-auto max-w-full"
            :class="{ 'w-full': wide }"
            :width="chosen?.image?.width"
            :height="chosen?.image?.height"
          >
            <source :src="source" type="video/mp4">
          </video>
        </section>

        <nav class="flex items-center gap-2 border-b border-accented bg-default px-2 py-1">
          <NuxtLink
            to="/mediathek"
            class="flex shrink-0 items-center gap-0.5 text-sm text-muted hover:text-highlighted"
          >
            <UIcon name="i-lucide-chevron-left" class="size-5" />
            Mediathek
          </NuxtLink>

          <h1 class="grow truncate text-center text-sm font-semibold text-highlighted">
            {{ series.title }}
            <span v-if="current" class="font-normal text-muted">· {{ current.title }}</span>
          </h1>

          <div class="flex shrink-0 items-center gap-1">
            <!-- Two sizes, and neither is the other's smaller version: the
                 picture is 406 lines tall, so filling the window enlarges it
                 past what was recorded. That is a choice about how it looks,
                 not a way to see more of it, which is why it is offered rather
                 than simply done. -->
            <UButton
              size="xs"
              variant="ghost"
              color="neutral"
              :icon="wide ? 'i-lucide-fold-horizontal' : 'i-lucide-unfold-horizontal'"
              :aria-label="wide ? 'Originalgröße' : 'Volle Breite'"
              :title="wide ? 'Originalgröße' : 'Volle Breite'"
              @click="wide = !wide"
            />
            <UButton
              size="xs"
              variant="ghost"
              color="neutral"
              icon="i-lucide-maximize"
              aria-label="Vollbild"
              title="Vollbild"
              @click="fullscreen()"
            />
          </div>
        </nav>
      </div>

      <!-- Side by side where there is room for it. On a phone half a screen is
           too narrow for either, so the cover goes above the list rather than
           beside it. -->
      <div class="sm:flex sm:items-start">
        <aside class="flex justify-center p-3 sm:w-1/2 sm:justify-end">
          <img
            :src="cover"
            :alt="series.title"
            class="max-h-[60vh] rounded-sm"
          >
        </aside>

        <ul class="divide-y divide-accented sm:w-1/2 sm:border-l sm:border-accented">
          <li
            v-for="episode in episodes"
            :key="episode.title + (episode.number ?? '')"
            class="px-3 py-1.5"
            :class="{ 'bg-primary-50 dark:bg-primary-950': episode === current }"
          >
            <button type="button" class="w-full cursor-pointer text-left" @click="play(episode)">
              <span class="flex items-baseline gap-2">
                <span v-if="episode.number" class="shrink-0 text-xs text-muted tabular-nums">{{ episode.number }}</span>
                <span class="truncate text-sm text-highlighted">{{ episode.title }}</span>
              </span>
              <!-- What the file is, under what it is called: the shelf has always
                   described its episodes this way, and it is the line that tells
                   two copies of one episode apart. -->
              <span class="block text-xs font-light text-dimmed">{{ describe(shown(episode)) }}</span>
            </button>

            <!-- A choice of copy, only where there is one to make: one episode on
                 the whole shelf is kept four times. Labelled by size, which is
                 what actually differs between them — see episodesOf. -->
            <div v-if="episode === current && episode.variants.length > 1" class="mt-1 flex flex-wrap gap-1">
              <UButton
                v-for="variant in episode.variants"
                :key="variant.name"
                size="xs"
                :variant="variant === chosen ? 'solid' : 'ghost'"
                :label="driveSize(variant.size)"
                @click="chosen = variant"
              />
            </div>
          </li>
        </ul>
      </div>
    </template>
  </article>
</template>

<script setup lang="ts">
// One series: the player above, the episodes beside the cover below it.
//
// The episodes are read out of the folder's filenames rather than recorded
// anywhere, with all the unevenness that implies — see useMediathek for what
// each of those readings assumes and where it is already known to be wrong.
const route = useRoute()
const dir = computed(() => String(route.params.dir ?? ''))

const { meta, stream } = useMediathek()
const { data: listing, pending, error } = await meta(dir)

const series = computed(() => titleOf(dir.value))
const cover = computed(() => stream(`${dir.value}/cover.jpg`))
const episodes = computed(() => (listing.value ? episodesOf(listing.value) : []))

// Shallow, and that is load-bearing: a deep ref hands back a reactive proxy of
// what was put in it, which is no longer identical to the object in the list, so
// every `episode === current` comparison in the template would be false.
const current = shallowRef<Episode | null>(null)
const chosen = shallowRef<DriveMeta | null>(null)

// The first episode, loaded but not started. With the player at the top of the
// page the alternative is opening on an empty black band, and the first episode
// is what somebody who opened a series was reaching for anyway.
watchEffect(() => {
  if (!current.value && episodes.value.length) select(episodes.value[0]!)
})

function select(episode: Episode) {
  current.value = episode
  // The variants come best-first, so the default is the best copy; a lesser one
  // is a choice somebody makes, not one made for them. See episodesOf.
  chosen.value = episode.variants[0] ?? null
}

function play(episode: Episode) {
  select(episode)
  // The source changes with it, so playing has to wait for the new one to be in
  // the element.
  nextTick(() => {
    player.value?.play().catch(() => {
      // Refused when the browser does not credit the click to this element. The
      // episode is loaded either way and its play button is right there.
    })
  })
}

const source = computed(() => (chosen.value ? stream(`${dir.value}/${driveName(chosen.value)}`) : ''))

/** The copy a row describes: the one being watched, else the one it would play. */
function shown(episode: Episode): DriveMeta | null {
  return episode === current.value ? chosen.value : (episode.variants[0] ?? null)
}

function describe(file: DriveMeta | null): string {
  if (!file) return ''

  const pixels = file.image ? `${file.image.width}×${file.image.height}` : ''

  return [pixels, file.mime_type, driveSize(file.size)].filter(Boolean).join(', ')
}

const player = useTemplateRef<HTMLVideoElement>('player')

// Wider than it was recorded, for a screen further from the eye. Not remembered
// between visits: it is a choice about this sitting, and the size the recording
// actually has is the honest one to open with.
const wide = ref(false)

function fullscreen() {
  player.value?.requestFullscreen().catch(() => {
    // Refused or unsupported is not worth saying: the player's own control in
    // the bottom corner offers the same thing.
  })
}
</script>
