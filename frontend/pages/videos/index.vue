<template>
  <main class="bg-neutral-200 min-h-screen">
    <NavigationBar />

    <section class="bg-neutral-200 max-w-screen-md mx-auto pt-8 pb-16 h-full">
      <h2 class="px-4 py-1 text-md font-bold text-neutral-600">Videos:</h2>
      <ul class="bg-neutral-100 border-y border-neutral-300 divide-y divide-neutral-300 divide-dashed">
        <nuxt-link v-for="video in videos" :key="video.slug" :to="`/videos/${video.slug}`" class="flex justify-between items-center cursor-pointer px-2 py-2">
          <h3 class="text-neutral-700 text-sm font-semibold">
            {{ video.desc }}
            <!-- <small class="text-neutral-300 text-xs font-medium">{{
              video.desc
            }}</small> -->

            <!-- <p class="text-neutral-500 text-sm font-light">{{ video.slug }}</p> -->
          </h3>
          <svg class="h-6 w-6 text-neutral-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
          </svg>
        </nuxt-link>
      </ul>
    </section>
  </main>
</template>

<script setup>
const { params } = useRoute();
const { videos } = useSuper8();
</script>
