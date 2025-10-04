<template>
  <main class="bg-gray-100">
    <section>
      <NuxtLink :to="`/home`">home</NuxtLink>
      <!-- <NuxtLink v-for="(slug, i) in params.slug" :key="i" :to="`/home/${params.slug.slice(0, i + 1).join('/')}/`">
        / {{ slug }}
      </NuxtLink> -->
    </section>

    <section v-if="meta?.category === 'image'">
      <img :src="`/api/raw${meta.path}`">
    </section>

    <section class="divide-y divide-dashed divide-gray-300 border-y border-gray-300 bg-white">
      <nuxt-link
        v-for="f in meta?.members"
        :key="f.name"
        :item="f"
        :to="`/home${[meta.path, f.name].join('/')}${f.type === 'dir' ? '/' : ''}`"
        class="flex"
      >
        {{ f.name }}
      </nuxt-link>
    </section>

    <HidriveGallery v-if="meta?.members" :images="meta.members" gallery="asdf" />
  </main>
</template>

<script setup>
const { metaURL } = useHidrive()
const url = metaURL()

const { data: meta } = await useFetch(url, {
  credentials: 'include',
  headers: { Accept: 'application/json' },
  server: false,
})
console.log('meta:', meta.value)
</script>
