<template>
  <div
    :id="gallery"
    class="grid grid-cols-[repeat(auto-fill,_minmax(128px,_1fr))] gap-1 bg-white p-1"
    @keyup="next"
  >
    <a
      v-for="f in files"
      :key="f.path"
      :href="media(join(dir, driveName(f)))"
      :data-pswp-width="f.image!.width"
      :data-pswp-height="f.image!.height"
      data-cropped="true"
      target="_blank"
      rel="noreferrer"
      class="aspect-square"
    >
      <img :src="thumb(join(dir, driveName(f)), 200)" :alt="driveName(f)" class="h-full w-full object-cover">
    </a>
  </div>
</template>

<script setup lang="ts">
import PhotoSwipeLightbox from 'photoswipe/lightbox'
import 'photoswipe/style.css'

// Named for the tag rather than for the folder: this app registers components
// with pathPrefix:false, so the directory contributes nothing to the name and a
// bare <Gallery> would be a collision waiting to happen.
const props = defineProps<{
  images: DriveMeta[]
  /** The element id the lightbox binds to; it needs one to find its children. */
  gallery: string
  dir: string
}>()

const { media, thumb } = useDrive()

// Only those the store measured: the lightbox needs real dimensions to size a
// slide, and guessing them makes it open at the wrong scale.
const files = computed(() => props.images.filter(i => i.image))

function join(dir: string, name: string): string {
  return [dir, name].filter(Boolean).join('/')
}

const lightbox = ref<PhotoSwipeLightbox | null>(null)

onMounted(() => {
  lightbox.value = new PhotoSwipeLightbox({
    gallery: '#' + props.gallery,
    children: 'a',
    pswpModule: () => import('photoswipe'),
  })
  lightbox.value.init()
})

onUnmounted(() => {
  lightbox.value?.destroy()
  lightbox.value = null
})

function next(event: KeyboardEvent) {
  if (event.key === 'Meta') lightbox.value?.pswp?.next()
  if (event.key === 'Shift') lightbox.value?.pswp?.prev()
}
</script>
