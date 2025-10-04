<template>
  <main class="bg-gray-50">
    {{ meta?.category }}
    <header class="border-b border-b-gray-900/10 p-8 lg:border-t lg:border-t-gray-900/5">
      <h2 class="text-sm font-medium leading-6 text-gray-500">
        {{ meta?.type !== 'dir' ? meta?.mime_type : meta?.category }}
      </h2>
      <h1 class="text-xl font-bold tracking-tight text-gray-800">
        {{ decodeURI(meta? meta.name:'') }}
      </h1>

      <section class="flex space-x-8">
        <aside class="w-1/2">
          <dl class="w-1/2 items-center justify-between">
            <dt class="text-xs font-medium leading-6 text-gray-400">
              size:
            </dt>
            <dd class="text-sm font-medium text-gray-700">
              {{ bytes(meta?.size) }}
              <template v-if="meta?.type === 'dir'">
                ({{ meta?.nmembers }} Dateien)
              </template>
            </dd>
          </dl>
          <dl class="w-1/2 items-center justify-between">
            <dt class="text-xs font-medium leading-6 text-gray-400">
              mtime:
            </dt>
            <dd class="text-sm font-medium text-gray-700">
              {{ meta? new Date(meta.mtime * 1000).toLocaleString() :'' }}
            </dd>
          </dl>
          <dl class="w-1/2 items-center justify-between">
            <dt class="text-xs font-medium leading-6 text-gray-400">
              ctime:
            </dt>
            <dd class="text-sm font-medium text-gray-700">
              {{ meta ? new Date(meta.ctime * 1000).toLocaleString() : '' }}
            </dd>
          </dl>
          <!-- <dl class="w-1/2 items-center justify-between">
            <dt class="text-xs font-medium leading-6 text-gray-400">path:</dt>
            <dd class="text-sm font-medium text-gray-700">{{ meta?.path }}</dd>
          </dl> -->
          <!-- <dl class="w-1/2 items-center justify-between">
            <dt class="text-xs font-medium leading-6 text-gray-400">id:</dt>
            <dd class="text-sm font-medium text-gray-700">{{ meta?.id }}</dd>
          </dl> -->
          <dl class="w-1/2 items-center justify-between">
            <dt class="text-xs font-medium leading-6 text-gray-400">
              {{ meta?.readable ? 'readable' : '' }} / {{ meta?.writable ? 'writable' : '' }}
            </dt>
            <dd class="text-sm font-medium text-gray-700" />
          </dl>
        </aside>
        <aside
          v-if="meta?.category === 'image'"
          class="w-1/2 text-xs"
        >
          <dl class="flex items-center justify-between text-xs font-medium leading-6">
            <dd class="text-gray-400">
              Image size:
            </dd>
            <dd class="text-gray-700">
              {{ meta?.image?.width }}x{{ meta?.image?.height }}
            </dd>
          </dl>
          <dl
            v-for="(v, k) in meta?.image?.exif"
            :key="k"
            class="flex items-center justify-between text-xs"
          >
            <dd class="font-medium text-gray-400">
              {{ k }}:
            </dd>
            <dd class="font-normal text-gray-900">
              {{ v }}
            </dd>
          </dl>
        </aside>
      </section>
    </header>
    <Breadcrumbs
      :path="route.path"
      class="sticky top-0 border-b border-gray-300 bg-gray-100"
    />
    <section class="">
      <!-- {{ meta.path }} -->

      <!-- {{ meta.parent_id }}<br /> -->
    </section>

    <div v-if="meta?.category==='office'" class="h-full">
      <!-- <ClientOnly>
        <HidriveMarkdown :path="path" />
      </ClientOnly> -->
      <embed :type="meta.mime_type" :src="`http://localhost:8000/hi/media/${meta?.path}`" class="h-screen w-full">
      <!-- <pre class="contain mx-auto max-h-screen text-xs p-4">{{ code }}</pre> -->
    </div>
    <section
      v-else-if="meta?.category === 'image'"
      class="bg-black"
    >
      <img
        :src="`http://localhost:8000/hi/media/${meta?.path}`"
        class="contain mx-auto max-h-screen"
      >
    </section>

    <section
      v-else-if="meta?.category === 'video'"
      class="bg-black"
    >
      <HidriveVideo
        :meta="meta"
        class="contain mx-auto max-h-screen"
      />
    </section>

    <section
      v-else-if="meta?.category === 'code'"
      class="bg-gray-50 overflow-y-scroll"
    >
      <!-- <embed :type="meta.mime_type" :src="`http://localhost:8000/hi/media/${meta?.path}`" /> -->
      <pre class="contain mx-auto max-h-screen text-xs p-4">{{ code }}</pre>
    </section>

    <!-- Verzeichnis -->
    <section
      v-else-if="meta?.members"
      class="mx-auto flex max-w-screen-md flex-col items-stretch border-x border-white bg-white shadow-lg md:flex-row"
    >
      <ul
        role="list"
        class="w-full divide-y divide-gray-300 border-r border-gray-300 md:w-2/4"
      >
        <li
          v-for="f in meta.members"
          :key="f.id"
        >
          <nuxt-link
            :to="`/hidrive/${[meta?.path, f.name].join('/')}${f.type === 'dir' ? '/' : ''}`"
            class="flex"
          >
            <div class="mr-2 flex h-14 w-14 flex-shrink-0 items-center justify-center p-1">
              <img
                v-if="f.category == 'image' || f.category == 'video'"
                :src="`http://localhost:8000/hi/media/thumbs?path=${[meta?.path, f.name].join('/')}&width=100`"
                class="aspect-square rounded object-cover object-center"
                aria-hidden="true"
              >
              <Icon
                v-else
                name="folder"
                class="h-8 w-8 stroke-1"
                filled
              />
            </div>
            <div>
              <h4 class="text-sm font-semibold">{{ decodeURI(f.name) }}</h4>
              <p class="text-xs font-normal text-gray-400">
                <template v-if="f.type == 'dir'">
                  {{ f.category }}, {{ f.nmembers }} files, {{ bytes(f.size) }}
                </template>
                <template v-else-if="f.category == 'image'">
                  {{ f.mime_type }}, {{ f.image.width }}x{{ f.image.height }}px, {{ bytes(f.size) }}
                </template>
                <template v-else-if="f.category == 'code'"> {{ f.category }}, {{ bytes(f.size) }} </template>

                <small class="block">
                  {{ new Date(f.mtime * 1000).toLocaleString() }} {{ f.readable ? 'R' : '' }}{{ f.writable ? 'W' : '' }}
                </small>
              </p>
            </div>
          </nuxt-link>
        </li>
      </ul>

      <div class="w-full md:w-1/2">
        <div class="h-full border-b border-gray-300 bg-white md:w-[50vw]">
          <HidriveGallery
            :images="images"
            gallery="asdf"
            :dir="'/' + meta?.path"
          />
        </div>
      </div>
    </section>
  </main>
</template>

<script setup lang="ts">
const config = useRuntimeConfig()
const { bytes } = useHelpers()

const route = useRoute()
const path = typeof route.params.slug === 'string' ? route.params.slug : route.params.slug.join('/')
console.log(`${config.public.hidrive.api}/meta/${path}`)

const { data } = await useFetch<Meta>(`${config.public.hidrive.api}/meta/${path}`, {
  server: false,
  credentials: 'include',
  headers: { Accept: 'application/json' },
})
const meta = computed(() => data.value)
// meta.value.name = decodeURI(meta.value.name)
console.log('meta:', import.meta.client, meta.value, meta.value?.category)
const isMarkdown = false // computed(() => meta.value.name === 'README.md')

const images = computed(() => meta.value?.members.filter(m => m.category === 'image'))
console.log('category:', meta.value?.name)

const code = ref(null)
if (meta.value?.category === 'code') {
  code.value = await $fetch(`http://localhost:8000/hi/media/${path}`, {
    server: false,
    credentials: 'include',
    headers: { Accept: 'application/json' },
  })
}
</script>
