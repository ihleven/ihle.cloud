<template>
  <article class="flex h-screen w-screen items-center justify-center bg-amber-400">
    <div class="w-96 rounded bg-white p-6 shadow">
      <h1 class="text-lg font-semibold">Register a device</h1>
      <p v-if="first" class="mt-2 text-sm text-gray-600">
        This link adds a passkey to your account. Choose a password to go with
        it — it is what you will need to add or remove a device later.
      </p>
      <p v-else class="mt-2 text-sm text-gray-600">
        This link lets you add a passkey to your account. Your password is
        needed as well, so that opening the link alone is not enough.
      </p>

      <UAlert
        v-if="reason"
        class="mt-4" color="error" variant="soft"
        title="Passkeys are not available here" :description="reason"
      />

      <form v-else class="mt-4 flex flex-col gap-3" @submit.prevent="submit">
        <UInput
          v-model="password"
          type="password"
          :placeholder="first ? 'Choose a password' : 'Your password'"
          :autocomplete="first ? 'new-password' : 'current-password'"
        />
        <UInput v-model="name" placeholder="Name for this device, e.g. Laptop" />
        <UButton type="submit" :loading="busy" block>Register this device</UButton>
      </form>

      <UAlert v-if="error" class="mt-4" color="error" variant="soft" :description="error" />
      <UAlert v-if="done" class="mt-4" color="success" variant="soft"
        description="Registered. You can now sign in with this device." />
    </div>
  </article>
</template>

<script setup lang="ts">
// Reachable without a session: whoever follows an enrollment link has an account
// but has not signed in, and the link carries an enrollment cookie rather than a
// session. Without this the app's sign-in overlay would cover the page they were
// sent to.
definePageMeta({ public: true })

// The enrollment link lands on /auth/enroll, which moves the token into a
// cookie and redirects here — so the token is not in this page's URL, and this
// page never sees it.
const passkey = usePasskey()
const { loadSession } = useAuth()

const password = ref('')
const name = ref('')
const busy = ref(false)
const done = ref(false)
const error = ref('')
const reason = ref('')

// Whether this account is choosing its first password or proving the one it
// has. Asked of the server, because the page cannot tell from the link — and
// asking the wrong question means a form that says "your password" to someone
// who does not have one yet.
const first = ref(false)

onMounted(async () => {
  reason.value = passkey.unavailable()
  try {
    const state = await $fetch<{ has_password: boolean }>('/auth/enroll/state', {
      baseURL: useRuntimeConfig().public.api.base as string,
      credentials: 'include',
    })
    first.value = !state.has_password
  }
  catch {
    // Not knowing is not worth blocking on: the form still works, it just asks
    // the more common question.
  }
})

async function submit() {
  busy.value = true
  error.value = ''
  try {
    await passkey.enrol(password.value, name.value)
    done.value = true
    password.value = ''
    // Finishing signs the account in, so pick that session up.
    await loadSession()
    setTimeout(() => navigateTo('/'), 1500)
  }
  catch (e) {
    error.value = passkey.describe(e)
  }
  finally {
    busy.value = false
  }
}
</script>
