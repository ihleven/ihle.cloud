<template>
  <header class="z-10 w-full border-b border-gray-300 bg-gray-50">
    <main class="container-fluid-lg flex items-center justify-between">
      <!-- <nuxt-link
        to="/"
        class="relative text-3xl font-black tracking-tighter text-white outline outline-0"
      >
        myTool
        <small class="absolute -translate-x-12 translate-y-5 text-lg font-extrabold tracking-tight"> Docs </small>
      </nuxt-link> -->
      <NuxtLink
        to="/"
      >
        <h1
          class="font-inter text-base font-black text-gray-600 not-last:text-base"
        >
          tool-api<span class="font-thin text-gray-400">console</span>
        </h1></NuxtLink>

      <!-- <UButton
        :label="open"
        color="neutral"
        variant="subtle"
        class="w-16"
        @click="open=!open"
      /> -->

      <nav class="mx-auto flex items-center justify-between">
        <NuxtLink
          v-for="n in nav"
          :key="n"
          :to="n.href"
          class="px-4 py-1.5 text-base font-light text-gray-500 hover:bg-black/60 hover:text-gray-100"
          :class="{ 'bg-gray-300': n.name === 'index'? n.name===currentRoute.name : currentRoute.path.startsWith(n.href) }"
        >
          {{ n.label }}
        </NuxtLink>

      <!-- <button
        class="ml-auto p-2"
        @click="emit('logout')"
      >
        Logout
      </button> -->
      </nav>

      <UPopover
        v-if="session"
        mode="hover"
      >
        <UButton
          icon="i-lucide-user-circle"
          size="md"
          color="neutral"
          variant="outline"
        >
          {{ session.id }}
        </UButton>

        <template #content>
          <div class="m-4 inline-flex w-48 flex-col items-end">
            <div class="text-sm font-semibold text-gray-500">
              {{ session.name }}
            </div>
            <small class="text-xs font-light text-gray-500">{{ session.email }}</small>
            <div class="py-4 text-xs text-gray-500">
              Session expiry:
              {{ expiry }}
              <br>
              <small class="text-gray-400">(aktuell nicht identisch mit cookie expiry)</small>
            </div>

            <UButton

              size="md"
              color="neutral"
              variant="solid"
              @click="logout"
            >
              Logout
            </UButton>
          </div>
        </template>
      </UPopover>
      <UButton
        v-else

        size="md"
        color="neutral"
        variant="solid"
        @click="login"
      >
        Login
      </UButton>
    </main>
    <!-- <UCollapsible
      v-model:open="open"
      class="flex flex-col gap-2 w-48"
    >
      <template #content>
        <Menu />
      </template>
    </UCollapsible> -->
  </header>
</template>

<script setup>
const { session, login, logout, expiry } = await useAuth()

const currentRoute = useRoute()
const nav = [
  { name: 'documentation', href: '/docs', label: 'Documentation' },
  { name: 'info', href: '/info', label: 'Info' },
  { name: 'schema', href: '/schema', label: 'Schema' },
  { name: 'forms', href: '/forms', label: 'Forms' },
  { name: 'accommodations', href: '/accommodations', label: 'Accommodations' },

]
// const open = ref(true)

// defineShortcuts({
//   o: () => open.value = !open.value,
// })
</script>
