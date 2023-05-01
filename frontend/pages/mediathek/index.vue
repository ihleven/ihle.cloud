<template>
  <main class="w-screen grid grid-cols-2">
    <NavigationBar /><i></i>
    <nuxt-link to="/mediathek/Die-Erben-der-Saurier/1.1-Eine-neue-Zeit">
      <article
        class="p-8 aspect-square text-white bg-sky-300 text-lg font-black hover:text-outline bg-[url('/api/proxy/public/mediathek/Die-Erben-der-Saurier/cover.jpg')] bg-cover bg-center bg-no-repeat"
      >
        Die Erben der Saurier
      </article>
    </nuxt-link>

    <section class="bg-green-500 bg-opacity-90 aspect-square">
      <nuxt-link v-if="dev" to="/kalender" class="block w-full h-full p-8">
        <h1 class="text-white text-lg font-black hover:text-outline">Kalender</h1>
      </nuxt-link>
    </section>

    <section class="bg-sky-500 bg-opacity-80 aspect-square">
      <nuxt-link v-if="dev" to="/mediathek" class="block w-full h-full p-8 text-white text-lg font-extralight hover:font-normal hover:text-outline"
        >Mediathek
      </nuxt-link>
      <!-- <dl class="text-white text-lg font-extralight hover:font-normal">
        <nuxt-link :to="'/videos'" class="text-lg font-black hover:text-shadow"
          >Arbeit</nuxt-link
        >
      </dl> -->
    </section>

    <section class="bg-cyan-300 bg-opacity-80 aspect-square"></section>
  </main>
</template>

<script setup></script>
