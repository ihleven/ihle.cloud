<template>
  <UCollapsible :open="open">
    <template #content>
      <div class="flex gap-4 border-b border-accented bg-default p-4 pb-4">

        <aside>
          <h3 class="text-base font-semibold text-highlighted">{{ entry.type }}</h3>
          <h4 class="text-xs font-semibold text-dimmed">{{ entry.mime }}</h4>
          <p class="text-xs font-extralight text-muted">{{ entry.path }}</p>
          <div class="mt-2 flex gap-2">
            <ul class="text-sm">
              <li><small class="inline-block w-[8ch]">ID:</small>{{ entry.id }}</li>
              <li><small class="inline-block w-[8ch]">space:</small>{{ entry.space }}</li>
              <li><small class="inline-block w-[8ch]">locale:</small>{{ entry.locale }}</li>
              <li><small class="inline-block w-[8ch]">version:</small>{{ entry.version }}</li>
              <li><small class="inline-block w-[8ch]">collection:</small>{{ entry.collection }}</li>
            </ul>
            <ul>
              <li class="text-xs font-extralight text-muted"><small>created:</small>{{ formatTS(entry.created) }}</li>
              <li class="text-xs font-extralight text-muted"><small>modified:</small>{{ formatTS(entry.modified) }}</li>
              <li class="text-xs font-extralight text-muted"><small>published:</small>{{ formatTS(entry.published) }}</li>
              <li class="text-xs font-extralight text-muted"><small>owner:</small>{{ entry.owner }}</li>
              <li class="text-xs font-extralight text-muted"><small>group:</small>{{ entry.group }}</li>
              <li class="text-xs font-extralight text-muted"><small>permissions:</small>{{ entry.permissions }}</li>

            </ul>

          </div>
        </aside>

        <ul class="grow divide-y divide-accented border-y-0 border-accented">
          <li>
            <UInput
              :model-value="entry.name"
              variant="none" :ui="{ base: 'pl-[8ch]', root: 'w-full' }"
              @update:model-value="update(`name`, $event)"
            >
              <template #leading><p class="text-sm text-muted">name</p></template>
            </UInput>
          </li>
          <li>
            <UInputTags
              :model-value="entry.tags"
              variant="none" :ui="{ base: 'pl-[8ch]', root: 'w-full' }"
              @update:model-value="update(`tags`, $event)"
            >
              <template #leading><p class="text-sm text-muted">tags</p></template>
            </UInputTags>
          </li>
          <li>
            <UTextarea
              autoresize placeholder="Take notes here..."
              :model-value="entry.notes" :rows="2"
              variant="none" :ui="{ base: 'pl-[8ch]', root: 'w-full' }"
              @update:model-value="update(`notes`, $event)"
            >
              <template #leading><p class="text-sm text-muted">notes</p></template>
            </UTextarea>
          </li>
          <li>
            <UInput
              :model-value="entry.slug"
              variant="none" :ui="{ base: 'pl-[8ch]', root: 'w-full' }"
              @update:model-value="update(`slug`, $event)"
            >
              <template #leading><p class="text-sm text-muted">slug</p></template>
            </UInput>
          </li>
          <li>
            <UInput
              :model-value="entry.full_slug"
              variant="none" :ui="{ base: 'pl-[8ch]', root: 'w-full' }"
              @update:model-value="update(`full_slug`, $event)"
            >
              <template #leading><p class="text-sm text-muted">full_slug</p></template>
            </UInput>
          </li>
          <li>
            <UInput
              :model-value="entry.status"
              variant="none" :ui="{ base: 'pl-[8ch]', root: 'w-full' }"
              @update:model-value="update(`status`, $event)"
            >
              <template #leading><p class="text-sm text-muted">status</p></template>
            </UInput>
          </li>
        </ul>

      </div>
    </template>
  </UCollapsible>
</template>

<script lang="ts" setup>
defineProps<{
  entry: Entry
  ordering: boolean
  open: boolean
}>()

const emit = defineEmits<{
  'patch:entry': [key: string, value: string | string[]]
}>()

function update(name: string, value: string | string[]) {
  emit('patch:entry', name, value)
}
</script>
