<template>
  <div
    :id="gallery"
    class="grid grid-cols-[repeat(auto-fit,_minmax(100px,_1fr))] gap-1 border-b border-t border-gray-300 bg-white p-1"
  >
    <a
      v-for="(image, key) in files"
      :key="image.id"
      :href="'/api/serve' + image.path"
      :data-pswp-width="image.image.width"
      :data-pswp-height="image.image.height"
      data-cropped="true"
      target="_blank"
      rel="noreferrer"
      class="= aspect-square"
    >
      <img :src="'/api/thumbs' + image.path + '?width=200'" :alt="image.name" class="h-full w-full object-cover" />
    </a>
  </div>
</template>

<script setup>
  import PhotoSwipeLightbox from 'photoswipe/lightbox'
  import 'photoswipe/style.css'

  const props = defineProps(['images', 'gallery'])

  const lightbox = ref(null)
  const files = computed(() => props.images.filter(i => i.image))

  onMounted(() => {
    if (!lightbox.value) {
      lightbox.value = new PhotoSwipeLightbox({
        gallery: '#' + props.gallery,
        children: 'a',
        pswpModule: () => import('photoswipe'),
      })
      lightbox.value.init()
    }
  })

  onUnmounted(() => {
    if (lightbox.value) {
      lightbox.value.destroy()
      lightbox.value = null
    }
  })
</script>
