<template>
  <article class="bg-gray-100">
    ENTRY
    <ClientOnly>
      <!-- <section class="relative">
      <h1 class=" top-2 left-2 rounded border border-white p-2 mx-auto bg-white/70 text-gray-700 text-xl">
        ENTRY ({{ entry.type }})
      </h1>
      <pre>{{ entry.content }}</pre>
    </section> -->

      <!-- <json-editor
        v-model:json="model"
        class=""
        :main-menu-bar="false"
        :navigation-bar="false"
        mode="tree"
        :query-languages-ids="queryLanguages"
        @error="onError"
        @focus="onFocus"
        @blur="onBlur"
      /> -->
    </ClientOnly>
    {{ entry }}
  </article>
</template>

<script setup lang="ts">
import { cloneDeep } from 'lodash'

const props = defineProps<{
  entry: object
}>()
const emit = defineEmits<{ 'update:entry': [object] }>()

const model = computed({
  get: () => props.entry,
  set(obj) {
    emit('update:entry', cloneDeep(obj))
  },
})
const mode = ref('tree')

const queryLanguages = ref<QueryLanguageId[]>(['javascript', 'lodash', 'jmespath'])

const onError = (error) => {
  //
}

const onFocus = () => {
  //
}

const onBlur = () => {
  //
}
</script>
