<template>
  <main class="grid w-screen grid-cols-2">
    <section class="col-span-2 flex aspect-[2/1] items-center justify-center bg-ral-7035">
      <Logo />
    </section>

    <nuxt-link to="/videos" class="aspect-square bg-blue-500 bg-opacity-90 p-8">
      <h1 class="hover:text-outline text-lg font-black text-white">Videos</h1>
    </nuxt-link>

    <section class="aspect-square bg-green-500 bg-opacity-90">
      <nuxt-link v-if="dev" to="/kalender" class="block h-full w-full p-8">
        <h1 class="hover:text-outline text-lg font-black text-white">Kalender</h1>
      </nuxt-link>
    </section>

    <section class="aspect-square bg-sky-500 bg-opacity-80">
      <nuxt-link
        v-if="dev"
        to="/mediathek"
        class="hover:text-outline block h-full w-full p-8 text-lg font-extralight text-white hover:font-normal"
        >Mediathek
      </nuxt-link>
      <!-- <dl class="text-white text-lg font-extralight hover:font-normal">
        <nuxt-link :to="'/videos'" class="text-lg font-black hover:text-shadow"
          >Arbeit</nuxt-link
        >
      </dl> -->
    </section>

    <section class="aspect-square bg-cyan-300 bg-opacity-80"></section>
  </main>
</template>
<script setup>
  const env = process.env.NODE_ENV
  const dev = env === 'development'
</script>
