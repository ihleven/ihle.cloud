<template>
  <div class="min-h-screen bg-white">
    <!-- <SearchBox /> -->
    <section class="flex items-center justify-center p-4">
      <UInput
        v-model="params.q"
      />
    </section>
    <section class="flex items-stretch justify-start divide-x divide-gray-300">
      <aside class="w-1/4 max-w-48">
        <FacetSelect
          v-if="result.facets?.version"
          v-model="params.version"
          :options="['published', 'draft']"
          :facet="result.facets.version"
        />
        <FacetSelectMulti
          v-if="result.facets?.type"
          v-model="params.type"
          :facet="result.facets.type"
        />
        <!-- <SearchFacet
          v-if="result.facets?.type"
          v-model:checked="params.type"
          name="ContentType"

          :facet="result.facets?.type"
          @update:checked="search"
        /> -->

        <!-- <UCommandPalette
          v-model="value"
          :autofocus="false"
          :multiple="true"
          placeholder="Content type"
          :groups="[
            {
              id: 'contenttype',
              label: 'Content types',
              items: facets.type,
            },
          ]"
          class="flex-1"
          @update:model-value="params.type = value.map(v => v.label);search()"
        >
          <template #item-trailing="{ item }">
            <UBadge
              :color="item.id === 'Page' ? 'primary' : 'neutral'"
              class="rounded-full px-1 py-0.5"
            >
              {{ item.count }}
            </UBadge>
          </template>
        </UCommandPalette> -->
      </aside>
      <section class="grow divide-y divide-gray-300 px-4">
        <header class="flex items-center justify-between">
          <h2 class="text-xl font-bold" />
          <small class="text-gray-500">{{ result.total_hits }} hits in {{ (result.took / 1000 / 1000).toFixed() }}ms,  max_score {{ result.max_score.toFixed(2) }}, cost: {{ result.cost }} </small>
        </header>
        <article
          v-for="hit in result.hits"
          :key="hit.id"
          class="py-2 ring-0 ring-sky-300"
        >
          <NuxtLink
            :to="entrylink(hit.id)"
            class="flex items-start text-xs"
          >
            <img
              v-if="hit.fields?.image"
              :src="`https://images.interhome.group/travelguide/${hit.fields.image}.jpg/tr:w-64,h-64`"
              :alt="hit.imgalt?.de"
              class="mr-2 h-8 w-8 rounded"
            >
            <UAvatar
              v-else
              icon="i-lucide-image"
              class="mr-2 h-8 w-8 rounded-full"
            />
            <header class="relative w-full">
              <h3 class="text-sm leading-none font-semibold text-gray-700">{{ hit.fields.name || 'Name' }}</h3>
              <h4 class="text-xs font-normal text-gray-400">{{ hit.fields.type }} | {{ hit.fields.path ||'Path' }}</h4>
              {{ hit.fields.id }}
              {{ hit.fields.modified }}
              <aside class="text-light absolute top-0 right-0 flex flex-row gap-1 text-xs">

                <small
                  v-if="hit.fields.version"
                  class="rounded-full bg-sky-300 px-2 py-0.5"
                >{{ hit.fields.version }}</small>
                <small
                  v-if="hit.fields.status"
                  class="rounded-full bg-green-300 px-2 py-0.5"
                >{{ hit.fields.status }}</small>

              </aside>
            </header>
          </NuxtLink>
          <small
            v-for="(fragment, key) in hit.fragments"
            :key="key"
          >
            <p
              v-for="f in fragment"
              :key="f"
              v-html="f"
            />
          </small>
        </article>

        <div class="flex h-16 items-center justify-center">
          <UPagination
            v-model:page="params.page"
            :items-per-page="params.size"
            :total="result.total_hits"
            :show-edges="true"
            class="text-center"
          />
        </div>
      </section>
      <aside class="w-1/4 max-w-48 space-y-4 px-4">
        <UFormField label="Sorting">
          <USelectMenu
            v-model="params.sorting"
            multiple
            :items="['-modified', 'id', '_score', 'version']"
            placeholder="Sorting"
            :search-input="false"
            class="w-full"
          />
        </UFormField>

        <UFormField
          label="Page size"
          help=""
        >
          <UInputNumber
            v-model="params.size"
            placeholder="Enter page size"
            :step="5"
            :min="5"
            :max="50"
          />
        </UFormField>

        {{ filters }}
        <SearchFacet
          v-for="(facet, name) in result?.facets"
          :key="facet.field"
          v-model:checked="filters[name]"
          :name="name"
          :facet="facet"
        />
      </aside>
    </section>
  </div>
</template>

<script setup lang="ts">
definePageMeta({
  middleware: [
    function (to, from) {
      const { session } = useAuth()

      if (!session.value.permissions.bar) {
        console.log('search not allowed', process.env, session.value.permissions)
        // return navigateTo('/login')
      }

      // counter.value = counter.value || Math.round(Math.random() * 1000)
      console.log('middleware', session.value, to.fullPath, from.fullPath)
    },

  ],
})
const { currentRoute, push } = useRouter()
const { query } = useRoute()

// facetConfig definiert die Facette vor, result.facet enthält ein objekt mit den Counts

const filters = ref({ ...currentRoute.value.query })

const { search, params, result, facets } = useSearch()

// onMounted(() => {
//   search()
// })
onBeforeMount(() => {
  params.value = { ...params.value, ...query }
})
// onBeforeUnmount(() => {
//   push({ query: { } })
// })

watchEffect(() => {
  if (currentRoute.path === '/api/v1/search') {
    push({ query: { ...params.value } })
  }
})

// const { data, status } = await useFetch(config.public.apiBaseURL + '/suggest', {
//   credentials: 'include',
//   // headers: { Cookie: useRequestHeader('Cookie') },
//   transform: (data: { id: number, name: string, email: string }[]) => {
//     return data?.map(user => ({ id: user.id, label: user.name, suffix: user.email, avatar: { src: `https://i.pravatar.cc/120?img=${user.id}` } })) || []
//   },
//   lazy: true,
//   query: {
//     fields: '*',
//     highlight: true,
//     suggest: searchTerm,
//   },
// })
</script>
