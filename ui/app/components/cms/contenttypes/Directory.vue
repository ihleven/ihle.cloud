<template>
  <article
    v-if="entry.type==='Dir'"
    class="bg-gray-100"
  >
    <header class="flex h-12 items-center justify-start border-b border-gray-300 bg-gray-50 px-2 py-2 text-gray-700">
      <span class="grow font-semibold text-gray-600">
        <!-- {{ active?.path.split('/')[active?.path.split('/').length-1] }} -->
        <!-- {{ active?.path.split('/') }}
          {{ active?.path.split('/').length }} -->
      </span>
      <USeparator
        orientation="vertical"
        class="h-8"
      />

      <!-- <USelect
        variant="none" :items="active?.path.split('/').reverse().map((p, i) => ({ label: p, value: i }))"
        class="w-fit"
        :ui="{ content: 'w-24', base: 'text-gray-400 font-mono' }"
      /> -->

      <UDropdownMenu
        color="primary"
        :disabled="folderItems.length === 0"
        arrow
        :items="folderItems"
        :ui="{
          content: 'w-fit',
          item: 'rounded hover:bg-sky-300',
        }"
      >
        <UButton
          icon="i-lucide-folder-tree"
          color="neutral"
          variant="subtle"
          class="mx-4"
        />
        <!-- <template #item="{ item }">
          <button class="text-gray-600 flex items-center gap-2 hover:text-gray-300 hover:bg-sky-300" @click="goto(item)">
            <UIcon name="i-lucide-folder" class="shrink-0 size-5 text-primary " />
            <span>{{ item.label }}</span>
          </button>
        </template> -->
      </UDropdownMenu>

      <USeparator
        orientation="vertical"
        class="h-8"
      />

      <UFieldGroup class="mx-4">
        <UButton
          color="neutral"
          variant="subtle"
          class="h-8 w-12 px-3"
          @click="setMode('cols')"
        >
          <UIcon
            name="i-lucide-columns-3"
            class="size-5"
            :class="{ 'text-primary-300': mode==='cols' }"
          />
        </UButton>
        <UButton
          color="neutral"
          variant="subtle"
          class="h-8 w-12 px-3"
          @click="setMode('list')"
        >
          <UIcon
            name="i-lucide-list"
            class="size-5 hover:text-primary-300"
            :class="{ 'text-primary-300': mode==='list' }"
          />
        </UButton>
      </UFieldGroup>
    </header>

    <Panel
      v-if="mode=='cols'"
      :entry="entry"
      @active="(e) => active = e"
    />
    <PanelList
      v-else-if="mode=='list'"
      :entry="entry"
      @active="(e) => active = e"
    />
  </article>
</template>

<script setup lang="ts">
import { useStorage } from '@vueuse/core'

defineProps<{ entry: Entry }>()
const cat = '' // localStorage.getItem('mode')

const mode = useStorage<'cols' | 'list'>('ihleven-dir-mode', cat || 'cols')

// const mode = ref<'cols' | 'list'>(cat || 'cols')

function setMode(m: 'cols' | 'list') {
  mode.value = m
  localStorage.setItem('mode', m)
}

const active = ref<Entry | null>(null)

const folderItems = computed(() => {
  console.log('active:', !active.value)
  if (!active.value) return []
  const basename = active.value?.path.replace(/.*\//, '') || ''
  const path = basename.indexOf('.') > -1 ? dirname(active.value.path) : active.value.path
  const slugs = path?.split('/').filter(slug => !!slug)
  return slugs.map((p, i) => ({
    label: p,
    icon: 'i-lucide-folder',
    onSelect: () => {
      navigateTo(`/entries/${slugs.slice(0, i + 1).join('/')}`)
    },
  }))
})
</script>
