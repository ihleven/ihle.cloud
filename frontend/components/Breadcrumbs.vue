<template>
  <header class="flex items-center p-2">
    <NuxtLink to="/" class="flex">
      <small class="flex items-center text-gray-500 hover:text-gray-700 hover:underline"
        >ihle<span class="-translate-x-px">.</span><Icon name="feather:cloud" class="-translate-x-0.5"
      /></small>

      <!-- <svg
        viewBox="0 0 24 24"
        fill="white"
        stroke="currentColor"
        stroke-width="1"
        stroke-linecap="round"
        stroke-linejoin="round"
        class="mr-1 h-6 w-6"
      >
        <path d="M18 10h-1.26A8 8 0 1 0 9 20h9a5 5 0 0 0 0-10z"></path>
      </svg> -->
    </NuxtLink>

    <NuxtLink
      v-for="(slug, i) in elems"
      :key="i"
      :to="`/${elems.slice(0, i + 1).join('/')}`"
      class="flex items-center text-gray-400"
    >
      <!-- <nuxt-icon name="chevron-right" :filled="true" class="m-1 h-4 w-4" /> -->
      <Icon name="feather:chevron-right" />
      <small class="text-gray-500 hover:text-gray-700 hover:underline">{{ slug }}</small>
    </NuxtLink>
  </header>
</template>

<script setup>
  const props = defineProps({ path: String })

  const path = props.path.startsWith('/') ? props.path.slice(1) : props.path

  const elems = path.split('/')
</script>

<style>
  .nuxt-icon svg {
    width: unset;
    height: unset;
    margin-bottom: unset;
    vertical-align: unset;
  }
</style>
