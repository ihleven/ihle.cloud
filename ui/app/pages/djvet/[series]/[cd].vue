<template>
  <article class="min-h-full w-full bg-elevated">
    <header class="sticky top-0 flex items-baseline gap-2 border-b border-accented bg-default px-3 py-2">
      <NuxtLink to="/djvet" class="text-xs text-muted hover:text-primary">DJ-Sets</NuxtLink>
      <span class="text-xs text-dimmed">/</span>
      <NuxtLink :to="`/djvet/${series}`" class="text-xs text-muted hover:text-primary">
        {{ seriesName }}
      </NuxtLink>
      <span class="text-xs text-dimmed">/</span>
      <h1 class="text-sm font-semibold text-highlighted">{{ detail?.title ?? slug }}</h1>
    </header>

    <p v-if="pending" class="px-3 py-8 text-sm text-muted">Wird geladen…</p>

    <p v-else-if="error" class="px-3 py-8 text-sm text-error">Diese CD ist nicht erreichbar.</p>

    <!-- The titles beside the pictures, which is the shape the old page had and
         the right one: a tracklist is read down one column while the sleeve and
         the photographed listing sit next to it. -->
    <section
      v-else-if="detail"
      class="mx-auto flex max-w-(--ui-container) flex-col items-stretch md:flex-row"
    >
      <div class="w-full md:w-1/2">
        <div class="flex items-center gap-2 px-3 py-2">
          <UButton
            v-if="detail.tracks.length"
            size="xs"
            color="neutral"
            variant="solid"
            icon="i-lucide-play"
            :label="`Alle ${detail.tracks.length} Titel`"
            @click="playFrom(0)"
          />
          <span v-if="runtime" class="text-xs text-dimmed tabular-nums">{{ runtime }}</span>
        </div>

        <ol v-if="detail.tracks.length" class="divide-y divide-accented border-y border-accented">
          <li
            v-for="(item, index) in detail.tracks"
            :key="item.path"
            class="flex items-baseline gap-2 px-3 py-1.5"
            :class="{ 'text-primary': item.path === current?.id }"
          >
            <span class="w-5 shrink-0 text-right text-xs text-muted tabular-nums">
              {{ item.number ?? index + 1 }}
            </span>
            <button
              type="button"
              class="min-w-0 grow cursor-pointer text-left hover:text-primary"
              @click="playFrom(index)"
            >
              <span class="block truncate text-sm text-highlighted">{{ item.title }}</span>
              <span v-if="item.artist" class="block truncate text-xs font-light text-dimmed">
                {{ item.artist }}
              </span>
            </button>
            <span class="shrink-0 text-xs text-dimmed tabular-nums">
              {{ item.length || driveSize(item.size) }}
            </span>
          </li>
        </ol>

        <p v-else class="px-3 py-8 text-sm text-muted">
          Von dieser CD liegt noch nichts da — nur die Bilder.
        </p>
      </div>

      <div class="w-full md:w-1/2">
        <!-- Each picture is shown only once it has loaded. The old page drew
             both unconditionally and a missing file left a broken-image icon
             taking up a square of the layout. -->
        <img
          v-show="coverOk"
          :src="detail.cover"
          :alt="`Cover von ${detail.title}`"
          class="w-full object-cover"
          @error="coverOk = false"
        >
        <img
          v-show="listOk"
          :src="detail.tracklist"
          :alt="`Titelliste von ${detail.title}`"
          class="w-full object-cover"
          @error="listOk = false"
        >
        <p v-if="!coverOk && !listOk" class="px-3 py-8 text-sm text-muted">
          Zu dieser CD gibt es keine Bilder.
        </p>
      </div>
    </section>
  </article>
</template>

<script setup lang="ts">
// One CD: what is on it, beside the sleeve and the photographed tracklist.
//
// The titles come from the files' own tags, with the filenames as the fallback,
// because these were written with their titles in them and the filenames are
// frequently a number and nothing else.
const route = useRoute()
const series = computed(() => String(route.params.series ?? ''))
const slug = computed(() => String(route.params.cd ?? ''))

const { cd: readCd, track: tracks } = useDjvet()
const { play, current } = usePlayer()

const seriesName = computed(() => seriesTitle(series.value))

const { data: detail, pending, error } = await useAsyncData(
  () => `djvet-cd-${series.value}-${slug.value}`,
  () => readCd(series.value, slug.value),
  { watch: [series, slug] },
)

const coverOk = ref(true)
const listOk = ref(true)
watch(detail, () => {
  coverOk.value = !!detail.value?.cover
  listOk.value = !!detail.value?.tracklist
}, { immediate: true })

/** How long the CD runs, where every file says how long it is. */
const runtime = computed(() => {
  const items = detail.value?.tracks ?? []
  if (!items.length || items.some(t => !t.length)) return ''

  const seconds = items.reduce((total, t) => {
    const [minutes, rest] = (t.length ?? '0:00').split(':')

    return total + Number(minutes) * 60 + Number(rest)
  }, 0)

  return `${Math.round(seconds / 60)} Min.`
})

/**
 * From one track onwards.
 *
 * The whole CD goes to the player, not the one track pressed: what follows is
 * part of pressing play, and the player then carries on without this page. The
 * old page linked each track to the file browser instead, which left the
 * archive to be listened to one file at a time.
 */
function playFrom(at: number) {
  const cd = detail.value
  if (!cd) return

  play(cd.tracks.map(t => ({
    id: t.path,
    src: tracks(t.path),
    title: t.title,
    subtitle: t.artist || `${seriesName.value} — ${cd.title}`,
    cover: cd.cover,
  })), at)
}
</script>
