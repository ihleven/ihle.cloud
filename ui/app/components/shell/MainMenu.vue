<template>
  <!-- Where you can go on one side, who you are on the other. Signing in and out
       lives here rather than in the bar: it is a rare action, and the bar is for
       where you are going.

       Side by side where there is room. On a phone the panel is the whole screen
       but only ~180px a column, which is not enough for an email address, so the
       two stack instead. -->
  <div class="grid grid-cols-1 divide-y divide-default sm:grid-cols-2 sm:divide-x sm:divide-y-0">
    <UNavigationMenu :items="items" orientation="vertical" class="pb-4 sm:pb-0 sm:pr-4" />
    <AuthPanel class="pt-4 sm:pt-0 sm:pl-4" />
  </div>
</template>

<script setup lang="ts">
const { items } = useAreas()
</script>
