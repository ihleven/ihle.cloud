<template>
  <footer class="bg-accented py-8">
    <div class="mx-auto max-w-7xl px-6 pt-8 pb-8 sm:pt-12 lg:px-4 lg:pt-16">
      <!-- Left (brand + link columns) and right (status) side by side on wide
           screens; on narrower ones the right block stacks below. -->
      <div class="flex flex-col gap-12 xl:flex-row xl:items-start xl:justify-between xl:gap-8">

        <!-- Left: the mark, with the link columns beneath -->
        <div class="space-y-8">
          <div class="flex items-center gap-x-6">
            <Wordmark />
          </div>

          <nav class="grid grid-cols-2 gap-8 sm:grid-cols-3">
            <div v-for="col in linkColumns" :key="col.title">
              <h3 class="text-sm/6 font-semibold text-highlighted">
                {{ col.title }}
              </h3>
              <ul role="list" class="mt-2 space-y-2">
                <li v-for="item in col.items" :key="item.label">
                  <span
                    v-if="item.mocked"
                    class="flex items-center gap-2 text-sm/6 whitespace-nowrap text-dimmed"
                    :title="`${item.label} gibt es noch nicht`"
                  >
                    <UIcon :name="item.icon" class="size-4 shrink-0" />{{ item.label }}
                  </span>
                  <NuxtLink
                    v-else
                    :to="item.to"
                    class="flex items-center gap-2 text-sm/6 whitespace-nowrap text-toned hover:text-highlighted"
                  >
                    <UIcon :name="item.icon" class="size-4 shrink-0" />{{ item.label }}
                  </NuxtLink>
                </li>
              </ul>
            </div>
          </nav>
        </div>

        <!-- Right: where the CMS footer puts its metrics. Nothing serves them
             here yet, so the shape is present and the values are not invented —
             see the note in the script. -->
        <!-- Wraps, and only refuses to shrink once it is beside the links
             rather than beneath them. A fixed 240px column that cannot shrink
             is wider than a phone has to spare, and pushes the whole page
             sideways. -->
        <div class="flex flex-wrap gap-8 xl:shrink-0 xl:flex-nowrap">
          <div class="shrink-0">
            <h3 class="text-sm/6 font-semibold text-highlighted">
              Sitzung
            </h3>
            <ul role="list" class="mt-2 space-y-2">
              <li class="text-sm/6 whitespace-nowrap text-toned">
                {{ session?.name || '—' }}
              </li>
              <li class="text-sm/6 whitespace-nowrap text-toned">
                {{ session ? `${modules.length} Bereiche` : '—' }}
              </li>
              <li class="text-sm/6 whitespace-nowrap text-toned">
                {{ expiresIn }}
              </li>
            </ul>
          </div>
          <div class="w-full sm:w-60">
            <h3 class="text-sm/6 font-semibold text-highlighted">
              System
            </h3>
            <ul role="list" class="mt-2 space-y-2">
              <li v-for="metric in mockedMetrics" :key="metric" class="text-sm/6 whitespace-nowrap text-dimmed">
                {{ metric }}: —
              </li>
            </ul>
          </div>
        </div>
      </div>

      <div class="mt-4 border-t border-inverted/10 pt-8 sm:mt-20 lg:mt-24">
        <p class="text-sm/6 text-toned">
          Copyright © {{ new Date().getFullYear() }} ihleven
        </p>
      </div>
    </div>
  </footer>
</template>

<script setup lang="ts">
// Built on the CMS footer's shape, for one concrete reason beyond looking alike:
// the previous footer put every area into a single horizontal UNavigationMenu,
// so it grew wider with each module until it overran the page. Columns inside a
// max-w container grow downward instead, which is a bound rather than a hope.
//
// Areas come from the same list the bar and the menu read, so the three cannot
// drift; entitlements decide which appear.
const { items } = useAreas()
const { session } = useAuth()
const modules = computed(() => session.value?.modules ?? [])

// Where the CMS shows CPU and memory, this app has nothing to show: it serves
// no /metrics or /info. The block is here so the layout is the layout, and the
// rows are dashes rather than plausible numbers — a footer quietly reporting
// invented figures is worse than one openly reporting none.
const mockedMetrics = ['CPUs', 'GORoutines', 'TotalAlloc', 'GoTotal']

const expiresIn = computed(() => {
  const seconds = session.value?.expires_in
  if (!seconds) return '—'
  const hours = Math.floor(seconds / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  return hours > 0 ? `noch ${hours} h ${minutes} min` : `noch ${minutes} min`
})

// Navigation mirrors what is reachable without an entitlement; Bereiche is the
// entitled areas; Konto is this account's own settings. Items marked `mocked`
// render as plain text: the page behind them does not exist yet, and a link
// that goes nowhere is worse than a label that says so.
// Geheimtipp is not in the shared areas list — the front page keeps its own copy
// — so it is named here, and only for an account entitled to it. Linking it
// unconditionally would offer a door that bounces whoever opens it.
const { may } = useModules()

const linkColumns = computed(() => [
  {
    title: 'Navigation',
    items: [
      { label: 'Start', icon: 'i-lucide-home', to: '/' },
      ...(may('geheimtipp') ? [{ label: 'Geheimtipp', icon: 'i-lucide-trophy', to: '/geheimtipp' }] : []),
    ],
  },
  { title: 'Bereiche', items: items.value },
  {
    title: 'Konto',
    items: [
      { label: 'Passkeys', icon: 'i-lucide-key-round', to: '/passkeys' },
      // The endpoint exists; the page does not. Shown as a label rather than a
      // link, so the column reads right without promising a route.
      { label: 'Passwort ändern', icon: 'i-lucide-lock', to: '', mocked: true },
    ],
  },
])
</script>
