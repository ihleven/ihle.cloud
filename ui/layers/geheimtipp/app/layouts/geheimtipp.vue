<template>
  <div class="min-h-screen bg-white text-gray-900">
    <header class="border-b border-gray-200">
      <div class="mx-auto flex max-w-4xl items-baseline gap-4 px-6 py-4">
        <NuxtLink to="/geheimtipp" class="text-lg font-semibold tracking-tight">
          Geheimtipp
        </NuxtLink>
        <span class="text-sm text-gray-500">Die Tipprunde</span>
      </div>
    </header>

    <main class="mx-auto max-w-4xl px-6 py-8">
      <slot />
    </main>

    <footer class="mx-auto max-w-4xl px-6 py-8 text-xs text-gray-400">
      ihleven
    </footer>
  </div>
</template>

<script setup lang="ts">
// The pool's own chrome, deliberately not the family app's.
//
// Someone reading a ranking has no account here and is not being asked for one,
// so the header carries no sign-in and no navigation into the rest of the app.
// When tipping arrives it gets its own sign-in, against the pool's backend.
</script>
