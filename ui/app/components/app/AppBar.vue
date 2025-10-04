<template>
  <header class="flex h-12 w-full items-center justify-start border-y border-muted">
    <!-- <div class="mx-2 flex items-center rounded-full bg-[#ff385f] px-2 shadow-lg shadow-[#ff385f]/50">
      <UIcon
        name="i-lucide-menu"
        class="mx-1 size-6 text-white"
        @click="open = !open"
      />
      <NuxtLink :to="'/'">
        <img
          src="/interhome-bird.svg"
          class="mx-2 h-8 scale-120 cursor-pointer rounded-full shadow shadow-white transition duration-300 ease-in-out hover:translate-x-0 hover:translate-y-0 hover:scale-144"
        >
      </NuxtLink>

      <span class="font-inter text-lg font-extrabold text-white">CMS</span>
    </div> -->

    <!-- Horizontal nav menu -->
    <UNavigationMenu
      :items="navitems"
      class="mx-8 w-fit justify-start"
      arrow
      content-orientation="vertical"
    />

    <!-- Drop down nav menu -->
    <!-- <UPopover :arrow="true">
      <UButton icon="i-lucide-menu" variant="soft" color="neutral" size="xl" />
      <template #content>
        <UNavigationMenu
          orientation="vertical"
          :items="menuitems"
          class="data-[orientation=vertical]:w-48"
        />
      </template>
    </UPopover> -->

    <!-- Search field -->
    <!-- <UInput
      ref="input"
      v-model="params.q"
      icon="i-lucide-search"
      placeholder="Search..."
      class="ml-auto"
      @input="search($event.target.value)"
    >
      <template #trailing>
        <UKbd value="/" />
      </template>
    </UInput> -->

    <UButton
      icon="i-lucide-menu"
      size="xl"
      variant="soft"
      color="neutral"
      class="fixed top-1 left-1 rounded-full"
      @click="open = true"
    />
    <!-- Main menu -->
    <UModal
      v-model:open="open"
      :portal="true"
      :overlay="false"
      :ui="{ content: 'thecontent right-1/2 left-0 top-0 bottom-10 translate-none  rounded-br-full h-2/3 text-right flex-row w-screen-xl max-w-xl justify-start pl-4' }"
    >
      <!-- <UButton
        icon="i-lucide-menu" variant="subtle"
        class="absolute top-2 left-1/2 rounded-full" size="lg"
        @click="open = true"
      /> -->
      <template #content>
        <!-- <UButton icon="i-lucide-x" class="absolute top-2 right-2 rounded-full data-[state=closed]:animate-none data-[state=open]:animate-none" size="xl" @click="open = false" /> -->
        <ul class="mt-32 w-full divide-y divide-muted">
          <li
            v-for="item in menuitems"
            :key="item.label"
            class="ml-auto pl-4"
          >
            <NuxtLink
              :to="item.to"
              class="flex items-center space-x-2 p-2 hover:bg-muted"
            >
              <UIcon
                :name="item.icon"
                class="size-6"
              />
              <span class="text-xl font-semibold">{{ item.label }}</span>
            </NuxtLink>
          </li>
          <li class="ml-auto pl-4">
            <a
              class="flex items-center space-x-2 p-2 hover:bg-muted"
              href="http://localhost:8000/apihle/auth/authorize?state=http://localhost:3000/"
            >
              <UIcon
                name="i-lucide-log-in"
                class="size-6"
              />
              <span class="text-xl font-semibold">Authorize</span></a>
          </li>
        </ul>
      </template>
    </UModal>

    <!-- Main menu close button -->
    <UModal
      v-model:open="open"
      :transition="false"
      :overlay="false"
      :modal="false"
      :ui="{ content: 'fixed bg-transparent top-0 left-0 w-14  shadow-none p-2 ring-0 translate-x-0 translate-y-0 ' }"
    >
      <template #content>
        <UButton
          icon="i-lucide-x"
          size="xl"
          variant="soft"
          color="neutral"
          class="rounded-full"
          @click="open = false"
        />
      </template>
    </UModal>

    <!-- Auth popover -->
    <UPopover
      :arrow="true"
      :content="{
        align: 'center',
        side: 'bottom',
        sideOffset: 0,
      }"
    >
      <UButton
        :icon="auth.sub ? 'i-lucide-user' : 'i-lucide-user-x'"
        color="neutral"
        variant="soft"
        size="xl"
        class="ms-auto me-2"
      />

      <template #content>
        <section class="divide-y divide-accented p-0">
          <UButton
            class="flex w-full rounded-none"
            icon="i-lucide-user"
            color="neutral"
            size="xl"
            variant="none"
          >
            <aside class="text-left text-xs font-semibold">
              <div class="font-light text-muted">
                {{ auth.id }}
              </div>
              <div class="">
                {{ auth.sub }}
              </div>
              <div class="text-dimmed">
                {{ auth.aud }}
              </div>
            </aside>
          </UButton>

          <UButton
            variant="ghost"
            class="w-full rounded-none"
            color="neutral"
            size="xl"
            icon="i-lucide-cloud-cog"
            label="Session details"
            @click="data.open=true"
          />

          <UButton
            variant="ghost"
            class="w-full rounded-none"
            color="neutral"
            icon="i-lucide-user-check"
            label="Login"
            size="xl"
            @click="navigateTo('/login?redirect=' + $route.path)"
          />

          <UButton
            variant="ghost"
            class="w-full rounded-none"
            color="neutral"
            icon="i-lucide-power"
            label="Logout"
            size="xl"
            @click="logout"
          />
        </section>
      </template>
    </UPopover>
  </header>
</template>

<script setup lang="ts">
const open = ref(false)

const { auth, expiry, delta, login, logout, data } = useAuth()

// const input = useTemplateRef('input')

// defineShortcuts({
//   '/': () => {
//     input.value?.inputRef?.focus()
//   },
// })

// const { params } = useSearch()

// function search(q: string) {
//   if (useRouter().currentRoute.path === '/search') {
//     params.value.q = q
//   }
//   else {
//     navigateTo('/search?q=' + q)
//   }
// }

const modules = ref([
  {
    label: 'Travelguide',
    icon: 'i-heroicons-globe-europe-africa',
    to: '/travelguide',
  },
  {
    label: 'Mytool',
    icon: 'i-heroicons-circle-stack',
    to: '/mytool',
  },
])

const navitems = ref([
  {
    label: 'Home',
    icon: 'i-lucide-home',
    to: '/',
  },

  {
    label: 'Content',
    icon: 'i-lucide-folder-tree',
    to: '/entries',
  },
  {
    label: 'Modules',
    icon: 'i-lucide-grip',
    children: modules,
  },
])

const menuitems = ref([
  {
    label: 'Home',
    icon: 'i-lucide-home',
    to: '/',
  },
  {
    label: 'Search',
    icon: 'i-feather-search',
    to: '/search',
  },
  {
    label: 'Content',
    icon: 'i-lucide-folder-tree',
    to: '/entries',
  },
  {
    label: 'Travelguide',
    icon: 'i-heroicons-globe-europe-africa',
    to: '/travelguide',
  },
  {
    label: 'Mytool',
    icon: 'i-heroicons-circle-stack',
    to: '/mytool',
  },

  // {
  //   label: 'Info',
  //   icon: 'i-heroicons-home',
  //   to: '/info',
  // },
  // {
  //   label: 'Admin',
  //   icon: 'i-heroicons-home',
  //   to: '/admin',
  // },
])
</script>
