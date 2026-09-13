<template>
  <div v-if="account" class="flex flex-col gap-8 pt-6">
    <AccountForm :account="account" @saved="load" />
    <AccountAccess :account="account" @changed="load" />
  </div>

  <UAlert v-else-if="error" class="mt-6" color="error" variant="soft" :description="error" />
</template>

<script setup lang="ts">
// One account, in two halves: who it is and what it may see, then what it can
// sign in with. They are kept apart because revoking a credential should not be
// something that happens while correcting a name.
const admin = useAdminAccounts()
const name = useRoute().params.name as string

const account = ref<AdminAccount | null>(null)
const error = ref('')

async function load() {
  try {
    account.value = await admin.get(name)
  }
  catch (e) {
    error.value = admin.describe(e)
  }
}

onMounted(load)
</script>
