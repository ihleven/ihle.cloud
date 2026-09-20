<template>
  <article class="min-h-full w-full bg-elevated">
    <!-- The way back, kept on screen. Installed, this is the only one there
         is: no address bar, no back button, no pull to reload. -->
    <header
      class="sticky top-0 flex items-center gap-2 border-b border-accented bg-default px-2 py-1 text-highlighted"
    >
      <UButton
        to="/retro"
        size="xs"
        variant="ghost"
        color="neutral"
        icon="i-lucide-chevron-left"
        label="Zeitschriften"
      />
      <span class="min-w-0 grow truncate text-xs">{{ title }}</span>
    </header>

    <!-- A page of the app rather than the whole screen, so the reel is given
         the same measure as everything else here. -->
    <div class="mx-auto w-full max-w-4xl p-2">
      <RetroReader :path="path" />
    </div>
  </article>
</template>

<script setup lang="ts">
// An issue at an address of its own.
//
// The shelf opens one over itself rather than coming here, so this exists for
// the links that are kept and shared — which is also why it is a page with the
// app's chrome rather than the sheet the shelf slides up.
// The reader has a header of its own, carrying the way back to the shelf, and
// a magazine is better read without a second bar above it.
definePageMeta({ shell: { bar: false, footer: false } })

const route = useRoute()

const path = computed(() => {
  const p = route.params.path

  return Array.isArray(p) ? p.join('/') : String(p ?? '')
})

const title = computed(() => decodeURIComponent(lastSegment(path.value)).replace(/\.pdf$/i, ''))
</script>
