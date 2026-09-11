<template>
  <div class="flex min-h-screen flex-col items-center justify-center gap-6 bg-default p-8">
    <UError :error="error" />

    <!-- The coded error the server sends as a JSON body: it names the cause, and
         is often the only thing distinguishing two otherwise identical 500s.
         Development only — in production it can carry internal paths and query
         detail a visitor has no business seeing. -->
    <pre
      v-if="detail"
      class="max-w-3xl overflow-x-auto rounded bg-elevated p-4 text-xs text-muted"
    >{{ detail }}</pre>
  </div>
</template>

<script setup lang="ts">
import type { NuxtError } from '#app'

// Nuxt renders this outside the layouts, passing the error that ended the
// render. UError draws the status code, message and the button that calls
// clearError; everything below it is ours.
const props = defineProps<{ error: NuxtError }>()

const detail = computed(() => {
  if (!import.meta.dev) return ''

  // data is whatever createError was given — a coded error from the API arrives
  // as a JSON string, so it is unwrapped when it is one and shown as it came
  // when it is not.
  const { data, stack } = props.error
  let body = typeof data === 'string' ? data : data ? JSON.stringify(data, null, 2) : ''
  try {
    body = JSON.stringify(JSON.parse(body), null, 2)
  }
  catch { /* not JSON */ }

  return [body, stack].filter(Boolean).join('\n\n')
})
</script>
