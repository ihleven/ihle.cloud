<template>
  <div layout="default">

    <!-- The menu is normally mobile-only; here it carries signing in and out, so
         it has to be reachable at every width or there is no way to sign out on
         a desktop. -->
    <UHeader
      mode="slideover"
      :ui="{ root: 'static', toggle: 'lg:flex', content: 'lg:block max-w-2xl' }"
    >
      <!-- The name, except on a single film, where the space is better spent on
           the way back to the films — the only place you can have come from.
           The slot is always filled: left empty, UHeader falls back to its own
           Nuxt logo. -->
      <template #title>
        <UButton
          v-if="film"
          to="/filme"
          variant="ghost"
          color="neutral"
          icon="i-lucide-chevron-left"
          label="Filme"
          class="cursor-pointer"
        />
        <Wordmark v-else />
      </template>

      <UNavigationMenu :items="items" />

      <!-- Signing in and out lives in the menu rather than the bar: it is a rare
           action, and the bar is for where you are going. -->
      <template #body>
        <!-- Side by side where there is room. On a phone the panel is the whole
             screen but only ~180px a column, which is not enough for an email
             address, so the two stack instead. -->
        <div class="grid grid-cols-1 divide-y divide-default sm:grid-cols-2 sm:divide-x sm:divide-y-0">
          <UNavigationMenu :items="items" orientation="vertical" class="pb-4 sm:pb-0 sm:pr-4" />
          <AuthPanel class="pt-4 sm:pt-0 sm:pl-4" />
        </div>
      </template>
    </UHeader>

    <UMain class="z-0 flex h-full bg-ral-7004">
      <slot />
    </UMain>

    <UFooter :ui="{ root: 'bg-ral-7032' }">
      <!-- The name lives here instead, so it is on every page rather than one. -->
      <template #left>
        <div class="flex items-baseline gap-3">
          <Wordmark />
          <p class="text-sm text-muted">
            Copyright © {{ new Date().getFullYear() }}
          </p>
        </div>
      </template>

      <UNavigationMenu :items="items" variant="link" />

    </UFooter>

  </div>
</template>

<script setup lang="ts">
import type { NavigationMenuItem } from '@nuxt/ui'

const route = useRoute()
const { allowed } = useModules()

// A film's own page, as opposed to the overview at /filme.
const film = computed(() => /^\/filme\/.+/.test(route.path))

// Each entry names the area it belongs to; the account's entitlements decide
// which are offered. Read through a computed rather than captured once, so the
// active marker follows navigation as well.
const areas = [
  { module: 'content', icon: 'i-lucide-folder-tree', label: 'Content', to: '/entries' },
  { module: 'filme', icon: 'i-lucide-clapperboard', label: 'Filme', to: '/filme' },
  { module: 'familie', icon: 'i-lucide-users', label: 'Familie', to: '/famihlie' },
  { module: 'kalender', icon: 'i-lucide-calendar', label: 'Kalender', to: '/kalender' },
  { module: 'mediathek', icon: 'i-lucide-library', label: 'Mediathek', to: '/mediathek' },
  { module: 'musik', icon: 'i-lucide-music', label: 'Musik', to: '/musik' },
  { module: 'search', icon: 'i-feather-search', label: 'Search', to: '/search' },
]

const items = computed<NavigationMenuItem[]>(() =>
  allowed(areas).map(({ icon, label, to }) => ({
    icon,
    label,
    to,
    active: route.path.startsWith(to),
  })),
)

</script>
