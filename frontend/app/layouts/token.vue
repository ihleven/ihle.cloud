<template>
  <div>
    <slot />
    <UModal
      v-model:open="open"
      :dismissible="false"
      :fullscreen="false"
    >
      <template #content>
        <div class="p-4">
          <UFormField>
            <UInput
              v-model="password"
              placeholder="Password"
              icon="i-lucide-at-sign"
              :type="show ? 'text' : 'password'"
              :ui="{ trailing: 'pe-1' }"
              :aria-invalid="false"
              aria-describedby="password-strength"
              class="w-full"
              size="xl"
              color="neutral"
            >
              <template #trailing>
                <UButton
                  color="neutral"
                  variant="link"
                  size="sm"
                  :icon="show ? 'i-lucide-eye-off' : 'i-lucide-eye'"
                  aria-label="show ? 'Hide password' : 'Show password'"
                  :aria-pressed="show"
                  aria-controls="password"
                  @click="show = !show"
                />
              </template>
            </UInput>
          </UFormField>
        </div>
      </template>
    </UModal>
  </div>
</template>

<script setup lang="ts">
const open = ref(true)

defineShortcuts({
  x: () => open.value = !open.value,
})
const show = ref(true)
const password = ref('')
</script>
