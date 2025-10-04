<template>
  <main class="flex h-screen w-screen items-center justify-center bg-gray-100">
    <form
      id="loginform"
      method="POST"
      :action="'http://localhost:8000/tokenauth?redirect=' + redirect"
    >
      <UFormField label="" :error="error">
        <UInput
          name="password"
          type="password"
          color="neutral"

          variant="subtle"
          :highlight="false"
          placeholder="Passwort"
          :ui="{ base: 'rounded' }"
        />
        <!-- <UButton type="submit" color="neutral" :ui="{ base: 'rounded' }">Login</UButton> -->
      </UFormField>
    </form>
  </main>
</template>

<script setup>
const { query } = useRoute()
const url = useRequestURL()
const redirect = url.origin + query.redirect
console.log(query, redirect, url.origin)
const error = ref('')
const { auth } = useAuth()
</script>
