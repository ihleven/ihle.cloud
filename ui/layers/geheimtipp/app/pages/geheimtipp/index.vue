<template>
  <div>
    <h1 class="text-xl font-semibold">Ausgaben</h1>

    <p v-if="pending" class="mt-4 text-sm text-gray-500">Wird geladen …</p>

    <UAlert
      v-else-if="error"
      class="mt-4" color="error" variant="soft"
      title="Die Tipprunde ist gerade nicht erreichbar"
      :description="String(error)"
    />

    <ul v-else class="mt-4 divide-y divide-gray-200 border-y border-gray-200">
      <li v-for="ausgabe in ausgaben" :key="ausgabe.id" class="flex items-baseline gap-4 py-3">
        <span class="w-16 shrink-0 font-mono text-sm text-gray-500">{{ ausgabe.code }}</span>
        <span class="grow">{{ ausgabe.name }}</span>
        <UBadge v-if="ausgabe.aktuell" variant="subtle" size="sm">aktuell</UBadge>
        <span class="w-24 text-right text-sm text-gray-500">
          {{ ausgabe.numTipper }} Tipper
        </span>
      </li>
    </ul>
  </div>
</template>

<script setup lang="ts">
// Reachable without signing in: the pool is public and always has been, and the
// family app's sign-in overlay would otherwise cover it. This is also why it is
// not an area — an entitlement would refuse a signed-in family member a page
// that strangers can read.
definePageMeta({ public: true, layout: 'geheimtipp' })

type Ausgabe = {
  id: number
  code: string
  name: string
  season: number
  aktuell: boolean
  numTipper: number
}

// The pool's own backend, reached through this app's proxy. The base is runtime
// config so that serving both from one origin later changes a value rather than
// this file.
const base = useRuntimeConfig().public.geheimtippBase as string

const { data: ausgaben, pending, error } = await useFetch<Ausgabe[]>('/ausgaben', {
  baseURL: base,
  // The pool keeps its own session in a cookie of its own; sending it costs
  // nothing here and is what the tipping pages will need.
  credentials: 'include',
})
</script>
