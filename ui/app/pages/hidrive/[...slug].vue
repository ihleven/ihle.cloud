<template>
  <main class="min-h-full w-full bg-gray-50">
    {{ meta?.category }}

    <header class="border-b border-b-gray-900/10 p-8 lg:border-t lg:border-t-gray-900/5">
      <h2 class="text-sm leading-6 font-medium text-gray-500">
        {{ meta?.type !== 'dir' ? meta?.mime_type : meta?.category }}
      </h2>
      <h1 class="text-xl font-bold tracking-tight text-gray-800">
        {{ meta ? driveName(meta) : '' }}
      </h1>

      <section class="flex space-x-8">
        <aside class="w-1/2">
          <dl class="w-1/2 items-center justify-between">
            <dt class="text-xs leading-6 font-medium text-gray-400">size:</dt>
            <dd class="text-sm font-medium text-gray-700">
              {{ driveSize(meta?.size) }}
              <template v-if="meta?.type === 'dir'">({{ meta?.nmembers }} Dateien)</template>
            </dd>
          </dl>
          <dl class="w-1/2 items-center justify-between">
            <dt class="text-xs leading-6 font-medium text-gray-400">mtime:</dt>
            <dd class="text-sm font-medium text-gray-700">{{ stamp(meta?.mtime) }}</dd>
          </dl>
          <dl class="w-1/2 items-center justify-between">
            <dt class="text-xs leading-6 font-medium text-gray-400">ctime:</dt>
            <dd class="text-sm font-medium text-gray-700">{{ stamp(meta?.ctime) }}</dd>
          </dl>
          <dl class="w-1/2 items-center justify-between">
            <dt class="text-xs leading-6 font-medium text-gray-400">
              {{ meta?.readable ? 'readable' : '' }} / {{ meta?.writable ? 'writable' : '' }}
            </dt>
            <dd class="text-sm font-medium text-gray-700" />
          </dl>
        </aside>

        <aside v-if="meta?.category === 'image'" class="w-1/2 text-xs">
          <dl class="flex items-center justify-between text-xs leading-6 font-medium">
            <dd class="text-gray-400">Image size:</dd>
            <dd class="text-gray-700">{{ meta?.image?.width }}x{{ meta?.image?.height }}</dd>
          </dl>
          <dl v-for="(v, k) in meta?.image?.exif" :key="k" class="flex items-center justify-between text-xs">
            <dd class="font-medium text-gray-400">{{ k }}:</dd>
            <dd class="font-normal text-gray-900">{{ v }}</dd>
          </dl>
        </aside>
      </section>
    </header>

    <DriveBreadcrumbs :path="path" class="sticky top-0 border-b border-gray-300 bg-gray-100" />

    <div v-if="meta?.category === 'office'" class="h-full">
      <embed :type="meta.mime_type" :src="media(path)" class="h-screen w-full">
    </div>

    <section v-else-if="meta?.category === 'image'" class="bg-black">
      <img :src="media(path)" class="mx-auto max-h-screen" :alt="meta.name">
    </section>

    <section v-else-if="meta?.category === 'video'" class="bg-black">
      <DriveVideo :path="path" :type="meta.mime_type" />
    </section>

    <!-- Audio has no branch in the browser this was ported from: an mp3 landed
         on a blank page. A player is the obvious thing to put there. -->
    <section v-else-if="meta?.category === 'audio'" class="bg-gray-100 p-8">
      <audio controls crossorigin="use-credentials" class="mx-auto w-full max-w-screen-md" :src="media(path)" />
    </section>

    <section v-else-if="meta?.category === 'code'" class="overflow-y-scroll bg-gray-50">
      <pre class="mx-auto max-h-screen p-4 text-xs">{{ code }}</pre>
    </section>

    <!-- Verzeichnis -->
    <section
      v-else-if="meta?.members"
      class="mx-auto flex max-w-screen-md flex-col items-stretch border-x border-white bg-white shadow-lg md:flex-row"
    >
      <ul role="list" class="w-full divide-y divide-gray-300 border-r border-gray-300 md:w-2/4">
        <li v-for="f in meta.members" :key="f.name">
          <NuxtLink :to="child(f)" class="flex">
            <div class="mr-2 flex h-14 w-14 shrink-0 items-center justify-center p-1">
              <img
                v-if="f.category === 'image' || f.category === 'video'"
                :src="thumb(join(path, driveName(f)), 100)"
                class="aspect-square rounded object-cover object-center"
                aria-hidden="true"
              >
              <UIcon v-else name="i-lucide-folder" class="h-8 w-8 stroke-1" />
            </div>
            <div>
              <h4 class="text-sm font-semibold">{{ driveName(f) }}</h4>
              <p class="text-xs font-normal text-gray-400">
                <template v-if="f.type === 'dir'">
                  {{ f.category }}, {{ f.nmembers }} files, {{ driveSize(f.size) }}
                </template>
                <template v-else-if="f.category === 'image'">
                  {{ f.mime_type }}, {{ f.image?.width }}x{{ f.image?.height }}px, {{ driveSize(f.size) }}
                </template>
                <template v-else-if="f.category === 'code'">
                  {{ f.category }}, {{ driveSize(f.size) }}
                </template>

                <small class="block">
                  {{ stamp(f.mtime ?? f.ctime) }} {{ f.readable ? 'R' : '' }}{{ f.writable ? 'W' : '' }}
                </small>
              </p>
            </div>
          </NuxtLink>
        </li>
      </ul>

      <div class="w-full md:w-1/2">
        <div class="h-full border-b border-gray-300 bg-white md:w-[50vw]">
          <DriveGallery :images="images" gallery="drive-gallery" :dir="path" />
        </div>
      </div>
    </section>
  </main>
</template>

<script setup lang="ts">
const route = useRoute()
const { meta: fetchMeta, media, thumb } = useDrive()

const path = computed(() => {
  const slug = route.params.slug
  return Array.isArray(slug) ? slug.join('/') : (slug ?? '')
})

const { data: meta } = await fetchMeta(path)

const images = computed(() => (meta.value?.members ?? []).filter(m => m.category === 'image'))

function join(dir: string, name: string): string {
  return [dir, name].filter(Boolean).join('/')
}

function child(f: DriveMeta): string {
  // Decoded: the router encodes a path parameter on its way into the URL, and
  // handing it an already-encoded name would encode it twice.
  return `/hidrive/${join(path.value, driveName(f))}`
}

function stamp(seconds?: number): string {
  return seconds ? new Date(seconds * 1000).toLocaleString() : ''
}

// A source file is shown rather than downloaded, so its text is fetched too.
const code = ref('')
watchEffect(async () => {
  code.value = meta.value?.category === 'code' ? await driveText(media(path.value)).catch(() => '') : ''
})
</script>
