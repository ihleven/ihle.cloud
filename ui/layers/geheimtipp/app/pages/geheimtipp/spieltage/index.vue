<template>
  <section class="min-h-screen bg-[#e8e8e8]">
    <h1 class="p-2 py-8 text-lg text-gray-500">Spieltage {{ edition }}</h1>

    <p v-if="status === 'pending'" class="px-2 text-sm text-gray-500">Wird geladen …</p>

    <UAlert
      v-else-if="error" class="mx-2" color="error"
      variant="soft"
      title="Die Spieltage sind gerade nicht erreichbar"
    />

    <article v-else class="divide-y divide-dashed divide-gray-300 border-y border-gray-300 bg-white">
      <NuxtLink
        v-for="cr in season?.rounds ?? []" :key="cr.num"
        :to="`/geheimtipp/spieltag-${edition}-${cr.num}`"
        class="flex items-center justify-between py-1 pl-2 hover:bg-sky-50"
      >
        <p>
          Spieltag {{ cr.num }}
          <small class="text-gray-500">{{ ghtDateRange(cr.from, cr.to) }}</small>
        </p>
        <UIcon name="i-heroicons-chevron-right" class="ml-auto h-8 w-8 text-gray-400" />
      </NuxtLink>
    </article>
  </section>
</template>

<script setup lang="ts">
import type { CompSeason } from '../../../types'

definePageMeta({ public: true, layout: 'geheimtipp' })

const { edition } = useGhtSession()
const { comp, load } = useGhtEdition()

// The edition names the competition; the competition's season carries the
// rounds. /compseasons without a season 500s, which is why this asks for both.
await load(edition.value)

const { data: season, status, error } = await useGht<CompSeason>(
  () => `/compseasons/${comp.value}/${edition.value}`,
)
</script>
