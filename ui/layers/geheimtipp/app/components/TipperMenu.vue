<template>
  <UDropdownMenu :items="items" :ui="{ content: 'w-48' }">
    <button class="flex rounded-full bg-gray-800 focus:ring-2 focus:ring-white focus:ring-offset-2 focus:ring-offset-gray-800 focus:outline-none">
      <img class="h-8 w-8 rounded-full" :src="ghtAvatar(tipper.avatar, 64)" :alt="tipper.name">
    </button>
  </UDropdownMenu>
</template>

<script setup lang="ts">
import type { EditionTipper } from '../types'

// The registration for the edition being looked at: the name, picture and motto
// this person plays under here, which are theirs to change and are not the same
// thing as the account they signed in with. Hence its place in the edition's own
// bar rather than in the header, and hence no way out of the pool from here —
// that is the account's business, one row up.
const props = defineProps<{ tipper: EditionTipper }>()

// Three groups rather than three items: the separator a group boundary draws is
// what the pool has as a bottom border on each of the first two rows.
const items = computed(() => [
  [{ label: props.tipper.name, type: 'label' as const }],
  [{ label: props.tipper.motto, type: 'label' as const }],
  // Their label reads "zur Registierung" — a typo, corrected here.
  [{ label: 'zur Registrierung', to: '/geheimtipp/register' }],
])
</script>
