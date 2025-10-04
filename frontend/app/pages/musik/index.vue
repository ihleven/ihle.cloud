<template>
  <main class="grid w-screen grid-cols-2">

    <nuxt-link v-for="album in alben" :key="album.id" :to="'/musik/' + album.name" class="relative bg-slate-100">
      <img :src="media(`public/alben/${album.name}/cover.jpeg`)" class="object-cover" />
      <!-- <h1 class="hover:text-outline absolute inset-0 text-lg font-black text-white">{{ album }}</h1> -->
    </nuxt-link>
  </main>
</template>

<script setup>
const {media} = useHidrive()







  const config = useRuntimeConfig()
  const { data } = await useFetch(`${config.public.hidrive.api}/meta/public/alben/`, {
    server: false,
    credentials: 'include',
    headers: { Accept: 'application/json' },
  })
  const alben = computed(()=> data.value? data.value.members.filter(m => m.category === 'directory') : [])

</script>
