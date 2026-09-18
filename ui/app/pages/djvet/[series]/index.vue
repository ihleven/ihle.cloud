<template>
  <article ref="page" class="min-h-full w-full bg-elevated">
    <header class="flex items-baseline gap-2 px-3 py-2">
      <NuxtLink to="/djvet" class="text-xs text-muted hover:text-primary">DJ-Sets</NuxtLink>
      <span class="text-xs text-dimmed">/</span>
      <h1 class="text-sm font-semibold text-highlighted">{{ title }}</h1>
      <span v-if="cds?.length" class="text-xs text-muted">{{ cds.length }} CDs</span>

      <div class="grow" />

      <!-- The same handle the record shelf has, and for the same reason: how
           large a sleeve wants to be depends on the screen it is on. -->
      <input
        v-model.number="tile"
        type="range"
        min="100"
        max="320"
        step="20"
        class="w-24 accent-primary"
        aria-label="Größe der Kacheln"
      >
    </header>

    <p v-if="description" class="px-3 pb-2 text-xs text-muted">{{ description }}</p>

    <p v-if="pending" class="px-3 py-8 text-sm text-muted">Wird geladen…</p>

    <p v-else-if="error" class="px-3 py-8 text-sm text-error">Die Reihe ist nicht erreichbar.</p>

    <p v-else-if="!cds?.length" class="px-3 py-8 text-sm text-muted">
      Von dieser Reihe liegt noch nichts da.
    </p>

    <!-- The wall, as the record shelf draws it: a CD is recognised by its
         sleeve, so the titles stay in the tooltip and under the tile rather
         than over the picture. -->
    <ul
      v-else
      class="grid gap-2 p-2"
      :style="{ gridTemplateColumns: `repeat(auto-fill, minmax(${tile}px, 1fr))` }"
    >
      <li v-for="cd in cds" :key="cd.dir" class="group">
        <div class="relative aspect-square overflow-hidden rounded-sm bg-black">
          <img
            :data-src="cd.cover"
            :alt="cd.title"
            src="data:image/gif;base64,R0lGODlhAQABAAAAACH5BAEKAAEALAAAAAABAAEAAAICTAEAOw=="
            class="size-full object-cover transition duration-300 group-hover:scale-105"
            @error="blank"
          >

          <NuxtLink :to="`/djvet/${slug}/${cd.slug}`" class="absolute inset-0" :title="cd.title">
            <span class="sr-only">{{ cd.title }}</span>
          </NuxtLink>

          <!-- Opening the CD and playing it are different things, so they are
               different handles: the tile opens, the button plays. On a touch
               screen there is no hover, which is why the tile is the larger of
               the two. -->
          <div
            class="pointer-events-none absolute inset-x-0 bottom-0 flex justify-end p-1.5 opacity-0 transition group-hover:opacity-100 focus-within:opacity-100"
          >
            <UButton
              size="xs"
              color="neutral"
              variant="solid"
              icon="i-lucide-play"
              :loading="loading === cd.dir"
              class="pointer-events-auto rounded-full"
              :aria-label="`${cd.title} abspielen`"
              :title="`${cd.title} abspielen`"
              @click.stop.prevent="playCd(cd)"
            />
          </div>
        </div>
        <p class="truncate pt-1 text-xs text-muted">{{ cd.title }}</p>
      </li>
    </ul>
  </article>
</template>

<script setup lang="ts">
// One series, as a wall of its CDs.
//
// The overview shows a series as a row because it shows eight of them at once;
// here there is one, so it gets the shelf's own grid and can be made larger.
const route = useRoute()
const slug = computed(() => String(route.params.series ?? ''))

const { cdsOf, cd: readCd, track: tracks } = useDjvet()
const { play } = usePlayer()

const title = computed(() => seriesTitle(slug.value))
const description = computed(() => seriesDescription(slug.value))
const tile = ref(160)

// Twice the tile, for the reason the record shelf asks for twice: a screen may
// have two pixels to the tile's one, and a sleeve drawn from exactly its
// displayed width is soft on half the devices there are.
const coverWidth = computed(() => [200, 320, 480, 640].find(width => width >= tile.value * 2) ?? 640)

const { data: cds, pending, error } = await useAsyncData(
  () => `djvet-series-${slug.value}`,
  () => cdsOf(slug.value, coverWidth.value),
  { watch: [slug] },
)

// Playing a CD from the wall has to read it first: the wall knows sleeves and
// nothing else, which is what makes drawing it one request rather than eleven.
const loading = ref<string | undefined>()

async function playCd(entry: Cd) {
  loading.value = entry.dir
  try {
    const detail = await readCd(slug.value, entry.slug)
    play(detail.tracks.map(t => ({
      id: t.path,
      src: tracks(t.path),
      title: t.title,
      subtitle: t.artist || `${title.value} — ${detail.title}`,
      cover: detail.cover,
    })))
  }
  finally {
    loading.value = undefined
  }
}

function blank(event: Event) {
  const img = event.target as HTMLImageElement
  img.classList.add('opacity-0')
}

// The sleeves load as they are scrolled to. A grid does not fool lazy loading
// the way a sideways row does, but eighty sleeves is still eighty requests if
// they all go at once.
const page = useTemplateRef<HTMLElement>('page')
let watching: IntersectionObserver | undefined

function watchSleeves() {
  watching?.disconnect()
  if (!page.value) return

  watching = new IntersectionObserver((entries) => {
    for (const entry of entries) {
      if (!entry.isIntersecting) continue

      const sleeve = entry.target as HTMLImageElement
      if (sleeve.dataset.src) {
        sleeve.src = sleeve.dataset.src
        delete sleeve.dataset.src
      }
      watching?.unobserve(sleeve)
    }
  }, { rootMargin: '300px' })

  for (const sleeve of page.value.querySelectorAll<HTMLImageElement>('img[data-src]')) {
    watching.observe(sleeve)
  }
}

onMounted(watchSleeves)
watch(cds, () => nextTick(watchSleeves))
onBeforeUnmount(() => watching?.disconnect())
</script>
