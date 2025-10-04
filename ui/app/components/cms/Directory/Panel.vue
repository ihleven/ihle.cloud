<template>
  <div class="h-full">
    <main
      v-if="entry.type ==='Dir'"
      class="flex h-full w-full max-w-full items-stretch divide-x divide-gray-300 overflow-auto bg-white"
    >
      <ul class="max-w-64 min-w-fit space-y-0.5 p-2">
        <li
          v-for="e in files"
          :key="e.path"
          @click.stop="load(e)"
        >
          <div
            :to="entrylink(e.path)"
            class="flex items-center rounded hover:bg-gray-300"
            :class="{ 'bg-gray-200': active?.path === e.path }"
          >
            <img
              v-if="e.type==='Dir'"
              :src="e.type==='Dir' ? '/folder.png' : ''"
              class="mx-1 my-0.5 h-6 w-6"
            >
            <i
              v-else-if="e.locale"
              class="relative mx-1 my-0.5 h-6 w-6 shrink-0"
            >
              <Icon
                :name="`flagpack:${e.locale.startsWith('en') ? 'gb-ukm' : e.locale.substring(0, 2)}`"
                class="h-full w-full"
              />
              <small class="absolute -top-1 -right-1 rounded-full bg-gray-200">{{ e.locale.substring(3) }}</small>
            </i>
            <UIcon
              v-else
              :name="e.type==='Dir' ? 'i-lucide-folder' : 'i-lucide-book'"
              class="size-5"
            />
            <div class="px-1">
              {{ e.name || basename(e.path) }}
            </div>
            <UIcon
              v-if="e.type==='Dir'"
              name="i-lucide-chevron-right"
              class="ml-auto size-4 text-gray-500"
            />
          </div>
        </li>
      </ul>
      <aside class="flex-initial border-0 border-black">
        <Panel
          v-if="active"
          :entry="active"
          class="h-full"
          @active="(e) => emit('active', e)"
        />
      </aside>
    </main>
    <section
      v-else-if="entry.type==='Person'"
      class="p-4"
    >
      <p class="text-gray-500">
        <NuxtLink
          :to="entrylink(entry.path)"
          class="group max-w-md"
        >

          {{ entry.content.key }}

        </NuxtLink>
      </p>
    </section>

    <section
      v-else-if="entry.type==='Article'"
      class="group flex flex-col justify-center p-4 hover:bg-gray-100"
    >
      <NuxtLink
        :to="entrylink(entry.path)"
        class="group max-w-md"
      >

        <img
          :src="`https://images.interhome.group/uploads/travelguide/articles/${entry.id}/${entry.content.media.teaser?.src || entry.content.commons.fallbackmedia.teaser?.src}`"
          class="bg-black text-gray-500"
        >

        <aside class="py-1">
          <h4 class="font-quattro">TRAVELGUIDE ARTICLE
            <small class="font-quattro text-gray-500">{{ entry.content.commons.novaID }}/{{ entry.content.translation_id }}</small>
          </h4>
          <h5 class="text-base font-semibold text-highlighted">{{ entry.content.commons.virtpath }}</h5>
          <div class="space-x-2">

            <UBadge
              :label="entry.content.commons.destination"
              color="primary"
              variant="outline"
            />

            <UBadge
              :label="entry.content.commons.topic"
              color="primary"
              variant="solid"
            />
            <UBadge
              v-for="category in entry.content.commons.categories"
              :key="category"
              :label="category"
              color="primary"
              variant="subtle"
            />
          </div>
          <!-- <img :src="`https://images.interhome.group/uploads/travelguide/articles/${entry.id}/${entry.content.commons.fallbackmedia?.hero?.src}`" class="ml-auto h-10">
          <img :src="`https://images.interhome.group/uploads/travelguide/articles/${entry.id}/${entry.content.commons.fallbackmedia?.teaser?.src}`" class="h-10"> -->
          <small>{{ entry.content.summary?.mobile || entry.content.summary?.leadin || entry.content.summary?.newsletter }}</small>
        </aside>

      </NuxtLink>
    </section>

    <section
      v-else
      class="p-4"
    >
      <NuxtLink
        :to="entrylink(entry.path)"
        class="text-blue-500 hover:underline"
      >
        <i class="i-lucide-folder" />
        <h4>NAME: {{ entry.name }}</h4>
        <h5>PATH: {{ entry.path }}</h5>
        <p>{{ entry.type }}</p>
      </NuxtLink>
    </section>
  </div>
</template>

<script setup lang="ts">
const props = defineProps<{
  entry: Entry
}>()

const emit = defineEmits<{
  active: [Entry]
}>()

function foldersBeforeFiles(a: Meta, b: Meta): number {
  if (a.type === 'Dir' && b.type !== 'Dir') {
    return -1
  }
  if (b.type === 'Dir' && a.type !== 'Dir') {
    return 1
  }
  return a.path > b.path ? 1 : -1
}
const files = props.entry.type === 'Dir' ? [...props.entry.content].sort(foldersBeforeFiles) : [] as Meta[]

const active = ref<Entry | null>(null)

async function load(e: Meta) {
  active.value = null

  try {
    const data = await $fetch<Entry>(`/api/v1/entries/${e.path}`, {
      baseURL: useRuntimeConfig().public.api.base,
      credentials: 'include',
    })
    active.value = data
    emit('active', active.value)
    console.log('loaded', active.value)
  }
  catch (error) {
    console.error('Error loading entry:', error)
  }
}
</script>
