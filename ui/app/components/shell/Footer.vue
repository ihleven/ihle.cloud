<template>
  <UFooter :ui="{ root: 'bg-ral-7032' }">
    <!-- The name lives here as well as in the bar, so it is on every page. -->
    <template #left>
      <div class="flex items-baseline gap-3">
        <Wordmark />
        <p class="text-sm text-muted">
          Copyright © {{ new Date().getFullYear() }}
        </p>
      </div>
    </template>

    <UNavigationMenu :items="items" variant="link" />
  </UFooter>
</template>

<script setup lang="ts">
const { items } = useAreas()
</script>
