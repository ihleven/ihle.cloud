<template>
  <UPopover :content="{ align: 'end' }" :ui="{ content: 'ght-panel w-48' }">
    <button
      class="inline-flex items-center gap-1.5 rounded bg-black/50 px-2 py-1 text-sm font-medium text-white hover:bg-black/60 focus:outline-none"
    >
      <span>{{ login }}</span>
      <UIcon name="i-heroicons-user-circle" class="h-4 w-4" />
    </button>

    <template #content>
      <!-- The inset is on the rows, not the panel, so every line starts at the
           same place and the sign-out's hover reaches the panel's edges. -->
      <div class="grid gap-2 py-2">
        <h3 class="px-2 text-xs font-light text-gray-200">{{ login }}</h3>

        <div v-if="account" class="px-2 text-sm text-gray-300">
          <p class="truncate">{{ account.vorname }} {{ account.nachname }}</p>
          <p class="truncate text-xs text-gray-400">{{ account.email }}</p>
        </div>

        <button
          class="flex items-center justify-start gap-1.5 px-2 py-1 text-sm text-gray-300 hover:bg-white/10 hover:text-white"
          @click="signOutAndLeave"
        >
          <UIcon name="i-heroicons-power" class="h-4 w-4" />
          abmelden
        </button>
      </div>
    </template>
  </UPopover>
</template>

<script setup lang="ts">
// Who is signed in to the pool, and the way out again.
//
// This belongs to the account rather than to any edition, which is why it sits
// in the header above and stays the same on every page of the pool. The name and
// address are the account's own — /aktuell only ever answers about its caller —
// so they are shown to the one person entitled to see them. The pool's version
// also had a disabled "Daten ändern" button, which is not carried over.
const { login, account, signOut } = useGhtSession()

async function signOutAndLeave() {
  await signOut()
  // The front page, because that is where signing in happens now: one form for
  // both, so there is no separate pool sign-in to come back to.
  await navigateTo('/')
}
</script>
