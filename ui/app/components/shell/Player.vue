<template>
  <!-- One element, two densities. The audio node is never wrapped in a branch
       and never moved: re-creating it or reparenting it interrupts playback in
       several browsers, which is the whole thing this component exists to stop.
       Only the box around it and what sits beside it change with the route, and
       every conditional sibling is keyed so the diff cannot patch the audio
       element against one of them. -->
  <div
    v-if="current"
    class="fixed z-30 flex items-center gap-2 border-accented bg-default"
    :class="detailed
      ? 'inset-x-0 bottom-0 border-t px-2 py-1'
      : 'right-3 bottom-3 max-w-[22rem] rounded-full border px-3 py-1 shadow-lg'"
  >
    <img
      v-if="detailed && current.cover"
      key="cover"
      :src="current.cover"
      alt=""
      class="size-9 shrink-0 rounded-sm object-cover"
    >

    <div key="what" class="min-w-0" :class="detailed ? 'grow' : 'shrink'">
      <p class="truncate text-xs text-highlighted">{{ current.title }}</p>
      <p v-if="detailed && current.subtitle" class="truncate text-xs font-light text-dimmed">
        {{ current.subtitle }}
      </p>
    </div>

    <template v-if="detailed">
      <UButton
        key="previous"
        size="xs"
        variant="ghost"
        color="neutral"
        icon="i-lucide-skip-back"
        :disabled="!hasPrevious"
        aria-label="Vorheriger Titel"
        @click="previous"
      />
      <UButton
        key="next"
        size="xs"
        variant="ghost"
        color="neutral"
        icon="i-lucide-skip-forward"
        :disabled="!hasNext"
        aria-label="Nächster Titel"
        @click="next"
      />
    </template>

    <!-- The browser's own transport. It is keyboard-reachable, it labels itself
         in the visitor's language and it already knows how to scrub, none of
         which a hand-drawn set of buttons would get for free. -->
    <audio
      key="audio"
      :src="current.src"
      controls
      autoplay
      preload="none"
      class="h-8"
      :class="detailed ? 'max-w-md grow' : 'w-40 shrink-0'"
      @ended="next"
    />

    <UButton
      key="close"
      size="xs"
      variant="ghost"
      color="neutral"
      icon="i-lucide-x"
      aria-label="Wiedergabe beenden"
      @click="stop"
    />
  </div>
</template>

<script setup lang="ts">
// The player, mounted once in app.vue and never by a page.
//
// Which shape it takes is a question about where the visitor is, not about what
// is playing: in the music shelf and the DJ archive the player is part of what
// the page is for, so it gets the full bar. Anywhere else it is something the
// visitor started earlier and has not asked to see, so it shrinks to a pill —
// still visible, and still one click from silence, because audio you can hear
// and cannot readily stop is worse than audio you can see.
const { current, hasNext, hasPrevious, next, previous, stop } = usePlayer()

const route = useRoute()
const detailed = computed(() => /^\/(musik|djvet)(\/|$)/.test(route.path))
</script>
