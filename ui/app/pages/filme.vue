<template>
  <article class="flex min-h-full w-full flex-col bg-elevated">

    <header class="flex shrink-0 justify-end gap-2 p-2">
      <UButton
        size="xs" variant="ghost" color="neutral"
        :icon="wall === 'black' ? 'i-lucide-sun' : 'i-lucide-moon'"
        :title="wall === 'black' ? 'Heller Hintergrund' : 'Dunkler Hintergrund'"
        @click="wall = wall === 'black' ? 'white' : 'black'"
      />
      <UFieldGroup size="xs">
        <UButton
          v-for="mode in modes" :key="mode.id"
          :variant="layout === mode.id ? 'solid' : 'ghost'" color="neutral"
          :icon="mode.icon" :title="mode.title"
          @click="layout = mode.id"
        />
      </UFieldGroup>
    </header>

    <ul v-if="layout === 'list' && searchresult" class="w-full divide-y divide-accented border-y border-accented bg-default">
      <li v-for="hit in searchresult.hits" :key="hit.id" class="ml-20 flex space-x-2">

        <!-- The box is 4:3 and the still is 16:9, so it has to be told how to
             fit; without this the picture is stretched. -->
        <img :src="hit.fields.img as string" class="my-2 h-12 w-16 -translate-x-18 bg-black object-cover">

        <!-- Absolute: a full slug is a path from the root, and left relative it
             would resolve against whatever page is listing it. -->
        <NuxtLink :to="`/${hit.fields.full_slug}`" class="flex -translate-x-18 p-1">

          {{ hit.fields.name }}
        </NuxtLink>
      </li>
    </ul>

    <!-- A wall of stills, filling the page. The background shows through as the
         outer padding and the gap between films, both the same size, so the
         frame around the wall and the lines within it are one surface. Rows stay
         at the top rather than stretching to fill the height.
         The hairline top and bottom sets the wall off against the header and the
         footer without becoming a band of its own. -->
    <div
      v-else-if="searchresult"
      class="grid grow content-start grid-cols-3 gap-1 border-y border-black p-1 sm:grid-cols-5 md:grid-cols-6 lg:grid-cols-8 xl:grid-cols-10"
      :class="wall === 'black' ? 'bg-black' : 'bg-white'"
    >
      <NuxtLink
        v-for="hit in searchresult.hits"
        :key="hit.id"
        :to="`/${hit.fields.full_slug}`"
        class="group relative block overflow-hidden bg-black"
        :class="layout === 'ratio' ? 'aspect-4/3' : 'aspect-square'"
      >
        <img
          :src="poster(hit)"
          :alt="hit.fields.name as string"
          class="absolute inset-0 size-full transition duration-300 group-hover:scale-105"
          :class="layout === 'fit' ? 'object-contain' : 'object-cover'"
        >
        <!-- The title sits on the still. The gradient is there so a name over a
             bright frame stays readable without a panel covering the picture. -->
        <div class="absolute inset-x-0 bottom-0 bg-linear-to-t from-black/80 via-black/40 to-transparent px-1.5 pb-1 pt-6">
          <h2 class="truncate text-xs font-medium text-white drop-shadow">{{ hit.fields.name }}</h2>
        </div>
      </NuxtLink>
    </div>
  </article>
</template>

<script setup lang="ts">
const { public: { api } } = useRuntimeConfig()

// 4:3 matches the box the list previews use. The two square modes differ only in
// what happens to a 16:9 still that does not fit one: filled crops the sides,
// embedded keeps the whole frame and lets the tile show through above and below.
const modes = [
  { id: 'list', icon: 'i-lucide-list', title: 'Liste' },
  { id: 'ratio', icon: 'i-lucide-rectangle-horizontal', title: '4:3' },
  { id: 'fill', icon: 'i-lucide-square', title: 'Quadratisch, gefüllt' },
  { id: 'fit', icon: 'i-lucide-scan', title: 'Quadratisch, eingepasst' },
] as const

type Layout = typeof modes[number]['id']

// Remembered across navigations, so going into a film and back does not throw
// away the choice.
const layout = useState<Layout>('film-layout', () => 'fit')

// The wall's own colour. The tiles stay black regardless, so a contained still
// keeps its bars and a tile keeps an edge even on a white wall.
const wall = useState<'white' | 'black'>('film-wall', () => 'white')

const { data: searchresult } = await useFetch<SearchResult>('/api/v1/search', {
  baseURL: api.base as string,
  credentials: 'include',
  query: {
    fields: 'full_slug,name,img',
    type: 'Film',
  },
})

// One size for every mode, so switching between them does not refetch. 400 is
// twice the widest tile, which is what a high-density display wants.
function poster(hit: { fields: Record<string, unknown> }) {
  const src = hit.fields.img as string
  return src ? `${src}?width=400&height=400` : ''
}
</script>
