<template>
  <article class="flex min-h-full w-full flex-col bg-elevated">
    <header class="flex shrink-0 items-baseline gap-2 px-3 py-2">
      <h1 class="text-sm font-semibold text-highlighted">Mediathek</h1>
      <span v-if="shelf" class="text-xs text-muted">{{ shelf.length }} Serien</span>
    </header>

    <p v-if="pending" class="px-3 py-8 text-sm text-muted">Wird geladen…</p>

    <div v-else-if="error" class="px-3 py-8 text-sm">
      <p class="text-error">{{ problem.what }}</p>
      <p v-if="problem.why" class="mt-1 text-muted">{{ problem.why }}</p>
    </div>

    <p v-else-if="!shelf?.length" class="px-3 py-8 text-sm text-muted">
      Hier steht noch nichts.
    </p>

    <!-- Fewer columns than the film wall: eight series with portrait covers,
         not two hundred stills, so the tiles can be large enough to read. -->
    <div
      v-else
      class="grid grow grid-cols-2 content-start gap-1 border-y border-black bg-black p-1 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6"
    >
      <NuxtLink
        v-for="series in shelf"
        :key="series.dir"
        :to="`/mediathek/${encodeURIComponent(series.dir)}`"
        class="group relative block aspect-2/3 overflow-hidden bg-black"
      >
        <img
          :src="series.cover"
          :alt="series.title"
          loading="lazy"
          class="absolute inset-0 size-full object-contain transition duration-300 group-hover:scale-105"
        >
        <div class="absolute inset-x-0 bottom-0 bg-linear-to-t from-black/85 via-black/50 to-transparent px-2 pt-8 pb-1.5">
          <h2 class="truncate text-sm font-medium text-white drop-shadow">{{ series.title }}</h2>
          <p class="text-xs text-white/70">
            <span v-if="series.year">{{ series.year }}</span>
            <span v-if="series.year && series.episodes.length"> · </span>
            <span v-if="series.episodes.length">{{ series.episodes.length }} {{ series.episodes.length === 1 ? 'Folge' : 'Folgen' }}</span>
          </p>
        </div>
      </NuxtLink>
    </div>
  </article>
</template>

<script setup lang="ts">
// The shelf, as a wall of covers — the same gesture as the film wall, and the
// same one the old Mediathek made before it: a grid of artwork you pick from by
// looking rather than reading.
//
// What is behind each tile is a folder on the family storage. Which folders
// count as series, what to call them and how many episodes they hold are all
// read out of the listing; see useMediathek for what each of those readings
// assumes and where it is already known to be wrong.
const { shelf: load } = useMediathek()

const { data: shelf, pending, error } = await useAsyncData('mediathek-shelf', load)

// Three quite different situations reach this page as "it did not work", and
// saying so is no help to whoever has to fix it: one is a setting nobody filled
// in, one is a permission, and only the third is a fault. The status says which.
const problem = computed(() => {
  switch ((error.value as { statusCode?: number } | null)?.statusCode) {
    case 501:
      return {
        what: 'Für diese Installation ist keine Mediathek eingerichtet.',
        why: 'Es fehlt die Einstellung MEDIATHEK_ROOT — der Ordner, in dem die Serien liegen.',
      }
    case 403:
      return {
        what: 'Dieses Konto ist nicht für die Mediathek freigeschaltet.',
        why: 'Die Berechtigung heißt „mediathek“.',
      }
    case 401:
      return { what: 'Bitte anmelden.', why: '' }
    default:
      return { what: 'Die Mediathek ist nicht erreichbar.', why: '' }
  }
})
</script>
