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
  </article>
</template>

<script setup>
const { loginajax } = await useAuth()
const { query } = useRoute()
const error = ref(null)
const form = ref(null)
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
