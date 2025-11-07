<template>
  <article

    class="flex grow flex-col items-stretch bg-green-100"
  >
    <header

      class="sticky top-0 z-1 flex h-12 max-w-screen items-center border-b border-gray-300 bg-white px-2"
    >
      <div class="mr-auto shrink overflow-x-hidden">
        <UBreadcrumb
          :items="breadcrumbs"
          class="shrink-0"
        />
      </div>

      <UButton
        :icon="show?.meta ? 'i-lucide-x' : 'i-lucide-square-m'"
        size="md"
        :ui="{ leadingIcon: 'size-4' }"
        variant="ghost"
        :highlight="true"
        color="primary"
        class="rounded-xs hover:ring"
        @click="show.meta = !show.meta"
      />

      <UButton
        icon="i-feather-rewind"
        size="md"
        :ui="{ leadingIcon: 'size-4' }"
        variant="ghost"
        :highlight="true"
        color="primary"
        class="rounded-xs hover:ring"
        label="Save"
        @click="reset"
      >
        reset
      </UButton>

      <UButton
        icon="i-feather-save"
        size="md"
        :ui="{ leadingIcon: 'size-4' }"
        variant="ghost"
        :highlight="true"
        color="primary"
        class="rounded-xs hover:ring"
        label="Save"
        @click="save"
      />

      <UButton
        :icon="show?.slideright ? 'i-feather-x' : 'i-feather-menu'"
        size="md"
        :ui="{ leadingIcon: 'size-4' }"
        variant="ghost"
        :highlight="true"
        color="primary"
        class="rounded-xs hover:ring"
        @click="show.slideright = !show.slideright"
      />
    </header>

    <EntryMeta
      v-if="entry"
      :open="show.meta"
      :entry="entry"
      :ordering="true"
      class="sticky top-12 z-1"
      @patch:entry="patchEntry"
    />

    <!-- <section v-if="error">
      error: {{ error.statusCode }}: {{ error.statusMessage }}
    </section> -->

    <component
      :is="component"
      v-model:entry="entry"
      @patch:entry="patchEntry"
    />

    <EntrySource
      :open="show.source"
      :entry="entry"
    />
  </article>
</template>

<script setup lang="ts">
import { set } from 'lodash'

const show = ref({
  meta: false,
  source: false,
  slideright: false,
})

defineShortcuts({
  m: () => show.value.meta = !show.value.meta,
  s: () => show.value.source = !show.value.source,
  d: () => show.value.slideright = !show.value.slideright,
})

const { public: { api } } = useRuntimeConfig()

const { path, params } = useRoute()

const { data, refresh, error } = await useFetch<Entry>(path == '/entries' ? '/api/v1/entries/' : '/api/v1' + path, {
  baseURL: api.base as string,
  credentials: 'include',
  headers: import.meta.server ? new Headers({ cookie: String(useRequestHeader('cookie') ?? '') }) : undefined,
})
if (error.value) {
  console.error('FETCH ERROR:', error.value.statusCode)
  // throw error.value
}

const entry = ref<Entry>(data.value as Entry)

const componentmap: Record<string, object | string> = {
  Person: resolveComponent('Person'),
  Work: resolveComponent('Artwork'),
  Super8: resolveComponent('Super8'),
  Dir: resolveComponent('Directory'),
  Default: resolveComponent('Entry'),
}

const component = computed(() => entry.value ? componentmap[(entry.value as Entry).type] : componentmap.Default)
console.log('component', component.value)
function patchEntry(key: string, value: string | number | object | boolean | null | string[]) {
  if (entry.value) {
    set(entry.value, key, value)
  }
}

function reset() {
  refresh()
  if (data.value) {
    entry.value = data.value
  }
}
const toast = useToast()
const state = ref('idle')

async function save() {
  try {
    state.value = 'saving'
    const data = await $fetch<Entry>(path, {
      baseURL: api.base as string + `/api/v1/`,
      credentials: 'include',
      method: 'PUT',
      body: entry.value,
    })
    console.log(' *** entry saved =>', data)
    const a = Object.entries(data)[0]
    if (a) {
      const [path, e] = a
      entry.value = e
      state.value = 'saved'
      toast.add({ title: 'saved' })
    }
  }
  catch (e: unknown) {
    console.error('SAVE ERROR:', e)

    const error = JSON.parse(e.data)

    toast.add({
      title: 'Uh oh! Something went wrong.',
      description: error.cause,
      icon: 'i-lucide-error',
      color: 'error',
    })
    state.value = 'error'
  }
}

const breadcrumbs = computed(() => {
  const breadcrumbs: Array<object> = [{ icon: 'i-lucide-folder-tree', to: '/entries' }]
  if (params.slug) {
    const b = params.slug.map((slug, i) => ({ label: slug, to: entrylink(params.slug.slice(0, i + 1).join('/')) }))
    breadcrumbs.push(...b)
  }

  return breadcrumbs
},
)
</script>
