<template>
  <article class="w-full bg-elevated">

    <ul v-if="searchresult" class="mt-12 w-full divide-y divide-accented border-y border-accented bg-default">
      <li v-for="hit in searchresult.hits" :key="hit.id" class="ml-20 flex space-x-2">

        <img :src="hit.fields.img as string" class="my-2 h-12 w-16 -translate-x-18">

        <NuxtLink :to="hit.fields.full_slug as string" class="flex -translate-x-18 p-1">

          {{ hit.fields.name }}
        </NuxtLink>
      </li>
    </ul>
  </article>
</template>

<script setup lang="ts">
const { public: { api } } = useRuntimeConfig()

const { data: searchresult } = await useFetch<SearchResult>('/api/v1/search', {
  baseURL: api.base as string,
  credentials: 'include',
  query: {
    fields: 'full_slug,name,img',
    type: 'Super8',
  },
})
</script>
