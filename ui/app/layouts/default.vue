<template>
  <div layout="default">

    <UHeader mode="slideover" :ui="{ root: 'static' }">
      <template #title>
        <h1 class="font-inter text-base font-black text-muted not-last:text-base">
          ihle<span class="font-thin text-dimmed">ven</span><UIcon name="i-lucide-cloud" class="inline-block h-4 w-4 align-middle text-dimmed" />
        </h1>
      </template>

      <UNavigationMenu :items="items" />

      <template #right>
        <!-- <UColorModeButton /> -->

        <AuthPopover />
      </template>

      <template #body>
        <UNavigationMenu :items="items" orientation="vertical" class="-mx-2.5" />
      </template>
    </UHeader>

    <UMain class="z-0 flex h-full bg-ral-7004">
      <slot />
    </UMain>

    <UFooter :ui="{ root: 'bg-ral-7032' }">
      <template #left>
        <p class="text-sm text-muted">
          Copyright © {{ new Date().getFullYear() }}
        </p>
      </template>

      <UNavigationMenu :items="items" variant="link" />

      <template #right>
        <UButton
          icon="i-simple-icons-discord"
          color="neutral"
          variant="ghost"
          to="https://go.nuxt.com/discord"
          target="_blank"
          aria-label="Discord"
        />
        <UButton
          icon="i-simple-icons-x"
          color="neutral"
          variant="ghost"
          to="https://go.nuxt.com/x"
          target="_blank"
          aria-label="X"
        />
        <UButton
          icon="i-simple-icons-github"
          color="neutral"
          variant="ghost"
          to="https://github.com/nuxt/nuxt"
          target="_blank"
          aria-label="GitHub"
        />
      </template>
    </UFooter>

  </div>
</template>

<script setup lang="ts">
import type { NavigationMenuItem } from '@nuxt/ui'

const { path } = useRoute()

const items: NavigationMenuItem[] = [{
  icon: 'i-lucide-home',
  label: 'Content',
  to: '/entries',
  active: path.startsWith('/entries'),
}, {
  label: 'Familie',
  to: '/famihlie',
  active: path.startsWith('/famihlie'),
}, {
  label: 'Search',
  to: '/search',
  active: path.startsWith('/search'),
}]
</script>
