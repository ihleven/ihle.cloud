<template>
  <div layout="default">
    <Header />

    <!-- Room at the bottom while something is playing: the player is fixed, so
         without it the last row of any page sits underneath the bar. Reserved
         here rather than by each page, because the player outlives all of
         them. -->
    <UMain class="z-0 flex h-full bg-ral-7004" :class="{ 'pb-16': playing }">
      <slot />
    </UMain>

    <Footer />
  </div>
</template>

<script setup lang="ts">
const { current } = usePlayer()
const playing = computed(() => !!current.value)
</script>
