<template>
  <section class="space-y-2 p-4">
    <img :src="poster" class="aspect-video w-full bg-black object-cover">
    <p class="text-sm text-muted">
      {{ [entry.content.filmed_from, entry.content.location].filter(Boolean).join(' · ') }}
    </p>
    <p v-if="entry.content.scenes?.length" class="text-sm">
      {{ entry.content.scenes.length }} Szenen, {{ entry.content.annotations?.length ?? 0 }} Annotationen
    </p>
  </section>
</template>

<script setup lang="ts">
const props = defineProps<{ entry: FilmEntry }>()

const poster = computed(
  () => `${useRuntimeConfig().public.api.base}/api/v1/films/${encodeURIComponent(props.entry.id)}/poster.jpg`,
)
</script>
