<template>
  <div class="w-full text-white">
    <header class="mb-6 text-center">
      <Wordmark class="!text-white [&_*]:!text-white" />
      <p class="mt-2 text-sm text-white/60">Bitte anmelden</p>
    </header>

    <form class="space-y-3" @submit.prevent="submit">
      <UInput
        v-model="username"
        size="xl"
        variant="none"
        icon="i-lucide-user"
        placeholder="Benutzername"
        autocomplete="username webauthn"
        autofocus
        :class="field"
        :ui="inputUi"
      />
      <UInput
        v-model="password"
        size="xl"
        variant="none"
        type="password"
        icon="i-lucide-lock"
        placeholder="Passwort"
        autocomplete="current-password"
        :class="field"
        :ui="inputUi"
      />

      <UButton
        type="submit"
        block
        size="xl"
        variant="none"
        :loading="busy"
        :disabled="!username || !password"
        class="mt-1 cursor-pointer justify-center rounded-xl bg-white/90 font-medium text-black transition hover:bg-white hover:shadow-lg hover:shadow-white/20 disabled:bg-white/25 disabled:text-white/50 disabled:shadow-none"
      >
        Anmelden
      </UButton>
    </form>

    <p v-if="error" class="mt-3 text-center text-sm text-red-300">{{ error }}</p>

    <!-- Usernameless: the authenticator says which credential it holds, so
         there is nothing to type. Shown even where it cannot be used, with
         the reason, because a control that vanishes explains nothing. -->
    <USeparator label="oder" class="my-5" :ui="{ label: 'text-white/40', border: 'border-white/15' }" />
    <UButton
      block
      size="xl"
      variant="none"
      icon="i-lucide-fingerprint"
      :loading="passkeyBusy"
      :disabled="!!reason"
      class="cursor-pointer rounded-xl bg-white/10 text-white ring-1 ring-white/15 transition hover:bg-white/25 hover:ring-white/40 disabled:opacity-40"
      @click="signInWithPasskey"
    >
      Mit Gerät anmelden
    </UButton>
    <p v-if="reason" class="mt-2 text-center text-xs text-white/40">{{ reason }}</p>
  </div>
</template>

<script setup lang="ts">
// The sign-in form itself. It lives apart from the gate that decides when to
// show it: app.vue answers "may this person in", this answers "let them in".
//
// Mounting with the dialog rather than with the app also puts the passkey offer
// below on the right lifecycle — it is made each time the form appears, not once
// when the app starts.
const { loginajax, loadSession } = useAuth()
const passkey = usePasskey()

const username = ref('')
const password = ref('')
const busy = ref(false)
const passkeyBusy = ref(false)
const error = ref('')
const reason = ref('')

// The fields sit on the dark panel rather than on a card of their own: a filled
// translucent box, no border until focus.
const field = 'w-full rounded-xl bg-white/10 ring-1 ring-white/15 focus-within:ring-white/40 transition'
const inputUi = {
  root: 'w-full',
  base: 'w-full bg-transparent text-white placeholder:text-white/40',
  leadingIcon: 'text-white/40',
}

async function submit() {
  busy.value = true
  error.value = ''
  try {
    if (!await loginajax(username.value, password.value)) {
      error.value = 'Benutzername oder Passwort stimmt nicht.'
      password.value = ''
    }
  }
  finally {
    busy.value = false
  }
}

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
  reason.value = passkey.unavailable()
  if (reason.value) return

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
    if (!passkey.aborted(e)) error.value = passkey.describe(e)
  }
}

async function signInWithPasskey() {
  cancelConditional()
  passkeyBusy.value = true
  error.value = ''
  try {
    await passkey.signIn()
    await loadSession()
  }
  catch (e) {
    error.value = passkey.describe(e)
  }
  finally {
    passkeyBusy.value = false
  }
}
</script>
