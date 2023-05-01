<template>
  <main class="">
    <NuxtLink :to="`/home`">home</NuxtLink>

    <hr />

    <!-- <NuxtLink :to="`/home/${route.params.slug.slice(0,i+1).join('/')}`" v-for="(slug, i) in route.params.slug" :key="i"> / {{slug}}</NuxtLink> -->

    <ul v-if="meta.type === 'dir'" class="divide-y divide-dashed divide-gray-300">
      <div v-for="f in meta.members" :key="f.id" :item="f" :to="`/home${f.path}`">
        {{ f.path }}
      </div>
    </ul>
    <div v-else-if="meta.mime_type === 'image/jpeg'">
      <img :src="`http://localhost:8000/serve${meta.path}`" />
    </div>
  </main>
</template>

<script setup>
  console.log('meta')

  const route = useRoute()
  const path = typeof route.params.slug === 'string' ? route.params.slug : route.params.slug.join('/')
  console.log('meta:', `/hidrive/${path}`)
  const { data: meta } = await useFetch(`/api/hidrive/${path}`)
  console.log('meta:', meta.value?.path)
</script>
