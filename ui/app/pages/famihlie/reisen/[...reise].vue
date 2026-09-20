<template>
  <div class="mx-auto max-w-3xl p-6">
    <Reise
      v-if="entry"
      :reise="entry.content"
    />
    <p
      v-else-if="error"
      class="text-muted"
    >
      Diese Reise gibt es nicht.
    </p>
  </div>
</template>

<script setup lang="ts">
// Under the familie area on purpose: a Reise is part of the familie content
// type set (app/familie/reise.go) and is served by the same API, so it is
// covered by the familie entitlement and guarded by the same registry prefix.
// A route of its own would have needed an entitlement of its own.
const { params } = useRoute()

const key = computed(() => (Array.isArray(params.reise) ? params.reise.join('/') : params.reise))

const { data: entry, error } = await useFetch<ReiseEntry>(() => `/reisen/${key.value}`, {
  baseURL: useRuntimeConfig().public.apiBaseURL,
})
</script>
