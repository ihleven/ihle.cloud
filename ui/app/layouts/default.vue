<template>
  <div layout="default">
    <AppBar v-if="shell.bar" />

    <!-- A plain <main>: the minimum height below the bar is one rule, which is
         all UMain ever contributed and which the layout overrode anyway. -->
    <main
      class="flex min-h-[calc(100vh-var(--ui-header-height))] bg-ral-7004"
      :style="{ paddingBottom: reserved }"
    >
      <slot />
    </main>

    <Footer v-if="shell.footer" />

    <!-- After the content, and in this order, because that order is the whole
         of the stacking rule here: the player paints over the page and the
         menu over the player. No z-index anywhere. See main.css. -->
    <Player />

    <!-- Last, so its button is above everything and its panel above everything
         but the button.

         A page that supplies its own way out — the reader does — can turn this
         off, or the round button would sit over it. -->
    <AppMenu v-if="shell.menu" />
  </div>
</template>

<script setup lang="ts">
// The default shell: a strip at the top, the footer, and the player over
// whatever is playing. Navigation is the menu, at every width, plus the areas
// written into the top strip where there is room for them.
//
// A page keeps all of it unless it says otherwise. What it can turn off is in
// ShellParts; what it cannot is anything depending on who is signed in, because
// definePageMeta is resolved at build time — that belongs here, or in a layout
// of the module's own, as geheimtipp's footer does.
const route = useRoute()

const shell = computed(() => ({
  bar: true,
  footer: true,
  menu: true,
  ...(route.meta.shell ?? {}),
}))

// Room at the foot for the player, when something is playing. Reserved here
// rather than by each page, because the player outlives any of them.
const { current } = usePlayer()

const reserved = computed(() => (current.value ? '4rem' : undefined))
</script>
