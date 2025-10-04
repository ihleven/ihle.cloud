<template>
  <UModal
    :close="{ onClick: () => emit('close', false) }"
    title="Link" description="Set link properties"
    :ui="{ wrapper: 'flex flex-col wrapper', body: 'bg-gray-100' }"
  >

    <template #body>
      <div class="grid grid-cols-2 gap-x-4 gap-y-2">
        <UFormField label="href" class="col-span-2 w-full">
          <UInput v-model="link.href" placeholder="href" class="w-full" />
        </UFormField>
        <UFormField label="title" class="col-span-2 w-full">
          <UInput v-model="link.title" class="w-full" />
        </UFormField>
        <UFormField label="target" class="">
          <!-- <UInput v-model="link.target" class="w-full" /> -->
          <USelect
            v-model="link.target" :items="[{
              label: 'gleicher Tab / Seite (default)',
              value: '_self',
            }, {
              label: 'neuer Tab oder Fenster',
              value: '_blank',
            }, {
              label: 'Elternfenster',
              value: '_parent',
            }, {
              label: 'ganzes Fenster',
              value: '_top',
            }]" class="w-full"
            label="target"
          />
        </UFormField>
        <UFormField label="id" class="ml-auto w-full">
          <UInput v-model="link.id" class="w-full" />
        </UFormField>
        <UFormField label="rel">
          <UInput v-model="link.rel" class="w-full" />
        </UFormField>
        <UFormField label="class">
          <UInput v-model="link.class" class="w-full" />
        </UFormField>
      </div>
    </template>

    <template #footer>
      <UButton label="Cancel" color="neutral" variant="outline" @click="() => emit('close', false)" />
      <UButton label="Delete link" color="primary" variant="outline" class="ml-auto" @click="() => emit('close', null)" />
      <UButton label="Set link" color="primary" @click="() => emit('close', link)" />
    </template>
  </UModal>
</template>

<script setup lang="ts">
type link = {
  href?: string
  rel?: string
  target?: string
  id?: string
  class?: string
  title?: string
}
const props = defineProps<{
  href?: string
  rel?: string
  target?: string
  id?: string
  class?: string
  title?: string
}>()

const emit = defineEmits<{
  close: [link | null | boolean]
}>()

const link = ref<link>(Object.assign({}, props))

// function updateRel(value: string): void {
//   if (value) {
//     link.value.rel = value
//   }
//   else {
//     link.value.rel = null
//   }
// }
</script>
