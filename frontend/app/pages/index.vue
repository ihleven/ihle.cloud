<template>
  <main class="grid w-screen grid-cols-2">
    <div class="h-96 w-full text-ral-7035">
      test
      <a href="http://localhost:8000/apihle/auth/authorize?state=http://localhost:3000/">AUTHORIZE</a>

      <NuxtLink to="/cmsasdf">asdf</NuxtLink>
    </div>
    <section class="col-span-2 flex items-center justify-between bg-ral-7035">
      <Logo />
      <NuxtLink
        to="/login"
        class="border border-transparent p-2"
      >
        <Icon name="settings" />
      </NuxtLink>
    </section>

    <nuxt-link
      to="/videos"
      class="bg-opacity-90 aspect-square bg-blue-500 p-8"
    >
      <h1 class="text-lg font-black text-white hover:text-outline">Videos</h1>
    </nuxt-link>

    <section class="bg-opacity-90 aspect-square bg-green-500">
      <nuxt-link
        v-if="dev"
        to="/kalender"
        class="block h-full w-full p-8"
      >
        <h1 class="text-lg font-black text-white hover:text-outline">Kalender</h1>
      </nuxt-link>
    </section>

    <section class="bg-opacity-80 aspect-square bg-sky-500">
      <nuxt-link
        v-if="true"
        to="/mediathek"
        class="block h-full w-full p-8 text-lg font-extralight text-white hover:font-normal hover:text-outline"
      >Mediathek
      </nuxt-link>
      <!-- <dl class="text-white text-lg font-extralight hover:font-normal">
        <nuxt-link :to="'/videos'" class="text-lg font-black hover:text-shadow"
          >Arbeit</nuxt-link
        >
      </dl> -->
    </section>

    <nuxt-link
      to="/musik"
      class="bg-opacity-80 relative aspect-square bg-cyan-300"
    >
      <h1 class="absolute inset-8 text-lg font-black text-white hover:text-outline">Musik</h1>
    </nuxt-link>
  </main>
</template>

<script setup>
const env = process.env.NODE_ENV
const dev = env === 'development'
</script>
