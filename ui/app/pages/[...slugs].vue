<template>
  <article class="h-full min-h-full w-full">
    <!-- <header><NuxtLink :to="`/entries/${entry?.path}`">{{ entry?.id }}</NuxtLink></header> -->
    <component
      :is="component"
      v-model:entry="entry"
      class="z-0 grow"
    />
    <!-- <pre>
      {{ entry }}
    </pre> -->
  </article>
</template>

<script setup lang="ts">
definePageMeta({
  middleware: [
    function (to) {
      const { session } = useAuth()
      if (!session.value?.aud.includes('famihli')) {
        console.log('access not allowed', session.value)
        // return navigateTo('/login?redirect=' + encodeURIComponent(to.fullPath) + '&reason=noperm&aud=famihlie')
      }
    },
  ],
})
// definePageMeta({ layout: false })

const { params: { slugs } } = useRoute()
const full_slug = slugs ? typeof slugs === 'string' ? slugs : slugs.join('/') : ''

// const breadcrumb = ref<{ label: string }[] | null>([{ label: 'home' }, ...full_slug.split('/').map(slug => ({ label: slug }))])

const { data: entry, error } = await useFetch<Entry>(`/api/v1/entry`, {
  baseURL: useRuntimeConfig().public.api.base,
  credentials: 'include',
  query: {
    slugs: full_slug,
  },
})

if (error.value) {
  throw createError({ statusCode: 404, message: 'not find' })
}

const componentmap: Record<string, object | string> = {
  Super8: resolveComponent('PageSuper8'),
  Ausstellung: resolveComponent('PageExhibition'),
  Person: resolveComponent('PagePerson'),
  Default: resolveComponent('PageExhibition'),
}

const component = computed(() => entry.value ? componentmap[(entry.value as Entry).type] : componentmap.Default)
</script>
