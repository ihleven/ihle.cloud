<template>
  <UApp :toaster="toaster">

    <NuxtLayout v-if="session">
      <NuxtPage />
    </NuxtLayout>
    <div v-else class="absolute inset-0 bg-[url(/Summer-Leaves.jpg)]" />

    <UModal
      :open="!session"
      title="Login" description="Login modal"
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

      </template>
    </UModal>

  </UApp>
</template>

<script setup lang="ts">
import * as z from 'zod'

import type { FormSubmitEvent, AuthFormField } from '@nuxt/ui'

const { toaster } = useAppConfig()
const { session, loadSession, loginajax } = useAuth()
// await loadSession()

const fields: AuthFormField[] = [{
  name: 'username',
  type: 'text',
  label: 'Username',
  placeholder: 'Enter your username',
  required: true,
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
  console.log('Submitted', payload.data.username)
  loginajax(payload.data.username, payload.data.password)
}
</script>
