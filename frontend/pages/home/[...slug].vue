<template>
  <main v-if="meta" class="bg-gray-100">
    <section>
      <NuxtLink :to="`/home`">home</NuxtLink>
      <NuxtLink
        v-for="(slug, i) in route.params.slug"
        :key="i"
        :to="`/home/${route.params.slug.slice(0, i + 1).join('/')}`"
      >
        / {{ slug }}
      </NuxtLink>
    </section>

    <section v-if="meta.category === 'image'">
      <img :src="`/api/raw${meta.path}`" />
    </section>

    <section class="divide-y divide-dashed divide-gray-300 border-y border-gray-300 bg-white">
      <nuxt-link v-for="f in meta.members" :key="f.name" :item="f" :to="`/home${f.path}`" class="flex">
        {{ f.name }}
      </nuxt-link>
    </section>

    <HidriveGallery v-if="meta.members" :images="meta.members" gallery="asdf"></HidriveGallery>
  </main>
</template>

<script setup>
  const route = useRoute()
  const path = typeof route.params.slug === 'string' ? route.params.slug : route.params.slug.join('/')
  console.log('path =>', path)
  const { data: meta } = await useFetch(`/api/meta/${path}`, {
    server: false,
    headers: { Accept: 'application/json' },
  })
  console.log('meta:', meta.value)
</script>
