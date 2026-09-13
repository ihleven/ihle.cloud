<template>
  <div class="flex h-screen w-screen items-center justify-center px-4">
    <div class="w-full max-w-sm space-y-8">
      <h2 class="text-center text-xl font-extrabold text-gray-900">
        <img class="mx-auto h-12 w-auto" src="/assets/geheimtipp/icon.png" alt="Geheimtipp">
        Geheimtipp
        <small class="block text-center font-medium text-gray-600">Anmeldung</small>
      </h2>

      <form novalidate class="space-y-8" @submit.prevent="submit">
        <!-- The two fields read as one control: no gap between them, and only
             the outer corners rounded. Labels are for screen readers; the
             placeholder is what is seen. -->
        <div class="-space-y-px rounded-md shadow-sm">
          <div>
            <label for="username" class="sr-only">Benutzername</label>
            <input
              id="username" v-model="username" name="username"
              type="text"
              autocomplete="username" required autofocus
              placeholder="Benutzername"
              class="relative block w-full appearance-none rounded-none rounded-t-md border border-blue-300 px-3 py-2 text-gray-900 placeholder-gray-500 focus:z-10 focus:border-blue-500 focus:ring-blue-500 focus:outline-none sm:text-sm"
            >
          </div>
          <div>
            <label for="password" class="sr-only">Passwort</label>
            <input
              id="password" v-model="password" name="password"
              type="password"
              autocomplete="current-password" required placeholder="Passwort"
              class="relative block w-full appearance-none rounded-none rounded-b-md border border-blue-300 px-3 py-2 text-gray-900 placeholder-gray-500 focus:z-10 focus:border-blue-500 focus:ring-blue-500 focus:outline-none sm:text-sm"
            >
          </div>
        </div>

        <button
          type="submit"
          :disabled="busy"
          class="flex w-full justify-center rounded-md border border-transparent bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 focus:outline-none disabled:opacity-60"
        >
          Anmelden
        </button>

        <p v-if="error" class="text-sm text-red-600">{{ error }}</p>
        <p v-else-if="failed" class="text-sm text-amber-600">Die Tipprunde ist gerade nicht erreichbar.</p>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
// No layout at all: the pool's header and its edition bar both describe a
// session that does not exist yet, and the page they frame is the one asking for
// it. `public` is still needed — the family app's own sign-in overlay would
// otherwise cover a form for a different account entirely.
definePageMeta({ public: true, layout: false })

const { signIn, failed } = useGhtSession()
const route = useRoute()

const username = ref('')
const password = ref('')
const error = ref('')
const busy = ref(false)

async function submit() {
  busy.value = true
  error.value = ''
  try {
    await signIn(username.value, password.value)
    // Back to whatever the gate interrupted, or the pool's landing page.
    await navigateTo((route.query.weiter as string) || '/geheimtipp')
  }
  catch {
    // The backend says only that the pair was refused, and which half was
    // wrong is not something to guess at out loud.
    error.value = 'Login oder Passwort stimmt nicht.'
  }
  finally {
    busy.value = false
  }
}
</script>
