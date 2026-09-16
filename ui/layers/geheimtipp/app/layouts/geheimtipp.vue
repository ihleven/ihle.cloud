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

    <!-- The family app's footer, for somebody who has a family app to go back
         to. A pool player sees the page exactly as it was. -->
    <Footer v-if="beyondThePool" />
  </div>
</template>

<script setup lang="ts">
// The pool's own chrome, deliberately not the family app's: nothing in these two
// bars leads out of the pool, and the sign-out in them ends the session for both
// at once, because there is only one.
//
// The one exception is at the very bottom, and only for an account that has
// somewhere else to go — see beyondThePool.
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

// Whether this account is more than a pool player.
//
// A pool account is *confined*: the server hands it exactly one area, the pool
// itself, and refuses it everywhere else. So an account entitled to anything
// besides the pool is one that reached this page from a family app it can go
// back to, and the footer is that way back — a column of areas, which for
// anybody else would list nothing and offer doors that bounce them.
//
// Read from the entitlements rather than from a flag saying "confined", because
// this asks the question the footer actually answers: is there anywhere else
// for this person to go? Someone unconfined but entitled to nothing else is, as
// far as this page is concerned, a pool player.
const { entitled } = useModules()

const beyondThePool = computed(() => entitled.value.some(area => area !== 'geheimtipp'))

const navigation = [
  { label: 'Tipprunde', to: '/geheimtipp' },
  { label: 'Spieltage', to: '/geheimtipp/spieltage' },
  { label: 'Wertung', to: '/geheimtipp/ranking' },
]
</script>
