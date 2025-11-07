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
      {{ session?.sub || 'anonym' }}
    </UButton>
    <template #content>
      <div v-if="session" class="m-4 inline-flex w-fit flex-col items-end">
        <div class="text-sm font-semibold text-gray-500">
          {{ session.name }}
        </div>
        <small class="text-xs font-light text-gray-500">{{ session.email }}</small>
        <div class="mr-auto pt-4 text-xs text-gray-500">
          Session valid until: <br>
        </div>
        <div class="mt-1 text-xs text-gray-500">
          {{ formatTime(session.expiry) }}  (in {{ formatDuration(session.expires_in) }})
          <br>
          <!-- <small class="text-gray-400">(aktuell nicht identisch mit cookie expiry)</small> -->
        </div>
        <UButton
          color="neutral"
          variant="subtle"
          size="md"
          class="mt-4 rounded"
          @click="logout"
        >
          Logout
        </UButton>
      </div>
    </template>
    <template #content2>
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
const { session, login, logout } = await useAuth()
</script>
