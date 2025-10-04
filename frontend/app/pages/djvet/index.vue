<template>
  <main class="mx-auto grid max-w-screen-md grid-cols-[repeat(auto-fill,_minmax(192px,_1fr))] gap-1 bg-white p-1">
    <nuxt-link
      v-for="s in series"
      :key="s.slug"
      :to="pathJoin([path, s.slug])"
      class="aspect-square bg-blue-500 bg-opacity-90 p-8"
    >
      <h1 class="hover:text-shadow text-lg font-black text-white">{{ s.name }}</h1>

      <!-- <img
        :src="'http://localhost:8000/hi/media/thumbs?path=' + dir + '/' + f.name + '&width=200'"
        :alt="f.name"
        class="h-full w-full object-cover"
      /> -->
    </nuxt-link>
  </main>
</template>

<script setup>
  const { path } = useRoute()
  const { pathJoin } = useHidrive()

  const series = [
    { slug: 'pool-party', name: 'Pool Party' },
    { slug: 'lunar-suite', name: 'Lunar Suite' },
    { slug: 'rxc', name: 'RXC' },
    { slug: 'lounge-club', name: 'Lounge Club' },
    { slug: 'rudi-bar', name: 'Rudi Bar' },
    { slug: 'ambient-meets-acid', name: 'Ambient meets Acid' },
  ]
</script>
