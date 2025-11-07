<template>
  <main class="grid w-screen grid-cols-[repeat(auto-fit,_minmax(200px,_1fr))] content-start">
    <div class="w-full bg-lime-300 text-ral-7035">
      INDEX
      <a href="http://localhost:8000/apihle/auth/authorize?state=http://localhost:3000/">AUTHORIZE</a>

      <NuxtLink to="/cmsasdf">asdf</NuxtLink>
    </div>
    <section class="flex items-center justify-between bg-ral-7035">
      <Logo />
      <NuxtLink
        to="/login"
        class="border border-transparent p-2"
      >
        <Icon name="settings" />
      </NuxtLink>
    </section>

    <nuxt-link to="/entries" class="aspect-square bg-blue-500/80 p-8">
      <h1 class="text-lg font-black text-white hover:text-outline">CMS</h1>
    </nuxt-link>

    <section class="aspect-square bg-green-500/90">
      <nuxt-link

        to="/kalender"
        class="block h-full w-full p-8"
      >
        <h1 class="text-lg font-black text-white hover:text-outline">Kalender</h1>
      </nuxt-link>
    </section>

    <section class="aspect-square bg-sky-500/80">
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
      class="relative aspect-square bg-cyan-300/80"
    >
      <h1 class="absolute inset-8 text-lg font-black text-white hover:text-outline">Musik</h1>
    </nuxt-link>

    <div class="col-span-2 aspect-square bg-amber-300 p-4">

      <pre class="bg-amber-200 text-xs inset-shadow-amber-900">{{ session }}</pre>

      <button @click="loadSession">Load session</button>
    </div>

    <nuxt-link to="/super-8" class="block aspect-square h-full w-full bg-rose-500/80 p-8">
      <h1 class="text-lg font-black text-white hover:text-outline">Super 8</h1>
    </nuxt-link>

  </main>
</template>

<script setup lang="ts">
const { session, loadSession } = useAuth()
</script>
