<template>
  <Teleport
    to="#header-title"
    defer
  >
    <slot />
  </Teleport>
</template>

<script setup lang="ts">
// What a page is called, in the header.
//
// A component rather than a bare Teleport, because the header needs to know
// whether anybody has claimed the title: where nobody has, it says what area
// this is, and the two must not be printed one after the other. Mounting claims
// it and unmounting gives it back.
//
// Still markup rather than a string, so a page can mute the part of its title
// that is a detail — an episode after a series, a count after a name.
const claimed = useState('page-title-claimed', () => false)

onMounted(() => (claimed.value = true))
onUnmounted(() => (claimed.value = false))
</script>
