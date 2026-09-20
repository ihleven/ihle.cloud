<template>
  <article class="h-full min-h-full w-full">
    <component
      :is="component"
      v-model:entry="entry"
      class="grow"
    />
  </article>
</template>

<script setup lang="ts">
definePageMeta({
  // A content entry is shown full-bleed and carries its own way around, so the
  // round menu button would sit on top of it.
  shell: { menu: false },
  middleware: [
    function () {
      // The audience a session was issued for. Only this app issues one, so a
      // mismatch means a session from somewhere else sharing the cookie —
      // worth saying, but not worth refusing over.
      const { session } = useAuth()
      if (session.value && !session.value.aud.includes('famihlie')) {
        console.warn('session is for another audience:', session.value.aud)
      }
    },
  ],
})

const { params: { slugs } } = useRoute()
const full_slug = slugs ? typeof slugs === 'string' ? slugs : slugs.join('/') : ''

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
  Film: resolveComponent('PageFilm'),
  Ausstellung: resolveComponent('PageExhibition'),
  Person: resolveComponent('PagePerson'),
  Work: resolveComponent('PageWork'),
  Default: resolveComponent('PageExhibition'),
}

const component = computed(() => entry.value ? componentmap[(entry.value as Entry).type] : componentmap.Default)
</script>
