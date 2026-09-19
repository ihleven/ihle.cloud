<template>
  <!-- Where you can go on one side, who you are on the other. Signing in and out
       lives here rather than in the bar: it is a rare action, and the bar is for
       where you are going.

       Side by side where there is room. On a phone the panel is the whole screen
       but only ~180px a column, which is not enough for an email address, so the
       two stack instead. -->
  <div>
    <div class="grid grid-cols-1 divide-y divide-default sm:grid-cols-2 sm:divide-x sm:divide-y-0">
      <UNavigationMenu :items="items" orientation="vertical" class="pb-4 sm:pr-4 sm:pb-0" />
      <AuthPanel class="pt-4 sm:pt-0 sm:pl-4" />
    </div>

    <!-- Only where there is no other way to do it. Installed, the app has no
         address bar to reload from and no pull-to-reload gesture, so a page
         that has got itself into a state can only be fixed by closing the app
         entirely — which starts it again at the front page rather than where
         you were. In a browser tab this would be a button next to the browser's
         own, so it is not offered there. -->
    <div v-if="standalone" class="mt-4 border-t border-default pt-4">
      <UButton
        variant="ghost"
        color="neutral"
        icon="i-lucide-rotate-cw"
        label="Neu laden"
        class="w-full justify-start"
        @click="reload"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
const { items } = useAreas()
const { standalone } = useStandalone()

// The page as it is, not the front page: reloading is for getting out of a
// state, and being sent home would lose where you were as surely as closing the
// app does.
function reload() {
  window.location.reload()
}
</script>
