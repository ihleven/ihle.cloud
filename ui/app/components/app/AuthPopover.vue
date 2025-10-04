<template>
  <UPopover

    mode="hover"
  >
    <UButton
      :icon="session ? 'i-lucide-user-circle' : 'i-lucide-log-in'"
      size="md"
      color="neutral"
      variant="ghost"
    >
      {{ session?.sub }}
    </UButton>

    <template #content>
      <div class="m-4 inline-flex w-fit flex-col items-end">
        <small class="text-xs font-light text-gray-500">{{ session?.sub }}</small>
        <div class="text-sm font-semibold text-gray-500">
          {{ session?.name }}
        </div>
        <small class="text-xs font-light text-gray-500">{{ session?.email }}</small>
        <UBadge v-for="a in session?.aud" :key="a" class="mt-2 px-1 py-0">{{ a }}</UBadge>
        <div class="py-4 text-xs text-gray-500">
          Session expiry:
          {{ formatDuration(session?.expires_in) }}
          <br>

        </div>

        <UButton

          size="md"
          color="neutral"
          variant="solid"
          @click="logout"
        >
          Logout
        </UButton>
      </div>
    </template>
  </UPopover>
</template>

<script setup lang="ts">
const { session, login, logout, expiry, remaining } = await useAuth()

function formatDuration(d: number) {
  if (d < 60)
    return Math.floor(d / 5) * 5 + 's'
  if (d < 3600)
    return Math.floor(d / 60) + 'm' + Math.floor((d % 60) / 5) * 5 + 's'
  return Math.floor(d / 3600) + 'h' + Math.floor((d % 3600) / 60) + 'm'
}
</script>
