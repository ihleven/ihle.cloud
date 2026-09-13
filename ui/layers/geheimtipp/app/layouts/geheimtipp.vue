<template>
  <div class="flex min-h-screen flex-col bg-zinc-100 text-gray-900">
    <div class="bg-[url('/assets/geheimtipp/rasen.jpg')] bg-cover">
      <header class="mx-auto flex w-full max-w-screen-md items-start justify-between px-4 py-6">
        <NuxtLink to="/geheimtipp" class="text-2xl font-light tracking-tight text-white drop-shadow">
          geheim<span class="font-semibold">tipp</span>
        </NuxtLink>
        <GhtAccountMenu v-if="signedIn" />
      </header>
    </div>

    <nav class="bg-gray-800">
      <div class="mx-auto flex max-w-screen-md items-center justify-between px-2 py-2">
        <div class="flex gap-2">
          <NuxtLink
            v-for="item in navigation" :key="item.to"
            :to="item.to"
            class="rounded-md px-3 py-2 text-sm font-medium text-gray-300 hover:bg-gray-700 hover:text-white"
            active-class="!bg-gray-900 !text-white"
          >
            {{ item.label }}
          </NuxtLink>
        </div>

        <div class="flex items-center gap-3">
          <span v-if="edition" class="text-sm text-gray-400">{{ edition }}</span>
          <GhtTipperMenu v-if="registration" :tipper="registration" />
          <!-- Signed in but not playing this edition: the way in, where the
               registration would otherwise be. -->
          <NuxtLink
            v-else-if="signedIn"
            to="/geheimtipp/register"
            class="rounded-md bg-gray-900 px-3 py-2 text-sm font-medium text-white hover:bg-gray-700"
          >
            Registrieren
          </NuxtLink>
        </div>
      </div>
    </nav>

    <div class="relative grow">
      <slot />
    </div>
  </div>
</template>

<script setup lang="ts">
// The pool's own chrome, deliberately not the family app's: no link back into
// the rest of the site, and a sign-out that ends the pool's session and not the
// family one. They share an origin and nothing else.
//
// Two bars, and which menu belongs to which is the point of having two. The
// header over the grass is the pool itself — the same on every page whatever
// edition is being read — so the account's menu sits there. The bar below is the
// edition: its pages, its name, and the registration this person plays it under.
//
// Nothing here is width-constrained. The tip matrix is as wide as the number of
// tippers makes it, and clamping it in the layout would push the columns the
// page exists to show off the side.
const { edition, registration, signedIn } = useGhtSession()

const navigation = [
  { label: 'Tipprunde', to: '/geheimtipp' },
  { label: 'Spieltage', to: '/geheimtipp/spieltage' },
  { label: 'Wertung', to: '/geheimtipp/ranking' },
]
</script>
