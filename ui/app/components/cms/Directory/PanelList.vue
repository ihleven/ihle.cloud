<template>
  <ul class="grid grid-cols-[50%,20%,20%,10%] rounded border-0 border-gray-300 bg-white p-2">
    <li
      v-for="(f, i) in files"
      :key="f.path"
      class="col-span-4 grid grid-cols-subgrid rounded even:bg-gray-200"
      :class="{ 'bg-sky-300 text-gray-100 even:bg-sky-300': active === i }"
      @click="activate(f.entry, i)"
      @dblclick="goto(i)"
    >
      <main class="flex w-full items-center overflow-hidden text-ellipsis">

        <UButton
          :icon="f.type === 'Dir'?f.opened ? 'i-lucide-chevron-down' : 'i-lucide-chevron-right':''"
          class="h-6 w-8 shrink-0"
          variant="link" color="neutral"

          :style="{ 'margin-start': `${f.level * 1.5}rem` }"
          @click.stop="toggle(i)"
        />
        <div
          class="flex w-full items-center truncate focus:outline-none"
        >
          <Icon
            :name="f.type === 'Dir'?f.opened ? 'heroicons:folder-open' : 'heroicons:folder':'i-lucide-file-text'"
            class="mr-1 h-5 w-5 shrink-0 stroke-4"
          />
          {{ f.name }}
        </div>
      </main>

      <aside class="">{{ f.level }}</aside>

      <aside>{{ f.children }} - {{ f.parent?.children }}</aside>

    </li>

  </ul>
</template>

<script lang="ts" setup>
const props = defineProps<{
  entry: Entry & { content: Entry[] }
}>()

const emit = defineEmits<{
  active: [Entry]
}>()

type File = {
  name: string
  path: string
  slugs: string[]
  type: 'Dir' | 'File'
  opened: boolean
  parent?: File
  children: number
  level: number
  entry: Entry
}

function entry2file(e: Entry, parent?: File): File {
  return {
    name: e.name,
    path: e.path,
    slugs: e.path.split('/'),
    type: e.type === 'Dir' ? 'Dir' : 'File',
    opened: false,
    parent: parent,
    children: parent?.children || 0,
    level: parent ? parent.level + 1 : -1,
    entry: e,
  }
}

const files = ref(props.entry.type === 'Dir' ? [...props.entry.content].sort(foldersBeforeFiles).map(e => entry2file(e, entry2file(props.entry))) : [] as File[])

function foldersBeforeFiles(a: Entry, b: Entry): number {
  if (a.type === 'Dir' && b.type !== 'Dir') {
    return -1
  }
  if (b.type === 'Dir' && a.type !== 'Dir') {
    return 1
  }
  return a.path > b.path ? 1 : -1
}

// either open or close Directory
async function toggle(index: number) {
  const e = files.value[index]
  if (e.type !== 'Dir') {
    return
  }
  if (e.opened) {
    close(index)
  }
  else { open(index) }
}

// open Directory
async function open(index: number) {
  const e = files.value[index]
  if (e.opened || e.type !== 'Dir') {
    return
  }

  const data = await $fetch<Entry & { content: Entry[] }>(`/api/v1/entries/${e.path}`, {
    baseURL: useRuntimeConfig().public.apiBaseURL,
    credentials: 'include',
  })
  const newfiles = [...data.content].sort(foldersBeforeFiles).map(f => entry2file(f, e))

  files.value.splice(index + 1, 0, ...newfiles)

  e.children = newfiles.length
  e.opened = true

  let tmp = e.parent
  while (tmp) {
    tmp.children += newfiles.length
    tmp = tmp.parent
  }
}

// close Directory
async function close(index: number) {
  const e = files.value[index]
  if (!e.opened || e.type !== 'Dir') {
    return
  }

  let tmp = e.parent
  while (tmp) {
    tmp.children -= e.children
    tmp = tmp.parent
  }
  e.opened = false
  files.value.splice(index + 1, e.children || 0)
  e.children = 0
}

// highlight the active file
const active = ref<number | null>(null)

function activate(e: Entry, i: number) {
  active.value = i
  emit('active', e)
}

function goto(index: number) {
  const e = files.value[index].entry
  navigateTo(`/entries/${e.path}`)
}

function up() {
  if (active.value === null) {
    active.value = files.value ? files.value.length - 1 : null
    return
  }
  active.value = active.value === 0 ? 0 : active.value - 1
}
function down() {
  if (active.value === null) {
    active.value = files.value ? 0 : null
    return
  }
  active.value = active.value === files.value.length - 1 ? files.value.length - 1 : active.value + 1
}

defineShortcuts({
  arrowup: () => up(),
  arrowdown: () => down(),
  arrowleft: () => close(active.value || 0),
  arrowright: () => open(active.value || 0),
  enter: () => goto(active.value || 0),
})
</script>
