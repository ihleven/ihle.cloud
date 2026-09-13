<template>
  <video
    autoplay
    playsinline
    muted
    controls
    crossorigin="use-credentials"
    class="mx-auto max-h-screen w-full"
  >
    <source :src="media(path)" :type="type">
  </video>
</template>

<script setup lang="ts">
// The old browser's player, reduced to the part that belongs to a file browser.
//
// Its chapter list and transport were wired to the super-8 pages — a back link
// to /videos, a hardcoded chapter label — and the app has a film player of its
// own for that. What is left is what this page needs: play the file that was
// clicked on.
//
// It takes the path the page was asked for, not meta.path. The provider returns
// that one percent-encoded and resolved against the drive's root, so building a
// URL from it encodes the escapes a second time and prefixes a root the browser
// never sees.
defineProps<{
  path: string
  type?: string
}>()

const { media } = useDrive()
</script>
