<template>
  <main class="bg-gray-50">
    <header class="border-b border-b-gray-900/10 p-8 lg:border-t lg:border-t-gray-900/5">
      <h2 class="text-sm font-medium leading-6 text-gray-500">{{ meta.mime_type }}</h2>
      <h1 class="text-xl font-bold tracking-tight text-gray-800">{{ decodeURI(meta.name) }}</h1>

      <dl class="w-1/2 items-center justify-between">
        <dt class="text-xs font-medium leading-6 text-gray-400">size:</dt>
        <dd class="text-sm font-medium text-gray-700">{{ bytes(meta.size) }}</dd>
      </dl>
      <dl class="w-1/2 items-center justify-between">
        <dt class="text-xs font-medium leading-6 text-gray-400">content:</dt>
        <dd class="text-sm font-medium text-gray-700">{{ new Date(meta.mtime * 1000).toLocaleString() }}</dd>
      </dl>
      <dl class="w-1/2 items-center justify-between">
        <dt class="text-xs font-medium leading-6 text-gray-400">metadata:</dt>
        <dd class="text-sm font-medium text-gray-700">{{ new Date(meta.ctime * 1000).toLocaleString() }}</dd>
      </dl>
    </header>
    <Breadcrumbs :path="route.path" class="sticky top-0 border-b border-gray-300 bg-gray-100" />
    <section class="">
      <!-- {{ meta.path }} -->

      <!-- {{ meta.parent_id }}<br /> -->
    </section>

    <div v-if="isMarkdown">
      <HidriveMarkdown :path="path" />
    </div>
    <section v-else-if="meta.category === 'image'" class="bg-black">
      <img :src="`/api/raw${meta.path}`" class="contain mx-auto max-h-screen" />
    </section>

    <section
      v-else-if="meta.members"
      class="mx-auto flex max-w-screen-md flex-col items-stretch border-x border-white bg-white shadow-lg md:flex-row"
    >
      <ul role="list" class="w-full divide-y divide-gray-300 border-r border-gray-300 md:w-2/4">
        <li v-for="f in meta.members" :key="f.id">
          <nuxt-link :to="`/hidrive${f.path}`" class="flex">
            <div class="mr-2 flex h-14 w-14 flex-shrink-0 items-center justify-center p-1">
              <img
                v-if="f.category == 'image'"
                :src="`/api/thumbs?path=${f.path}&width=100`"
                class="aspect-square rounded object-cover object-center"
                aria-hidden="true"
              />
              <Micon v-else name="folder" class="h-8 w-8 stroke-1" filled />
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
          <HidriveGallery :images="meta.members" gallery="asdf"></HidriveGallery>
        </div>
      </div>
    </section>
  </main>
</template>

<script setup>
  const route = useRoute()
  const { bytes } = useHelpers()
  console.log(bytes)
  const path = typeof route.params.slug === 'string' ? route.params.slug : route.params.slug.join('/')

  const { data: meta } = await useFetch(`/api/meta/${path}`, {
    headers: { Accept: 'application/json' },
  })
  meta.name = decodeURI(meta.name)

  const isMarkdown = computed(() => meta.value.name === 'README.md')
</script>
