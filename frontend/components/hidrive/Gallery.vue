<template>
  <div :id="gallery" class="grid grid-cols-[repeat(auto-fill,_minmax(128px,_1fr))] gap-1 bg-white p-1" @keyup="next">
    <a
      v-for="f in files"
      :key="f.id"
      :href="'/api/raw' + f.path"
      :data-pswp-width="f.image.width"
      :data-pswp-height="f.image.height"
      data-cropped="true"
      target="_blank"
      rel="noreferrer"
      class="aspect-square"
    >
      <img :src="'/api/thumbs?path=' + f.path + '&width=200'" :alt="f.name" class="h-full w-full object-cover" />
    </a>
  </div>
  <!-- <Keypress key-event="keyup" :key-code="13" @success="next" /> -->
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

  function next(event) {
    switch (event.key) {
      case 'Meta':
        lightbox.value.pswp.next()
        break
      case 'Shift':
        lightbox.value.pswp.prev()
        break
      default:
        break
    }
  }
</script>
