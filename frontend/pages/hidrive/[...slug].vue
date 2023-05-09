<template>
  <main class="min-h-screen bg-gray-100">
    <NuxtLink :to="`/hidrive`" class="p-8">hidrive</NuxtLink>

    <HidriveGallery v-if="meta.members" :images="meta.members" gallery="asdf"></HidriveGallery>
  </main>
</template>

<script setup>
  // import PhotoSwipeLightbox from 'photoswipe/lightbox'
  // import 'photoswipe/style.css'

  const route = useRoute()
  const path = typeof route.params.slug === 'string' ? route.params.slug : route.params.slug.join('/')
  console.log('meta:', `/hidrive/${path}`)

  const { data: meta } = await useFetch(`/api/hidrive/${path}`)
  // const { data: meta } = await useFetch(`/api/hidrive/${path}`)
  // console.log('meta:', meta.value?.path)
  // const lightbox = ref(null)
  // onMounted(() => {
  //   if (!lightbox.value) {
  //     lightbox.value = new PhotoSwipeLightbox({
  //       gallery: '#galleryID', // + this.$props.galleryID,
  //       children: 'a',
  //       pswpModule: () => import('photoswipe'),
  //     })
  //     lightbox.value.init()
  //   }
  // })
</script>
