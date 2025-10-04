<template>
  <figure class="w-1/2 float-right ml-4">
    <MDCSlot unwrap="p" />
    <figcaption class="text-center">
      <MDCSlot name="caption" unwrap="p" />
    </figcaption>
  </figure>
</template>

<script setup>
// import { inject } from 'vue'

// const props = defineProps(['src', 'alt'])

// const basePath = inject('basePath', '/')
// const dirPath = inject('dirPath', '')
// const src = props.src.startsWith('/') ? `${basePath}${props.src}` : `${basePath}/${dirPath}/${props.src}`
</script>
