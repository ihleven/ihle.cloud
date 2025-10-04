<template>
  <article
    v-if="variant==='listitem'"
    class="border-0 border-accented p-1"
  >
    <UButton
      class="float-right -m-1"
      size="xs"
      variant="link"
      @click="openEditModal"
    >
      edit
    </UButton>
    <h4 class="text-sm font-bold">
      {{ entry.name }} <UBadge
        v-for="tag in entry.tags"
        :key="tag"
        :label="tag"
      />
    </h4>
    <p class="text-xs font-light text-muted">
      {{ entry.notes }}
    </p>
    {{ entry.content.geburtstag }} - {{ entry.content.todestag }}
    <div>
      <UBadge
        v-for="(tr, locale) in entry.content.locales"
        :key="locale"
        variant="outline"
        size="sm"
      >
        {{ locale }}
      </UBadge>
      <UButton
        class=""
        size="xs"
        variant="link"
        @click="open=!open"
      >
        show details...
      </UButton>
    </div>

    <UCollapsible
      v-model:open="open"
      class="p-0"
    >
      <template #content>
        {{ entry.content.locales }}
      </template>
    </UCollapsible>

    <UModal
      v-model:open="edit"
      :ui="{ content: 'max-w-3xl', body: 'sm:p-0' }"
    >
      <template #content>
        <Person
          :entry="entry"
          @update:entry="emit('update:entry', $event)"
        />
      </template>
    </UModal>
  </article>

  <article
    v-else
    class="p-4"
  >
    <div class="grid grid-cols-2 gap-4">
      <UFormField label="Vater">
        <UInput
          placeholder="Vater"
          :model-value="entry.content.vater"
          @update:model-value="update('content.vater', $event)"
        />
      </UFormField>
      <UFormField label="Mutter">
        <UInput
          placeholder="Mutter"
          :model-value="entry.content.mutter"
          @update:model-value="update('content.mutter', $event)"
        />
      </UFormField>
    </div>
    <UPopover>
      <UButton
        color="neutral"
        variant="subtle"
        icon="i-lucide-calendar"
      >
        <template v-if="modelValue.start">
          <template v-if="modelValue.end">
            {{ df.format(modelValue.start.toDate(getLocalTimeZone())) }} - {{ df.format(modelValue.end.toDate(getLocalTimeZone())) }}
          </template>

          <template v-else>
            {{ df.format(modelValue.start.toDate(getLocalTimeZone())) }}
          </template>
        </template>
        <template v-else>
          Pick a date
        </template>
      </UButton>

      <template #content>
        <UCalendar
          v-model="range"
          class="p-2"
          :number-of-months="2"
          range
        />
      </template>
    </UPopover>

    <TipTap

      ref="tiptap"
      v-model="entry.content.markdown"
      bubble-menu="link,B,I,S,U,H,C,sub,sup"
      floating-menu="h2,h3,h4,ul,ol"
      class="prose prose-sm max-w-none p-4 font-inter"
    />
  </article>
</template>

<script setup lang="ts">
import { CalendarDate, DateFormatter, getLocalTimeZone } from '@internationalized/date'

const props = defineProps<{
  entry: PersonEntry
  variant?: 'standalone' | 'embed' | 'modal' | 'listitem'
}>()
const emit = defineEmits<{
  'update:entry': [PersonEntry]
  'patch:entry': [key: string, value: string | number | boolean | object]
}>()
const { variant = 'standalone' } = props // default value
const visible = computed(() => ({
  commons: variant === 'embed' ? false : true,
}))

function update(key: string, value: string) {
  emit('update:entry', cloneSetPath(props.entry, key, value) as PersonEntry)
}

function patch(key: string, value: string) {
  console.log('patch:', key, value)
  emit('patch:entry', key, value)
}

function openEditModal() {
  // open modal
  edit.value = !edit.value
}

const edit = ref(false)
const open = ref(false)

const df = new DateFormatter('de-DE', {
  dateStyle: 'medium',
})

const range = computed({
  get: () => {
    return {
      start: new CalendarDate(2022, 1, 20),
      end: new CalendarDate(2022, 2, 10),
    }
  },
  set: (val: object) => {
    // emit('update:entry', cloneSetPath(props.entry, 'content.from', val.start) as ArticleEntry)
  },
})

const modelValue = shallowRef({
  start: new CalendarDate(2022, 1, 20),
  end: new CalendarDate(2022, 2, 10),
})
</script>
