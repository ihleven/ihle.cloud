<template>
  <article
    v-if="variant==='listitem'"
    class="border-0 border-accented p-1 grow"
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
    <div class="grid grid-cols-12 gap-4 justify-stretch">
      <UFormField
        label="Titel:"
        class="col-span-9"
      >
        <UInput
          placeholder="Titel"
          :model-value="entry.content.title"
          class="w-full"
          @update:model-value="update('content.title', $event)"
        />
      </UFormField>

      <UFormField
        label="Teile:"
        class="col-span-3"
      >
        <USelect
          :model-value="entry.content.teile"
          class="w-full"
          @update:model-value="update('content.teile', $event)"
        />
      </UFormField>

      <UFormField
        label="Jahr:"
        class="col-span-4"
      >
        <UInputNumber
          placeholder="Jahr ..."
          :model-value="entry.content.year"
          :min="1970"
          :max="2030"
          :format-options="{
            minimumIntegerDigits: 4,
            maximumFractionDigits: 0,
            useGrouping: false,
          }"
          class="w-full"
          @update:model-value="update('content.year', $event)"
        />
      </UFormField>

      <UFormField
        label="Schaffensphase:"
        class="col-span-4"
      >
        <!-- -->

        <USelect
          :model-value="entry.content.phase"
          :items="[{ value: '-', label: '---' }, 'Frühwerk', 'Natur I Landschaft, Figur', 'Natur II Abstraktion', 'Entgegenständlichung', 'Monochrome Malerei']"
          class="w-full"
          @update:model-value="update('content.phase', $event)"
        />
      </UFormField>

      <UFormField
        :label="'Gattung:'"
        class="col-span-4"
      >
        <USelect
          :model-value="entry.content.gattung"
          :items="[{ M: 'Malerei' }, { Z: 'Zeichnung' }, { P: 'Plastik' }]"
          class="w-full"
          @update:model-value="update('content.gattung', $event)"
        />
      </UFormField>

      <UFormField
        label="Technik:"
        class="col-span-6"
      >
        <!-- -->
        <USelect
          :model-value="entry.content.medium"
          :items="['Öl', 'Aquarell', 'Acryl', 'Pastell']"
          class="w-full"
          @update:model-value="update('content.medium', $event)"
        />
      </UFormField>
      <UFormField
        label="Träger:"
        class="col-span-6"
      >
        <USelect
          :model-value="entry.content.support"
          :items="['Leinwand', 'Papier', 'Kupfer', 'Holz']"
          class="w-full"
          @update:model-value="update('content.support', $event)"
        />
      </UFormField>

      <UFormField
        label="Höhe:"
        class="col-span-3"
      >
        <UInputNumber
          placeholder="Höhe ..."
          :model-value="entry.content.height"
          :min="0"
          :max="200"
          :format-options="{
            maximumFractionDigits: 0,
            useGrouping: false,
          }"
          class="w-full"
          @update:model-value="update('content.height', $event)"
        />
      </UFormField>

      <UFormField
        label="Breite:"
        class="col-span-3"
      >
        <UInputNumber
          placeholder="Breite ..."
          :model-value="entry.content.width"
          :min="0"
          :max="200"
          :format-options="{
            maximumFractionDigits: 0,
            useGrouping: false,
          }"
          class="w-full"
          @update:model-value="update('content.width', $event)"
        />
      </UFormField>

      <UFormField
        label="Tiefe:"
        class="col-span-3"
      >
        <UInputNumber
          placeholder="in cm"
          :model-value="entry.content.depth"
          :min="0"
          :max="200"
          :format-options="{
            maximumFractionDigits: 0,
            useGrouping: false,
          }"
          class="w-full"
          @update:model-value="update('content.depth', $event)"
        />
      </UFormField>

      <UFormField
        label="Anmerkungen:"
        class="col-span-12"
      >
        <UTextarea
          placeholder="Anmerkungen"
          :model-value="entry.content.remarks"
          class="w-full"
          @update:model-value="update('content.remarks', $event)"
        />
      </UFormField>

      <UFormField
        label="Kommentar:"
        class="col-span-12"
      >
        <UTextarea
          placeholder="Kommentar"
          :model-value="entry.content.comments"
          class="w-full"
          @update:model-value="update('content.comments', $event)"
        />
      </UFormField>
    </div>
  </article>
</template>

<script setup lang="ts">
const props = defineProps<{
  entry: ArtworkEntry
  variant?: 'standalone' | 'embed' | 'modal' | 'listitem'
}>()
const emit = defineEmits<{
  'update:entry': [ArtworkEntry]
  'patch:entry': [key: string, value: string | number | boolean | object]
}>()
const { variant = 'standalone' } = props // default value
const visible = computed(() => ({
  commons: variant === 'embed' ? false : true,
}))

function update(key: string, value: string) {
  emit('update:entry', cloneSetPath(props.entry, key, value) as ArtworkEntry)
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
</script>
