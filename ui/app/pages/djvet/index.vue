<template>
  <article ref="page" class="min-h-full w-full bg-elevated">
    <header class="flex items-baseline gap-2 px-3 py-2">
      <h1 class="text-sm font-semibold text-highlighted">DJ-Sets</h1>
      <span v-if="cdCount" class="text-xs text-muted">{{ cdCount }} CDs</span>
    </header>

    <p v-if="pending" class="px-3 py-8 text-sm text-muted">Wird geladen…</p>

    <div v-else-if="error" class="px-3 py-8 text-sm">
      <p class="text-error">{{ problem.what }}</p>
      <p v-if="problem.why" class="mt-1 text-muted">{{ problem.why }}</p>
    </div>

    <p v-else-if="!archive?.length" class="px-3 py-8 text-sm text-muted">
      Hier steht noch nichts.
    </p>

    <section v-for="series in archive" v-else :key="series.slug" class="pb-4">
      <NuxtLink :to="`/djvet/${series.slug}`" class="block px-3 pt-3 pb-1 hover:text-primary">
        <h2 class="flex items-baseline gap-2">
          <span class="text-sm font-semibold text-highlighted">{{ series.title }}</span>
          <span class="text-xs text-muted">{{ series.cds.length }}</span>
        </h2>
        <p v-if="series.description" class="text-xs text-muted">{{ series.description }}</p>
      </NuxtLink>

      <!-- One row per series, scrolled sideways, and it stops on a sleeve rather
           than between two: scroll-snap gives each tile a line the row comes to
           rest on, so a flick always leaves a whole CD at the left edge.
           scroll-px matches the row's own padding, or the first tile would snap
           under it. -->
      <ul
        v-if="series.cds.length"
        class="flex snap-x snap-mandatory scroll-px-3 gap-1 overflow-x-auto overscroll-x-contain px-3 pb-2"
      >
        <li
          v-for="cd in series.cds"
          :key="cd.dir"
          class="w-32 shrink-0 snap-start sm:w-36"
        >
          <NuxtLink :to="`/djvet/${series.slug}/${cd.slug}`" class="group block" :title="cd.title">
            <div class="relative aspect-square overflow-hidden rounded-sm bg-black">
              <img
                :data-src="cd.cover"
                :alt="cd.title"
                src="data:image/gif;base64,R0lGODlhAQABAAAAACH5BAEKAAEALAAAAAABAAEAAAICTAEAOw=="
                class="absolute inset-0 size-full object-cover transition duration-300 group-hover:scale-105"
                @error="blank"
              >
              <!-- A CD whose sleeve will not load still has to be findable, so
                   the title takes the tile rather than leaving a black square. -->
              <span
                class="absolute inset-0 hidden place-items-center p-2 text-center text-xs text-white group-[.no-cover]:grid"
              >{{ cd.title }}</span>
            </div>
            <p class="truncate pt-1 text-xs text-muted">{{ cd.title }}</p>
          </NuxtLink>
        </li>
      </ul>

      <p v-else class="px-3 pb-2 text-xs text-dimmed">Von dieser Reihe liegt noch nichts da.</p>
    </section>
  </article>
</template>

<script setup lang="ts">
// The DJ archive: every series as a row of its sleeves.
//
// Arranged like the magazines rather than like the record shelf, because this
// collection is arranged and that one is not: a series is a run somebody made,
// so it is a heading with its CDs beneath it, not one wall to be searched.
const { archive: load } = useDjvet()

const { data: archive, pending, error } = await useAsyncData('djvet-archive', () => load())

const cdCount = computed(() => (archive.value ?? []).reduce((n, s) => n + s.cds.length, 0))

// A sleeve that will not load marks its tile, which is what reveals the title
// underneath. Cheaper than checking eighty folders for a cover.jpeg that is
// there in every one of them.
function blank(event: Event) {
  const img = event.target as HTMLImageElement
  img.classList.add('opacity-0')
  img.closest('.group')?.classList.add('no-cover')
}

// Sleeves are fetched when they come into view, and not before.
//
// loading="lazy" is not enough here: a browser counts everything inside a
// horizontally scrolled row as near the viewport, so all eighty sleeves would be
// asked for at once. An observer judges what is actually visible, because the
// intersection it reports is clipped by the row the tile sits in.
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
    // A margin, so a sleeve is on its way by the time it is scrolled to rather
    // than starting then.
  }, { rootMargin: '300px' })

  for (const sleeve of page.value.querySelectorAll<HTMLImageElement>('img[data-src]')) {
    watching.observe(sleeve)
  }
}

onMounted(watchSleeves)
watch(archive, () => nextTick(watchSleeves))
onBeforeUnmount(() => watching?.disconnect())

// The three ways this arrives as "it did not work" are worth telling apart: one
// is a setting nobody filled in, one is a permission, and only the third is a
// fault.
const problem = computed(() => {
  switch ((error.value as { statusCode?: number } | null)?.statusCode) {
    case 501:
      return {
        what: 'Für diese Installation sind keine DJ-Sets eingerichtet.',
        why: 'Es fehlt die Einstellung DJVET_ROOT — der Ordner, in dem die Reihen liegen.',
      }
    case 403:
      return {
        what: 'Dieses Konto ist nicht für die DJ-Sets freigeschaltet.',
        why: 'Die Berechtigung heißt „djvet“.',
      }
    case 401:
      return { what: 'Bitte anmelden.', why: '' }
    default:
      return { what: 'Das Archiv ist nicht erreichbar.', why: '' }
  }
})
</script>
