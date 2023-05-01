<template>
  <figure class="w-full">
    <img :src="src" :alt="alt" />
    <figcaption><slot /></figcaption>
  </figure>
</template>

<script setup>
  import { inject } from 'vue'
  const props = defineProps(['src', 'alt'])

  const basePath = inject('basePath', '/')
  const dirPath = inject('dirPath', '')
  const src = `${basePath}/${dirPath}/${props.src}`
</script>
