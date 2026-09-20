<template>
  <!-- The strip along the top, in the flow, so it scrolls away with the page.
       Two reasons it is not fixed. iOS blurs the top of scrolling content and
       nothing a page can write prevents it, so a bar pinned there would be
       permanently hazed — better the haze falls on ordinary content. And a
       fixed bar would need a z-index to paint over the page, which this app
       does not use: see main.css. -->
  <header
    class="flex h-(--ui-header-height) items-center justify-between gap-3 border-b border-default bg-default px-4 sm:px-6"
  >
    <!-- The name, except on a single film, where the space is better spent on
         the way back to the films — the only place you can have come from. -->
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

    <!-- Where there is room for the words. Below that the areas are in the bar
         along the foot of the screen, within reach of a thumb. -->
    <UNavigationMenu :items="items" class="hidden lg:flex" />

    <!-- At every width, not just on a phone: the menu carries signing in and
         out, so a desktop without it has no way to sign out. -->
    <UButton
      variant="ghost"
      color="neutral"
      icon="i-lucide-menu"
      aria-label="Menü öffnen"
      @click="open = true"
    />

    <!-- Rendered in a portal, so where it sits in this markup costs nothing —
         and being last in the document is what puts it over everything else
         without a z-index. -->
    <USlideover v-model:open="open" title="Menü" :ui="{ content: 'max-w-2xl' }">
      <template #body>
        <MainMenu />
      </template>
    </USlideover>
  </header>
</template>

<script setup lang="ts">
const route = useRoute()
const { items } = useAreas()

const open = ref(false)

// A film's own page, as opposed to the overview at /filme.
const film = computed(() => /^\/filme\/.+/.test(route.path))

// Any navigation closes the menu: it is a way somewhere, and staying open over
// the page it just reached would hide what it was asked for.
watch(() => route.fullPath, () => {
  open.value = false
})
</script>
