<template>
  <aside class="flex h-full flex-col">
    <header class="sticky top-0 flex h-12 flex-none justify-between border-y border-sky-100 bg-neutral-300">
      <img
        src="/interhome-bird.svg"
        class="h-full p-1"
      >
      <UButton
        icon="feather-x"
        class="m-2 rounded-full"
        @click="emit('close')"
      />
    </header>
    <section class="overflow-auto">
      <UAccordion
        :items="items"
        type="multiple"
        :unmount-on-hide="false"
        :ui="{ root: 'root', header: 'header px-4 ' }"
      >
        <template #folders="{ item }">
          <div
            v-if="folders.folders[item.folder]"
            class="border-y border-(--ui-primary) p-2 text-sm"
          >
            <Folders
              ref="sidebar"
              :folder="folders.folders[item.folder]"
              :root="true"
              path="/"
              @entry="onEntryClick"
            />
          </div>
        </template>
      </UAccordion>
    </section>

    <UNavigationMenu
      orientation="vertical"
      :items="navitems"
      class="data-[orientation=vertical]:w-full"
    />
  </aside>
</template>

<script setup lang="ts">
const props = defineProps<{
  open: boolean
}>()

const modelopen = computed({
  get: () => props.open,
  set: (state: boolean) => emit('update:open', state),
})

const emit = defineEmits<{ 'close': [boolean], 'update:open': [boolean] }>()

const config = useRuntimeConfig()

const { data: folders } = await useFetch(config.public.apiBaseURL + '/folders/', {
  credentials: 'include',
  headers: { Cookie: useRequestHeader('Cookie') },
})

console.log(folders)

function onEntryClick(path: string) {
  console.log('oEntryClick', path)
  emit('close')
  navigateTo(entrylink(path))
}
const items = [

  {
    label: 'Pages',
    icon: 'i-lucide-swatch-book',
    slot: 'folders',
    folder: 'pages',
    content: 'Nebenseiten',
  },
  {
    label: 'Travelguide',
    icon: 'i-lucide-box',
    slot: 'folders',
    folder: 'travelguide',
  },
]

function mapper(foldermapitem: FolderMapItem) {
  console.log('mapper', foldermapitem)
  const subfolders = Object.values(foldermapitem.folders || {}).map(folder => mapper(folder))
  const entries = foldermapitem.entries ? foldermapitem.entries.map(e => ({ label: e, icon: 'i-lucide-file', to: entrylink(foldermapitem.path + '/' + e) })) : []

  const v = {
    label: foldermapitem.name,
    icon: 'i-lucide-folder',
    // slot: 'folders',
    // to: foldermapitem.path,
    children: [...subfolders, ...entries],
  }
  return v
}

const pages = mapper(folders.value.folders.pages)

const navitems = ref<NavigationMenuItem[][]>([
  [
    {
      label: 'Links',
      type: 'label',
    },

    pages,
    mapper(folders.value.folders.travelguide),
    // {
    //   label: 'pages',
    //   icon: 'i-lucide-folder',
    //   // to: 'pages',
    //   children: [
    //     {
    //       label: 'topics',
    //       icon: 'i-lucide-folder',
    //       // to: 'pages/topics',
    //       children: [{ label: 'alps-resorts', icon: 'i-lucide-folder',
    //       // to: 'pages/topics/alps-resorts',
    //         children: [] }] }] },
  ],
])
</script>
