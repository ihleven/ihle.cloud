<template>
  <UApp :toaster="toaster">

    <NuxtLayout v-if="session || isPublic">
      <NuxtPage />
    </NuxtLayout>

    <div v-else class="absolute inset-0 bg-[url(/Summer-Leaves.jpg)]" />

    <!-- Outside the layout on purpose. There are several layouts and a layer
         brings its own, so a player rendered by one of them would be torn down
         and rebuilt on the way between them — which stops the music. Here
         nothing about navigating can reach it. It draws nothing until something
         is played. -->
    <Player />

    <UModal
      :open="!session && !isPublic"
      title="Login"
      description="Login modal"
      :ui="{
        content: 'w-full max-w-sm rounded-2xl border-0 bg-black/75 p-8 shadow-2xl backdrop-blur-2xl',
        // bg-transparent overrides the app-wide modal theme, which tints the
        // overlay with bg-inverted/75; the picture behind should stay itself.
        overlay: 'bg-transparent backdrop-blur-md',
      }"
    >
      <template #content>
        <LoginForm />
      </template>
    </UModal>

  </UApp>
</template>

<script setup lang="ts">
const { toaster } = useAppConfig()
const { session } = useAuth()

// Signing in is required for everything except the few pages that say otherwise,
// which is what keeps a route from leaking by being added without a guard. The
// exception exists because some pages are reached precisely without a session.
const route = useRoute()
const isPublic = computed(() => route.meta.public === true)
</script>
