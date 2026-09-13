<template>
  <header class="flex items-center p-2">
    <NuxtLink to="/hidrive" class="flex">
      <small class="flex items-center text-gray-500 hover:text-gray-700 hover:underline">
        ihle<span class="-translate-x-px">.</span>
        <UIcon name="i-feather-cloud" class="-translate-x-0.5" />
      </small>
    </NuxtLink>

    <NuxtLink
      v-for="(slug, i) in elems"
      :key="i"
      :to="`/hidrive/${elems.slice(0, i + 1).join('/')}`"
      class="flex items-center text-gray-400"
    >
      <UIcon name="i-feather-chevron-right" />
      <small class="text-gray-500 hover:text-gray-700 hover:underline">{{ decodeURI(slug) }}</small>
    </NuxtLink>
  </header>
</template>

<script setup lang="ts">
// The trail back up. The root link is the browser itself rather than the site's
// front page, because a path here is meaningless outside it.
const props = defineProps<{ path: string }>()

const elems = computed(() => props.path.split('/').filter(Boolean))
</script>
