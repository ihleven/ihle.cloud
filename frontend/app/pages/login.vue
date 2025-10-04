<template>
  <main class="flex h-screen w-screen items-center justify-center bg-gray-100">
    <form
      id="loginform"
      method="POST"
      :action="'http://localhost:8000/auth/signin?redirect=' + redirect"
    >
      <UFormField label="" :error="error">
        <UFieldGroup orientation="vertical" size="xl">
          <UInput
            name="username"
            type="text"
            color="neutral"

            variant="subtle"
            :highlight="false"
            placeholder="Username"
            :ui="{ base: 'rounded' }"
          />
          <UInput
            name="password"
            type="password"
            color="neutral"

            variant="subtle"
            :highlight="false"
            placeholder="Passwort"
            :ui="{ base: 'rounded' }"
          />
          <UButton type="submit" color="neutral" :ui="{ base: 'rounded' }">Login</UButton>
          <!-- <UButton color="neutral" :ui="{ base: 'rounded' }" @click="login">Login</UButton> -->
        </UFieldGroup>
      </UFormField>
    </form>
  </main>
</template>

<script setup>
import { UButton } from '#components'

const { query } = useRoute()
const url = useRequestURL()
const redirect = url.origin + query.redirect
console.log(query, redirect, url.origin)
const error = ref('')
const { auth } = useAuth()

function login() {
  const form = document.getElementById('loginform')
  const formData = new FormData(form)
  $fetch('http://localhost:8000/auth/signin?redirect=' + redirect, { method: 'POST', body: formData })
}
</script>
