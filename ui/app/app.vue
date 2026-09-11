<template>
  <UApp :toaster="toaster">

    <NuxtLayout v-if="session">
      <NuxtPage />
    </NuxtLayout>
    <div v-else class="absolute inset-0 bg-[url(/Summer-Leaves.jpg)]" />

    <UModal
      :open="!session"
      title="Login" 
      description="Login modal"
      :ui="{ content: 'rounded p-4', overlay: 'bg-default/25 backdrop-blur-md' }"
    >
      <template #content>

        <UAuthForm

          title="Login"
          description="Enter your credentials to access your account."
          icon="i-lucide-user"
          :fields="fields"
          @submit="onSubmit"
        />

        <!-- Usernameless: the authenticator says which credential it holds, so
             there is nothing to type. Shown even where it cannot be used, with
             the reason, because a control that vanishes explains nothing. -->
        <USeparator label="or" class="my-4" />
        <UButton
          block
          color="neutral"
          variant="subtle"
          icon="i-lucide-fingerprint"
          :loading="passkeyBusy"
          :disabled="!!passkeyReason"
          @click="signInWithPasskey"
        >
          Use a device
        </UButton>
        <p v-if="passkeyReason" class="mt-2 text-xs text-muted">{{ passkeyReason }}</p>
        <p v-else-if="passkeyError" class="mt-2 text-xs text-error">{{ passkeyError }}</p>

      </template>
    </UModal>

  </UApp>
</template>

<script setup lang="ts">
import * as z from 'zod'

import type { FormSubmitEvent, AuthFormField } from '@nuxt/ui'

const { toaster } = useAppConfig()
const { session, loginajax, loadSession } = useAuth()
// await loadSession()

const passkey = usePasskey()
const passkeyBusy = ref(false)
const passkeyError = ref('')
const passkeyReason = ref('')

// Holds the background offer so it can be called off: only one credential
// request may be outstanding, so pressing the button or submitting the password
// has to cancel it first.
let conditional: AbortController | undefined

function cancelConditional() {
  conditional?.abort()
  conditional = undefined
}

// Evaluated after mount: these depend on browser globals, which are not there
// while the component is being set up.
onMounted(async () => {
  passkeyReason.value = passkey.unavailable()
  if (passkeyReason.value) return

  if (await passkey.conditionalAvailable()) {
    offerPasskeyInAutofill()
  }
})

onBeforeUnmount(cancelConditional)

// Started in the background and left pending: the browser resolves it only if
// the person picks a passkey from the username field, and never prompts on its
// own. If they have no passkey for this site, nothing happens at all.
async function offerPasskeyInAutofill() {
  conditional = new AbortController()
  try {
    await passkey.signIn('conditional', conditional.signal)
    await loadSession()
  }
  catch (e) {
    if (!passkey.aborted(e)) passkeyError.value = passkey.describe(e)
  }
}

async function signInWithPasskey() {
  cancelConditional()
  passkeyBusy.value = true
  passkeyError.value = ''
  try {
    await passkey.signIn()
    await loadSession()
  }
  catch (e) {
    passkeyError.value = passkey.describe(e)
  }
  finally {
    passkeyBusy.value = false
  }
}

const fields: AuthFormField[] = [{
  name: 'username',
  type: 'text',
  label: 'Username',
  placeholder: 'Enter your username',
  required: true,
  // The token the browser needs in order to offer a passkey inline here rather
  // than in a dialog you have to ask for.
  autocomplete: 'username webauthn',
}, {
  name: 'password',
  label: 'Password',
  type: 'password',
  placeholder: 'Enter your password',
  required: true,
}, {
  name: 'remember',
  label: 'Remember me',
  type: 'checkbox',
}]

const schema = z.object({
  email: z.string('Invalid email'),
  password: z.string('Password is required').min(3, 'Must be at least 3 characters'),
})

type Schema = z.output<typeof schema>

function onSubmit(payload: FormSubmitEvent<Schema>) {
  cancelConditional()
  loginajax(payload.data.username, payload.data.password)
}
</script>
