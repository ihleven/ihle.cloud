<template>
  <div>
    <h1>{{ entry.path }}</h1>
    <div
      v-if="entry"
    >
      <MDC
        v-if="markdown"
        :value="markdown"
        tag="article"
        class="prose prose-sm"
      />
      <aside class="float-right rounded bg-gray-300 p-4">
        {{ entry.content }}
      </aside>
    </div>

    <!-- {{ key }} -->

    <!-- <ContentRenderer
      v-if="person"
      :value="person"
    />
    <div v-else>
      Home not found {{ params }} p: {{ path }}
    </div> -->
  </div>
</template>

<script setup lang="ts">
const props = defineProps<{
  entry: PersonEntry
}>()

const markdown = computed(() => props.entry?.content?.markdown ? props.entry.content.markdown : '')
</script>
