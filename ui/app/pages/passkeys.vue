<template>
  <article class="mx-auto max-w-2xl p-6">
    <h1 class="text-lg font-semibold">Your devices</h1>
    <p class="mt-2 text-sm text-gray-600">
      A passkey signs you in without a password. Registering or removing one
      needs your password, so a stolen session cannot quietly add a device.
    </p>

    <ul v-if="keys.length" class="mt-4 divide-y divide-gray-200 border-y border-gray-200">
      <li v-for="key in keys" :key="key.id" class="flex items-center gap-4 py-3">
        <div class="grow">
          <p class="text-sm font-medium">{{ key.name || 'Unnamed device' }}</p>
          <p class="text-xs text-gray-500">
            Added {{ formatDate(key.created_at) }} ·
            {{ key.last_used_at ? `last used ${formatDate(key.last_used_at)}` : 'never used' }} ·
            <!-- Whether the credential can leave the device it was made on. Both
                 kinds are accepted; showing which is which is why it is stored. -->
            {{ key.syncable ? 'synced between devices' : 'this device only' }}
          </p>
        </div>
        <UButton
          color="error" variant="ghost" size="xs"
          :loading="removing === key.id"
          @click="remove(key.id)"
        >
          Remove
        </UButton>
      </li>
    </ul>
    <p v-else class="mt-4 text-sm text-gray-500">No devices registered yet.</p>

    <form class="mt-6 flex flex-col gap-3" @submit.prevent="add">
      <h2 class="text-base font-medium">Add this device</h2>
      <UInput v-model="password" type="password" placeholder="Your password" autocomplete="current-password" />
      <UInput v-model="name" placeholder="Name for this device, e.g. Laptop" />
      <UButton type="submit" :loading="busy" :disabled="!!reason">Register</UButton>
      <p v-if="reason" class="text-xs text-gray-600">{{ reason }}</p>
    </form>

    <UAlert v-if="error" class="mt-4" color="error" variant="soft" :description="error" />
  </article>
</template>

<script setup lang="ts">
const passkey = usePasskey()

const keys = ref<Awaited<ReturnType<typeof passkey.list>>>([])
const password = ref('')
const name = ref('')
const busy = ref(false)
const removing = ref('')
const error = ref('')
const reason = ref('')

function formatDate(value: string): string {
  return new Date(value).toLocaleDateString()
}

async function refresh() {
  try {
    keys.value = await passkey.list()
  }
  catch (e) {
    error.value = passkey.describe(e)
  }
}

async function add() {
  busy.value = true
  error.value = ''
  try {
    await passkey.enrol(password.value, name.value)
    password.value = ''
    name.value = ''
    await refresh()
  }
  catch (e) {
    error.value = passkey.describe(e)
  }
  finally {
    busy.value = false
  }
}

onMounted(() => {
  reason.value = passkey.unavailable()
  refresh()
})

// Removing asks for the password too: taking a device away is as sensitive as
// adding one, and the field above is where it is already typed.
async function remove(id: string) {
  if (!password.value) {
    error.value = 'Enter your password below to remove a device.'
    return
  }
  removing.value = id
  error.value = ''
  try {
    await passkey.remove(id, password.value)
    await refresh()
  }
  catch (e) {
    error.value = passkey.describe(e)
  }
  finally {
    removing.value = ''
  }
}
</script>
