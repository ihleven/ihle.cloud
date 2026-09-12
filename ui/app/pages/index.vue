<template>
  <main class="grid w-screen grid-cols-[repeat(auto-fit,_minmax(200px,_1fr))] content-start">

    <section class="flex items-center justify-between bg-ral-7035">
      <Logo />
      <NuxtLink to="/login" class="border border-transparent p-2">
        <Icon name="settings" />
      </NuxtLink>
    </section>

    <!-- Only the areas this account is entitled to. Someone who may see one
         thing gets one tile, not a wall of doors that refuse to open. -->
    <NuxtLink
      v-for="tile in tiles"
      :key="tile.module"
      :to="tile.to"
      class="block aspect-square p-8"
      :class="tile.class"
    >
      <h1 class="text-lg font-black text-white hover:text-outline">{{ tile.label }}</h1>
    </NuxtLink>

    <p v-if="!tiles.length" class="col-span-full p-8 text-muted">
      <template v-if="session">Für dieses Konto ist noch nichts freigeschaltet.</template>
      <template v-else><NuxtLink to="/login" class="underline">Anmelden</NuxtLink></template>
    </p>

  </main>
</template>

<script setup lang="ts">
const { session } = useAuth()
const { allowed } = useModules()

// Each tile names the area it belongs to; the entitlement decides whether it is
// rendered at all.
const tiles = computed(() => allowed([
  { module: 'content', label: 'CMS', to: '/entries', class: 'bg-blue-500/80' },
  { module: 'filme', label: 'Super 8', to: '/filme', class: 'bg-rose-500/80' },
  { module: 'familie', label: 'Familie', to: '/famihlie', class: 'bg-violet-500/80' },
  { module: 'kalender', label: 'Kalender', to: '/kalender', class: 'bg-green-500/90' },
  { module: 'mediathek', label: 'Mediathek', to: '/mediathek', class: 'bg-sky-500/80' },
  { module: 'musik', label: 'Musik', to: '/musik', class: 'bg-cyan-300/80' },
]))
</script>
