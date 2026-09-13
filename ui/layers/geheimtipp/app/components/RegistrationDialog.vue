<template>
  <UModal v-model:open="model" title="Registrierung" description="Ohne Registrierung keine Tipps">
    <template #content>
      <div class="p-4">
        <p class="p-2 text-lg text-gray-500">
          Ohne
          <NuxtLink to="/geheimtipp/register" class="underline decoration-sky-300 underline-offset-4">
            Registrierung
          </NuxtLink>
          können keine Tipps abgegeben werden.
        </p>
        <div class="flex w-full justify-end border-t pt-2">
          <NuxtLink
            to="/geheimtipp/register"
            class="rounded-md border border-sky-300 px-4 py-2 text-sm font-medium text-sky-300 hover:border-transparent hover:bg-sky-300 hover:text-white"
            @click="model = false"
          >
            Zur Registrierung
          </NuxtLink>
        </div>
      </div>
    </template>
  </UModal>
</template>

<script setup lang="ts">
// Signing in to the pool and being registered for the current edition are two
// different things: an account that has not registered can read everything and
// tip nothing. This is what says so.
const model = defineModel<boolean>('open', { default: false })
</script>
