<template>
  <article class="flex h-screen w-screen items-center justify-center bg-amber-400">
    <form
      ref="form"
      @submit.prevent="submit"
    >
      <UFormField :error="error">
        <UFieldGroup
          orientation="vertical"
          size="xl"
        >
          <UInput
            id="un"
            name="username"
            type="text"
            color="warning"
            variant="soft"
            :highlight="true"
            placeholder="Username"
            :ui="{ base: 'rounded' }"
          />
          <UInput
            id="pwd"
            name="password"
            type="password"
            color="warning"
            variant="soft"
            :highlight="true"
            placeholder="Passwort"
            :ui="{ base: 'rounded' }"
          />
          <UButton
            type="submit"
            color="warning"

            :ui="{ base: 'rounded' }"
          >
            Login
          </UButton>
        </UFieldGroup>
      </UFormField>
    </form>

    <!-- Usernameless: the authenticator says which credential it holds, so
         there is nothing to type. Shown even where it cannot be used, with the
         reason, because a button that vanishes explains nothing. -->
    <div class="ml-4 flex flex-col items-start gap-1">
      <UButton
        color="warning"
        variant="subtle"
        :loading="busy"
        :disabled="!!reason"
        :ui="{ base: 'rounded' }"
        @click="signInWithPasskey"
      >
        Use a device
      </UButton>
      <p v-if="reason" class="max-w-48 text-xs text-amber-900">{{ reason }}</p>
    </div>
  </article>
</template>

<script setup>
const { loginajax, loadSession } = await useAuth()
const passkey = usePasskey()
const { query } = useRoute()
const error = ref(null)
const form = ref(null)
const busy = ref(false)
const reason = ref('')

// Evaluated after mount: it depends on browser globals, so it cannot be decided
// while the template is first being set up.
onMounted(() => {
  reason.value = passkey.unavailable()
})

async function signInWithPasskey() {
  busy.value = true
  error.value = null
  try {
    await passkey.signIn()
    await loadSession()
    navigateTo(query.redirect || '/')
  }
  catch (e) {
    error.value = passkey.describe(e)
  }
  finally {
    busy.value = false
  }
}
// const actionurl = `http://localhost:8000/auth/login?redirect=${query.redirect}`
async function submit() {
  // try {
  //   console.log('submit', form.value.password.value)
  //   const formData = new FormData(form.value)
  //   console.log(formData)
  //   await $fetch('http://localhost:8000/auth/login', { method: 'POST', body: formData, credentials: 'include' })
  //   navigateTo(query.redirect)
  // }
  // catch (e) {
  //   console.error('=>', e)
  //   error.value = e.message
  // }

  const success = await loginajax(form.value.username.value, form.value.password.value)
  if (success === true) {
    navigateTo(query.redirect)
  }
  else {
    error.value = 'Login failed'
  }
}
</script>
