<template>
  <main class="min-h-screen bg-neutral-200">
    <NavigationBar />

    <section class="mx-auto h-full max-w-screen-md bg-neutral-200 pb-16 pt-8">
      <h2 class="text-md px-4 py-1 font-bold text-neutral-600">Videos:</h2>
      <ul class="divide-y divide-dashed divide-neutral-300 border-y border-neutral-300 bg-neutral-100">
        <nuxt-link
          v-for="video in videos"
          :key="video.slug"
          :to="`/videos/${video.slug}`"
          class="flex cursor-pointer items-center justify-between px-2 py-2"
        >
          <h3 class="text-sm font-semibold text-neutral-700">
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
  // const { params } = useRoute()
  const { videos } = useSuper8()
</script>
