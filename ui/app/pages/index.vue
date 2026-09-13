<template>
  <!-- Signed in: the areas this account has, inside the app's own chrome. -->
  <NuxtLayout v-if="session" name="default">
    <main class="grid w-screen grid-cols-[repeat(auto-fit,_minmax(200px,_1fr))] content-start">
      <section class="flex items-center justify-between bg-ral-7035">
        <Logo />
      </section>

      <!-- Only the areas this account is entitled to. Someone who may see one
           thing gets one tile, not a wall of doors that refuse to open. -->
      <NuxtLink
        v-for="tile in tiles"
        :key="tile.label"
        :to="tile.to"
        class="block aspect-square p-8"
        :class="tile.class"
      >
        <h1 class="text-lg font-black text-white hover:text-outline">{{ tile.label }}</h1>
      </NuxtLink>

      <p v-if="!tiles.length" class="col-span-full p-8 text-muted">
        Für dieses Konto ist noch nichts freigeschaltet.
      </p>
    </main>
  </NuxtLayout>

  <!-- Signed out: the front door, and it is the pool's. -->
  <main v-else class="relative flex min-h-screen w-screen flex-col items-center justify-center bg-zinc-100">
    <NuxtLink
      to="/geheimtipp"
      class="rounded-2xl bg-sky-500/90 px-16 py-12 text-4xl font-black text-white shadow-lg hover:bg-sky-500"
    >
      Geheimtipp
    </NuxtLink>

    <!-- Only for a browser that has signed in here before. The rest of the app
         is not advertised to people who came for the tipping. -->
    <NuxtLink
      v-if="known"
      to="/famihlie"
      class="absolute right-4 bottom-4 text-xs text-gray-400 hover:text-gray-600"
    >
      ihlvn
    </NuxtLink>
  </main>
</template>

<script setup lang="ts">
// Public, and the only page in the app that is.
//
// Two audiences share this domain and only one of them has an account here. A
// sign-in shown to the other is a dead end, so the front page asks for nothing:
// it offers the pool, and — to a browser that has been here before — a quiet way
// back into the rest. Every other route is gated exactly as before.
definePageMeta({ public: true, layout: false })

const { session } = useAuth()
const { allowed } = useModules()

// Set wherever a session begins and left alone by signing out; see
// app/auth/session.go. Not a permission: it only decides whether the way back
// is shown, and everything it leads to still asks for a session.
const known = useCookie<string | null>('ihlvn_known')

// The chrome is chosen by rendering it, not by setPageLayout.
//
// setPageLayout writes the name onto route meta, which for "/" persists for the
// rest of the visit — so signing out left the app's bar on a page that is meant
// to show none. On a first load it defers instead to a beforeResolve hook that
// fires on the *next* navigation, which stamped "default" onto whatever was
// opened from here: the pool's own header and background were replaced by this
// app's. Both are avoided by letting the template decide.

// Each tile names the area it belongs to; the entitlement decides whether it is
// rendered at all.
const tiles = computed(() => allowed([
  { module: 'content', label: 'CMS', to: '/entries', class: 'bg-blue-500/80' },
  { module: 'filme', label: 'Super 8', to: '/filme', class: 'bg-rose-500/80' },
  { module: 'familie', label: 'Familie', to: '/famihlie', class: 'bg-violet-500/80' },
  { module: 'hidrive', label: 'Dateien', to: '/hidrive', class: 'bg-amber-500/80' },
  { module: 'kalender', label: 'Kalender', to: '/kalender', class: 'bg-green-500/90' },
  { module: 'mediathek', label: 'Mediathek', to: '/mediathek', class: 'bg-sky-500/80' },
  { module: 'musik', label: 'Musik', to: '/musik', class: 'bg-cyan-300/80' },
  { module: 'geheimtipp', label: 'Geheimtipp', to: '/geheimtipp', class: 'bg-sky-400/80' },
]))
</script>
