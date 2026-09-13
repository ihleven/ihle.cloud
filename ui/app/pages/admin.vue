<template>
  <section class="mx-auto w-full max-w-4xl p-6">
    <header class="flex items-baseline justify-between border-b border-gray-200 pb-3">
      <div>
        <h1 class="text-lg font-semibold">Konten</h1>
        <p class="mt-1 text-sm text-gray-600">
          Wer sich anmelden kann, und was jede und jeder zu sehen bekommt.
        </p>
      </div>
      <NuxtLink
        v-if="route.path !== '/admin'"
        to="/admin"
        class="text-sm text-gray-500 hover:text-gray-900"
      >
        &lt; Alle Konten
      </NuxtLink>
    </header>

    <NuxtPage />
  </section>
</template>

<script setup lang="ts">
// The section shell. The area guard in middleware/modules.global.ts keeps
// accounts without the entitlement out of /admin; this page assumes it ran.
const route = useRoute()
</script>
