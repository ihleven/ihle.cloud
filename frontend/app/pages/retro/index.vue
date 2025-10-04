<template>
  <main class="w-screen">
    <section class="grid grid-cols-4 bg-cyan-300">
      <a
        v-for="m in meta?.members"
        :key="m.id"
        :href="`http://localhost:8000/hi/media/${meta.path}/${m.name}`"
        class="aspect-square"
      >
        <article
          class="hover:text-outline bg-sky-300 bg-cover bg-center bg-no-repeat p-8 text-lg font-black text-white"
          :class="`bg-[url('http://localhost:8000/hi/media/thumbs/${meta.path}/${m.name}?width=300')]`"
        >
          <img :src="`http://localhost:8000/hi/media/thumbs/${meta.path}/${m.name}?width=300`">
        </article>
      </a>
    </section>
  </main>
</template>

<script setup>
const config = useRuntimeConfig()

const { data: meta } = await useFetch(`${config.public.hidrive.api}/meta/public/zeitschriften/HappyComputer/`, {
  credentials: 'include',
  headers: { Accept: 'application/json' },
  server: false,
})
console.log('meta:', meta.value)
</script>
