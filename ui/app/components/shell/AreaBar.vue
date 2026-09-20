<template>
  <!-- The areas along the foot of the screen, where a thumb reaches and where
       iOS does not blur. Narrow screens only: from lg the same list is in the
       strip along the top, which has the room for words.

       Fixed, and rendered after the page, which is what puts it over the
       content without a z-index. The menu is rendered after this again, so it
       covers this in turn. See main.css. -->
  <nav
    v-if="shown.length"
    class="fixed inset-x-0 bottom-0 flex items-stretch border-t border-default bg-default pb-[env(safe-area-inset-bottom)] lg:hidden"
  >
    <NuxtLink
      v-for="area in shown"
      :key="area.id"
      :to="area.at"
      class="flex grow basis-0 flex-col items-center gap-0.5 px-1 py-1.5 text-muted"
      active-class="text-primary"
    >
      <UIcon :name="area.icon!" class="size-5 shrink-0" />
      <span class="w-full truncate text-center text-[10px]">{{ area.label }}</span>
    </NuxtLink>

    <!-- Always, even when everything fits: it carries signing out, and the rest
         of the areas when there are more than there is room for. -->
    <button
      type="button"
      class="flex grow basis-0 cursor-pointer flex-col items-center gap-0.5 px-1 py-1.5 text-muted"
      :aria-label="rest.length ? `Menü, und ${rest.length} weitere Bereiche` : 'Menü'"
      @click="open = true"
    >
      <UIcon name="i-lucide-menu" class="size-5 shrink-0" />
      <span class="w-full truncate text-center text-[10px]">Mehr</span>
    </button>

    <USlideover v-model:open="open" title="Menü" :ui="{ content: 'max-w-2xl' }">
      <template #body>
        <MainMenu />
      </template>
    </USlideover>
  </nav>
</template>

<script setup lang="ts">
// How many areas the bar shows before the rest go to the menu. Four and the
// menu button is five across, which is what fits a phone without the labels
// turning into two characters and an ellipsis.
const room = 4

const route = useRoute()
const { areas } = useAreas()

const open = ref(false)

const split = computed(() => splitForBar(areas.value, room))
const shown = computed(() => split.value.shown)
const rest = computed(() => split.value.rest)

watch(() => route.fullPath, () => {
  open.value = false
})
</script>
