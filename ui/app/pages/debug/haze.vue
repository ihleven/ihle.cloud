<template>
  <div>
    <!-- The pasted boilerplate's "Seamless Full-Bleed Layout", as written. The
         :root and body rules are in the script below, because they belong to
         the document rather than to any element this page draws.

         Header component spanning the full viewport width. The inset is added
         here as well as on the body, which is the proposal as given: the body
         pushes the page down by the inset and the header pads itself by it
         again.

         The proposal's declarations, written as utilities rather than under an
         .app-header selector — the selector was only ever somewhere to hang
         them. The z-index among them, which the rest of this app does without;
         here it is part of what is being tested. -->
    <header
      class="sticky top-0 w-full bg-[#121212] pt-[env(safe-area-inset-top)]"
      style="z-index: 1000"
    >
      <div class="flex h-14 items-center justify-center text-sm font-semibold">
        app-header
      </div>
    </header>

    <!-- The two additions that are not in the proposal: a ruler and a readout.
         Without them the page can be looked at but not measured, and every
         round so far has turned on a number nobody had.

         The ruler is fixed at the top of the window and numbered every 10px, so
         it sits outside the body's padding and measures from the edge of the
         screen rather than from where the content begins. Alternating stripes
         rather than lines, so a blur shows as two stripes bleeding into each
         other rather than as a line that might just be thin.

         Over the header rather than under it, which is the one place this
         departs from the proposal on purpose: the header is opaque and would
         cover the top 60px of the scale, and a haze over a flat dark header is
         a haze nobody can see. The stripes are the texture that makes it
         visible at all. -->
    <div class="pointer-events-none fixed top-0 left-0 w-20" style="z-index: 1001">
      <div
        v-for="n in 16"
        :key="n"
        class="flex h-[10px] items-center border-b border-black/30 pl-1 text-[8px] leading-none"
        :class="n % 2 ? 'bg-white text-black' : 'bg-black text-white'"
      >
        {{ n * 10 }}
      </div>
    </div>

    <main class="px-4 pb-24">
      <!-- What the device says about itself, printed on the page: installed
           there is no inspector and no address bar, so this is the only way the
           numbers can be read at all. -->
      <dl class="py-4 pl-20 font-mono text-xs">
        <div><dt class="inline">inset-top</dt> <dd class="inline font-bold">{{ insetTop }}px</dd></div>
        <div><dt class="inline">standalone</dt> <dd class="inline font-bold">{{ standalone }}</dd></div>
        <div><dt class="inline">innerHeight</dt> <dd class="inline font-bold">{{ innerHeight }}</dd></div>
        <div><dt class="inline">screen</dt> <dd class="inline font-bold">{{ screenHeight }}</dd></div>
      </dl>

      <p class="pb-4 pl-20 text-sm">
        Oben links die Skala. Wo die Streifen ineinanderlaufen, reicht der
        Schleier. Die Zahl ablesen.
      </p>

      <p
        v-for="n in 60"
        :key="n"
        class="border-b border-white/10 py-3 text-sm"
      >
        {{ n }} — etwas zum Scrollen
      </p>

      <!-- The way back. The page draws none of the shell, so there is no menu
           on it; installed, there is no address bar either, and without this
           the page is a room with no door. -->
      <NuxtLink to="/" class="mt-8 inline-block text-sm underline">
        zurück
      </NuxtLink>
    </main>
  </div>
</template>

<script setup lang="ts">
// A page with none of the shell: no header of ours, no band, no footer, no
// menu. `layout: false` is what does that; `unguarded` keeps the sign-in modal
// off it so it can be opened from anywhere.
//
// It exists to measure the iOS scroll edge effect rather than argue about it.
// See shell-plan, "The haze, and what ends it".
definePageMeta({ layout: false, unguarded: true })

const { standalone } = useStandalone()

// The inset as a number rather than as a length, which is the only way to put
// it on the screen: env() can be used in CSS and read nowhere else. Measured
// off a throwaway element sized by it, because that is what CSS will answer.
const insetTop = ref(0)
const innerHeight = ref(0)
const screenHeight = ref(0)

function measure() {
  const probe = document.createElement('div')
  probe.style.cssText = 'position:fixed;visibility:hidden;height:env(safe-area-inset-top)'
  document.body.append(probe)
  insetTop.value = Math.round(probe.getBoundingClientRect().height)
  probe.remove()

  innerHeight.value = window.innerHeight
  screenHeight.value = window.screen.height
}

onMounted(() => {
  measure()
  window.addEventListener('resize', measure)
})

onBeforeUnmount(() => window.removeEventListener('resize', measure))

// The document half of the boilerplate: the :root colours, the body reset, and
// the body's own inset padding, which is what is meant to push the content
// safely clear of the clock and the battery.
//
// The theme colour matches the background, so the launch transition has nothing
// to flash.
useHead({
  htmlAttrs: { style: 'background-color:#121212;color:#ffffff' },
  bodyAttrs: { style: 'margin:0;padding:0;min-height:100vh;padding-top:env(safe-area-inset-top)' },
  meta: [{ name: 'theme-color', content: '#121212' }],
})
</script>
