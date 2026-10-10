<template>
  <UApp :toaster="toaster">

    <NuxtLayout v-if="session || isPublic">
      <NuxtPage />
    </NuxtLayout>

    <div v-else class="absolute inset-0 bg-[url(/Summer-Leaves.jpg)]" />

    <!-- The one thing that cannot be drawn by a layout: there are several
         layouts and a layer brings its own, so what one draws is destroyed on
         the way to another — which stops the music. Everything else the shell
         used to put here is in the layout now, where being rebuilt costs
         nothing.

         Not wrapped. It was, in a box that measured how much of the foot was
         covered — and a fixed child adds no height to its parent, so the box
         measured nothing and the page was never told anything. What the box did
         do was carry pointer-events-none, which the player inherited: its
         buttons could not be pressed, and a click went through to whatever was
         behind it. On the music page that was an album, so pressing pause
         opened a track list.

         Only for somebody signed in: there is nothing to play until then, and
         the login modal is the whole of the page. -->
    <Player v-if="session" />

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
const isPublic = computed(() => route.meta.unguarded === true)

/**
 * The page's colour, as a literal, for whoever reads `theme-color`.
 *
 * Not iOS: installed, it paints the status bar from the pixels under the bar
 * and ignores this. It is a browser's own chrome that reads it, and the splash
 * screen reads the manifest instead.
 *
 * Resolved rather than passed through, because a page says its background as a
 * custom property — `var(--color-ral-7004)` — and a meta tag takes a colour.
 */
const { shell } = useShell()
const painted = ref('#ffffff')

function resolve(colour: string) {
  if (!import.meta.client) return
  const probe = document.createElement('div')
  probe.style.cssText = `position:fixed;visibility:hidden;background:${colour}`
  document.body.append(probe)
  painted.value = getComputedStyle(probe).backgroundColor || '#ffffff'
  probe.remove()
}

watch(() => shell.value.background, c => resolve(String(c)), { immediate: true })

// What the page says the top of the screen is, as custom properties on the
// document: the header reads them, and so does the rule in main.css that paints
// html and body. One source, because the two surfaces showing different colours
// is a status bar that changes as you scroll.
//
// Here rather than in the header, because it is the document that carries them
// and because a page that draws no header still has a body behind it.
const inks: Record<string, string> = { light: 'white', dark: 'black' }

// How much of the top the system has taken for its own clock and battery.
//
// Only where the app is installed; in a browser tab the page is inside the
// window's chrome rather than under it. Asked here rather than in a CSS
// display-mode query, because iPadOS does not reliably answer one and
// useStandalone checks navigator.standalone as well.
//
// What the system has taken at the top for its own clock and battery, and what
// the header's row and the round button are therefore placed under.
//
// The inset, and nothing added to it. A notched iPhone reports ~47px and that
// is already past its haze — the bar is clean there and has been. An iPad
// without a notch reports ~31px and its haze (on iOS 27) reaches further, so
// the top of the round button is tinted there; a floor deep enough for that
// puts 13px of empty black above the button on the phone, which trades a
// working device for a broken one. No single number serves both, so this is the
// one that is true: what the system took.
//
// Only where the app is installed; in a browser tab the page is inside the
// window's chrome rather than under it. Asked here rather than in a CSS
// display-mode query, because iPadOS does not reliably answer one and
// useStandalone checks navigator.standalone as well.
const { standalone } = useStandalone()
const statusBar = computed(() => (standalone.value
  ? 'env(safe-area-inset-top)'
  : '0px'))

useHead({
  meta: [{ name: 'theme-color', content: painted }],
  htmlAttrs: {
    style: computed(() => {
      const { colour, ink } = shell.value.top

      return `--shell-top:${colour};--shell-ink:${inks[ink] ?? 'white'};--status-bar:${statusBar.value}`
    }),
  },
})
</script>
