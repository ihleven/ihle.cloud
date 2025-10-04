<template>
  <UCollapsible :open="open">
    <template #content>
      <pre class="overflow-auto rounded bg-red-100 p-4 font-mono text-xs">{{ entry }}</pre>
    </template>
  </UCollapsible>
</template>

<script lang="ts" setup>
defineProps<{
  entry: Entry
  open: boolean
}>()
</script>
