<template>
  <div class="text-sm">
    <template v-if="session">
      <p class="font-semibold text-highlighted">{{ session.name }}</p>
      <p class="text-xs break-words text-muted">{{ session.email }}</p>

      <dl class="mt-4 space-y-1 text-xs text-muted">
        <div>
          <dt class="inline">Angemeldet als</dt>
          <dd class="ml-1 inline font-medium">{{ session.sub }}</dd>
        </div>
        <div>
          <dt>Sitzung gültig bis</dt>
          <dd>{{ formatTime(session.expiry) }} <span class="text-dimmed">(noch {{ formatDuration(session.expires_in) }})</span></dd>
        </div>
      </dl>

      <UButton class="mt-4 cursor-pointer" color="neutral" variant="subtle" label="Logout" @click="logout" />
    </template>

    <template v-else>
      <p class="text-muted">Nicht angemeldet.</p>
    </template>
  </div>
</template>

<script setup lang="ts">
// The account details on their own, so they can be shown inline — in the menu
// there is nothing to drop down from, and a dropdown inside a panel is one
// layer too many.
const { session, logout } = await useAuth()
</script>
