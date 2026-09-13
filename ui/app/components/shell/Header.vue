<template>
  <!-- The menu is normally mobile-only; here it carries signing in and out, so
       it has to be reachable at every width or there is no way to sign out on
       a desktop. -->
  <UHeader
    mode="slideover"
    :ui="{ root: 'static', toggle: 'lg:flex', content: 'lg:block max-w-2xl' }"
  >
    <!-- The name, except on a single film, where the space is better spent on
         the way back to the films — the only place you can have come from.
         The slot is always filled: left empty, UHeader falls back to its own
         Nuxt logo. -->
    <template #title>
      <UButton
        v-if="film"
        to="/filme"
        variant="ghost"
        color="neutral"
        icon="i-lucide-chevron-left"
        label="Filme"
        class="cursor-pointer"
      />
      <Wordmark v-else />
    </template>

    <UNavigationMenu :items="items" />

    <template #body>
      <MainMenu />
    </template>
  </UHeader>
</template>

<script setup lang="ts">
const route = useRoute()
const { items } = useAreas()

// A film's own page, as opposed to the overview at /filme.
const film = computed(() => /^\/filme\/.+/.test(route.path))
</script>
