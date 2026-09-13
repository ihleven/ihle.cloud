<template>
  <div>
    <ul v-if="accounts.length" class="divide-y divide-gray-200">
      <li v-for="account in accounts" :key="account.name" class="flex items-center gap-4 py-3">
        <div class="grow">
          <NuxtLink :to="`/admin/${account.name}`" class="text-sm font-medium hover:underline">
            {{ account.display_name || account.name }}
          </NuxtLink>
          <p class="text-xs text-gray-500">
            {{ account.name }} · {{ account.email }}
          </p>
        </div>

        <div class="flex items-center gap-2">
          <UBadge v-if="account.disabled" color="error" variant="subtle" size="sm">gesperrt</UBadge>
          <!-- An account with neither cannot sign in at all, which is the state a
               freshly created one is in and the thing an enrollment link fixes. -->
          <UBadge v-if="!account.has_password && !account.passkeys" color="warning" variant="subtle" size="sm">
            noch keine Anmeldung
          </UBadge>
          <UBadge v-for="area in account.modules" :key="area" variant="subtle" size="sm">{{ area }}</UBadge>
        </div>
      </li>
    </ul>
    <p v-else-if="!pending" class="py-6 text-sm text-gray-500">Noch keine Konten.</p>

    <form class="mt-6 flex flex-wrap items-end gap-3 border-t border-gray-200 pt-6" @submit.prevent="add">
      <UFormField label="Anmeldename" class="grow">
        <UInput v-model="draft.name" placeholder="wolfgang" autocomplete="off" class="w-full" />
      </UFormField>
      <UFormField label="Name" class="grow">
        <UInput v-model="draft.display_name" placeholder="Wolfgang Ihle" autocomplete="off" class="w-full" />
      </UFormField>
      <UFormField label="E-Mail" class="grow">
        <UInput v-model="draft.email" type="email" placeholder="wolfgang@example.de" autocomplete="off" class="w-full" />
      </UFormField>
      <UButton type="submit" :loading="busy">Anlegen</UButton>
    </form>
    <p class="mt-2 text-xs text-gray-500">
      Der Anmeldename lässt sich später nicht ändern — Inhalte gehören ihm.
      Anmelden kann sich das neue Konto erst mit einem Einladungslink.
    </p>

    <UAlert v-if="error" class="mt-4" color="error" variant="soft" :description="error" />
  </div>
</template>

<script setup lang="ts">
const admin = useAdminAccounts()

const accounts = ref<AdminAccount[]>([])
const draft = reactive({ name: '', display_name: '', email: '' })
const busy = ref(false)
const pending = ref(true)
const error = ref('')

async function refresh() {
  try {
    accounts.value = await admin.list()
  }
  catch (e) {
    error.value = admin.describe(e)
  }
  finally {
    pending.value = false
  }
}

async function add() {
  busy.value = true
  error.value = ''
  try {
    const created = await admin.create({ ...draft })
    draft.name = draft.display_name = draft.email = ''
    // Straight to the new account: it cannot sign in yet, and issuing the link
    // is what finishes creating it.
    await navigateTo(`/admin/${created.name}`)
  }
  catch (e) {
    error.value = admin.describe(e)
  }
  finally {
    busy.value = false
  }
}

onMounted(refresh)
</script>
