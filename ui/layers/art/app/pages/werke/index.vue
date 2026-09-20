<template>
  <div class="">
    <USlideover
      title="Slideover with title"
      side="left"
    >
      <UButton
        label="Open slideover"
        color="neutral"
        variant="link"
      />
      <template #footer>
        <div>Cost: {{ data.cost }}</div>
        <div>MaxScore: {{ data.max_score }}</div>

        <div>took: {{ data.took }}</div>

        <div>Total hits: {{ data.total_hits }}</div>
      </template>
      <template #body>
        <div class="w-1/2">
          <h4>Gattungen</h4>
          <URadioGroup
            v-model="filters.form"
            :items="forms"
            :ui="{ wrapper: 'w-full' }"
          >
            <template #label="{ item }">
              <p class="flex w-full items-center justify-between">
                {{ item.label }}
                <small v-if="item.count">({{ item.count }})</small>
              </p>
            </template>
          </URadioGroup>
        </div>
        <aside class="flex w-full justify-between space-x-8">
          <section class="w-1/2">
            <h4 class="mt-4">
              Technik
            </h4>
            <URadioGroup
              v-model="filters.medium"
              :items="media"
              :ui="{ wrapper: 'w-full' }"
            >
              <template #label="{ item }">
                <p class="flex w-full items-center justify-between">
                  {{ item.label }}
                  <small v-if="item.count">({{ item.count }})</small>
                </p>
              </template>
            </URadioGroup>
          </section>
          <section class="w-1/2">
            <h4 class="mt-4">
              Träger
            </h4>
            <URadioGroup
              v-model="filters.support"
              :items="supports"
              :ui="{ wrapper: 'w-full' }"
            >
              <template #label="{ item }">
                <p class="flex w-full items-center justify-between">
                  {{ item.label }}
                  <small v-if="item.count">({{ item.count }})</small>
                </p>
              </template>
            </URadioGroup>
          </section>
        </aside>
      </template>
    </USlideover>

    <!-- Why there is nothing to show. The endpoint refuses when the deployment
         indexes below fulltext, and without this the page would simply look
         empty — which is the failure it exists to prevent. -->
    <UAlert
      v-if="error"
      color="warning"
      variant="subtle"
      class="m-2"
      title="Die Werksuche ist nicht verfügbar"
      :description="reason"
    />

    <UContainer
      v-if="data"
      class="m-2 space-y-2 rounded border border-gray-300"
    >
      <div>{{ data.status }}</div>

      <section class="divide-y divide-gray-300">
        <article
          v-for="hit in data.hits"
          :key="hit.id"
        >
          {{ hit.fields }}
        </article>
      </section>
    </UContainer>
  </div>
</template>

<script setup lang="ts">
import type { ArtSearchResult } from '../../types'

definePageMeta({
  layout: 'default',
})

// The three filters the page offers, sent as query parameters. Empty ones are
// dropped by the fetch, so `undefined` is "no filter" rather than a value.
const filters = ref<{ form?: string, medium?: string, support?: string }>({
  form: undefined,
  medium: undefined,
  support: undefined,
})

const query = computed(() => Object.assign({ facets: '*' }, filters.value))

const { data, error } = useFetch<ArtSearchResult>('/art/search', {
  baseURL: useRuntimeConfig().public.apiBaseURL,
  server: false,
  query,
})

// The server says something useful — which SEARCH_LEVEL it is running at, when
// that is the problem — and it says it in `message`. Falling back to the fetch
// library's own string would leave the reader with a method and a status code.
const reason = computed(() => {
  const body = error.value?.data as { message?: string } | undefined

  return body?.message ?? error.value?.message ?? 'Unbekannter Fehler.'
})

// How many works carry this value, from the facet of the same name.
//
// The counts come back as bleve returns them — a list of {term, count} under
// `facets.<name>.terms` — rather than as a map keyed by value, so they are
// looked up rather than indexed. The "all" option has no term of its own and
// takes the facet's total.
function countOf(facet: string, value?: string) {
  const f = data.value?.facets?.[facet]
  if (!f) return undefined
  if (value === undefined) return f.total

  return f.terms?.find(t => t.term === value)?.count
}

// The values are the archive's own, exactly as the old database spelled them:
// a single capital for the Gattung, a capitalised word for Technik and Träger.
// They are compared as keywords, so "m" finds nothing where "M" finds 275.
const formoptions = [
  { value: undefined, label: 'alle' },
  { value: 'M', label: 'Malerei' },
  { value: 'Z', label: 'Zeichnung' },
  { value: 'P', label: 'Plastik' },
]

const mediumoptions = [
  { value: undefined, label: 'alle' },
  { value: 'Öl', label: 'Öl' },
  { value: 'Aquarell', label: 'Aquarell' },
  { value: 'Acryl', label: 'Acryl' },
]

const supportoptions = [
  { value: undefined, label: 'alle' },
  { value: 'Leinwand', label: 'Leinwand' },
  { value: 'Papier', label: 'Papier' },
  { value: 'Kupfer', label: 'Kupfer' },
]

const forms = computed(() => formoptions.map(o => ({ ...o, count: countOf('forms', o.value) })))
const media = computed(() => mediumoptions.map(o => ({ ...o, count: countOf('media', o.value) })))
const supports = computed(() => supportoptions.map(o => ({ ...o, count: countOf('support', o.value) })))
</script>
