<template>
  <main class="grid w-screen grid-cols-2">
    <!-- <NavigationBar class="col-span-2" /> -->

    <nuxt-link v-for="album in alben" :key="album" :to="'/musik/' + album" class="relative bg-slate-100">
      <img :src="`/api/raw/public/alben/${album}/cover.jpeg`" class="object-cover" />
      <!-- <h1 class="hover:text-outline absolute inset-0 text-lg font-black text-white">{{ album }}</h1> -->
    </nuxt-link>
  </main>
</template>

<script setup>
  const alben = ['PoolParty12', 'Trash', 'RX Reloaded']

  // const { params } = useRoute()

  // const prefix = 'http://localhost:10815/proxy/public/mediathek/'

  // const { data: dir } = await useFetch(`/api/meta/public/alben/PoolParty12`, { server: false })
  // console.log('dir', dir)
</script>
