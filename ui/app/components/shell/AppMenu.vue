<template>
  <!-- Four layers, and their order in this file is the whole of the stacking:
       the scrim over the page, the disc over the scrim, the menu over the disc,
       the button over everything. No z-index is involved — see main.css. -->

  <!-- The page, dimmed, so you can see what you are on top of. Clicking closes. -->
  <div
    class="fixed inset-0 bg-inverted/60 transition-opacity duration-380 motion-reduce:transition-none"
    :class="open ? 'opacity-100' : 'pointer-events-none opacity-0'"
    :inert="!open"
    @click="open = false"
  />

  <!-- The menu's background, and nothing else: a circle grown out of the button.
       It is drawn as a border rather than as a box — an element of no size at
       all, fully rounded, whose border thickness is the radius. Growing a
       border grows the circle outward from the button in every direction, and
       leaves the menu itself free to sit wherever it likes rather than being
       clipped to this shape.

       Its radius is measured rather than written: far enough to reach the
       panel's far corner and no further, so the circle is the size of what it
       has to hold. What is left over is where the curve shows, with the dimmed
       page behind it. -->
  <div
    class="pointer-events-none fixed box-content rounded-full border-[var(--ui-bg)] transition-[border-width] duration-380 ease-[cubic-bezier(0.4,0,0.2,1)] motion-reduce:transition-none"
    :style="{
      left: centreX,
      top: centreY,
      width: '0px',
      height: '0px',
      transform: 'translate(-50%, -50%)',
      borderWidth: open ? radius : '0px',
    }"
  />

  <!-- The menu itself, laid out against the screen rather than against the
       circle. It fades rather than grows: it is already where it belongs, and
       the disc arriving underneath is what makes it readable. -->
  <div
    class="fixed inset-0 flex flex-col overflow-y-auto transition-opacity duration-200 motion-reduce:transition-none"
    :class="open ? 'opacity-100 delay-150' : 'pointer-events-none opacity-0'"
    :inert="!open"
    :aria-hidden="!open"
    @click.self="open = false"
  >
    <!-- Along the top, which is also where the disc reaches: it grows from the
         button in that corner, so the top of the screen is its own background
         and the bottom is where it runs out. The top padding clears the
         button, which the areas would otherwise run under.

         Two ceilings, and the lower one wins. Neither is needed to keep the
         panel on the disc any more — the disc is measured from this panel, so
         it covers whatever the panel turns out to be. Both are about how the
         menu should look.

         42rem stops the areas and the account drifting to opposite edges of a
         large screen with a third of it empty between them.

         70vh keeps the panel narrow on a short screen, which is what leaves
         the circle small enough to still read as a circle: let the panel have
         its full 42rem on a phone held sideways and the disc grown to cover it
         fills the screen, and the curve that the whole shape depends on is
         gone. The cost is that the account details wrap hard at that size.

         On a phone neither binds and the panel is the full width. -->
    <div
      ref="panel"
      class="relative flex w-full max-w-[min(42rem,70vh)] items-start justify-between gap-6 self-end p-6 pt-[calc(5.5rem+env(safe-area-inset-top))] pb-[calc(2rem+env(safe-area-inset-bottom))]"
      @click.self="open = false"
    >
      <!-- The name, level with the button so the two bracket the top of the
           menu. Against the panel's left edge rather than the screen's: on a
           wide screen the disc never reaches that corner, and the name would
           be left on the dimmed page instead of on the menu. On a phone the
           panel is the full width, so the two amount to the same place.

           Positioned out of flow, or it would be a third item in a row laid
           out for two and the account and areas would stop meeting the edges.

           The way home, which the areas do not offer: none of them is the
           front page. Labelled, because what it reads as is the name of the
           place rather than the name of the destination.

           Closed on the way out, rather than left to the watcher on the route:
           opening the menu while already on the front page and pressing this
           would otherwise navigate nowhere and leave the menu standing. -->
      <NuxtLink
        to="/"
        class="absolute top-[calc(1rem+env(safe-area-inset-top))] left-6 flex h-14 items-center"
        aria-label="Startseite"
        @click="open = false"
      >
        <Wordmark />
      </NuxtLink>

      <!-- Where you can go, against the panel's left edge, under the name.
           Written out rather than taken from the navigation component: a list
           of links is structure, and this one has to sit a particular way.

           The column is as wide as its longest label and every row fills it —
           which is what the stretch of a flex column gives for free — so the
           icons line up and each row is the same thing to aim at. -->
      <nav class="flex shrink-0 flex-col gap-1">
        <NuxtLink
          v-for="area in areas"
          :key="area.id"
          :to="area.at"
          class="flex items-center gap-2 py-1 text-lg text-muted hover:text-highlighted"
          active-class="text-highlighted font-medium"
        >
          <UIcon :name="area.icon!" class="size-5 shrink-0" />
          {{ area.label }}
        </NuxtLink>
      </nav>

      <!-- Who you are, at the far end from the areas and under the button that
           opened this. Right-aligned, because it sits against the panel's
           right edge and ragged-right text read against a straight edge looks
           like a mistake rather than a choice. -->
      <div class="min-w-0 shrink text-right">
        <AuthPanel />

        <!-- Installed, there is no address bar to reload from and no pull-to-
             reload gesture. In a browser tab it would sit beside the browser's
             own, so it is not offered there. -->
        <UButton
          v-if="standalone"
          variant="ghost"
          color="neutral"
          icon="i-lucide-rotate-cw"
          label="Neu laden"
          class="mt-4 -mr-2"
          @click="reload"
        />
      </div>
    </div>
  </div>

  <!-- The ring is border-current, so it is the same colour as the bars inside
       it and stays that way: the button reads dark-on-white here and would
       invert with the surface, where a literal black would disappear. -->
  <button
    ref="button"
    type="button"
    class="fixed top-[calc(1rem+env(safe-area-inset-top))] right-4 flex size-14 cursor-pointer items-center justify-center rounded-full border-2 border-current bg-default text-highlighted shadow-xl transition-transform duration-300 ease-out motion-reduce:transition-none"
    :aria-expanded="open"
    :aria-label="open ? 'Menü schließen' : 'Menü öffnen'"
    @click="open = !open"
  >
    <!-- Three bars rather than two icons swapped over: the middle one fades
         away and the outer two meet in the middle and cross.

         Each turns about its own centre, not its end. Turning about an end is
         fewer moving parts but the crossing then falls twelve pixels along a
         twenty-eight pixel bar, which leaves one arm of the cross longer than
         the other. About the centre the four arms come out equal.

         One transform apiece and nothing else moves. Drawn here rather than
         taken from an icon set because an icon is a picture and cannot come
         apart.

         aria-hidden: the button already says what it is and what it does. -->
    <span class="relative block h-5 w-7" aria-hidden="true">
      <span
        class="absolute top-1/2 left-0 h-[3px] w-full rounded-full bg-current transition-transform duration-300 ease-out motion-reduce:transition-none"
        :style="{ transform: open ? 'translateY(-50%) rotate(45deg)' : 'translateY(calc(-50% - 7px))' }"
      />
      <span
        class="absolute top-1/2 left-0 h-[3px] w-full -translate-y-1/2 rounded-full bg-current transition-opacity duration-300 ease-out motion-reduce:transition-none"
        :class="open ? 'opacity-0' : ''"
      />
      <span
        class="absolute top-1/2 left-0 h-[3px] w-full rounded-full bg-current transition-transform duration-300 ease-out motion-reduce:transition-none"
        :style="{ transform: open ? 'translateY(-50%) rotate(-45deg)' : 'translateY(calc(-50% + 7px))' }"
      />
    </span>
  </button>
</template>

<script setup lang="ts">
// The main menu: a round button in the corner, a circle that grows out of it to
// stand behind the menu, and the menu itself.
//
// The circle and the menu are separate things. The circle is only a background,
// so the menu is laid out against the screen and never cut by it — which is
// what a clip path did, and why the menu had to be arranged around the curve.
const route = useRoute()
const { areas } = useAreas()
const { standalone } = useStandalone()

const open = ref(false)

// The middle of the button, which is where the circle grows from. Written from
// the same numbers the button is positioned with, or the two would disagree.
const size = '3.5rem'
const inset = '1rem'

const centreX = `calc(100% - ${inset} - ${size} / 2)`
const centreY = `calc(env(safe-area-inset-top) + ${inset} + ${size} / 2)`

const panel = useTemplateRef<HTMLElement>('panel')
const button = useTemplateRef<HTMLElement>('button')

// How big the circle has to be, in pixels, kept up to date as the things it
// is measured from change. The arithmetic is in discRadius, where it can be
// tested; what is here is only when to ask it.
const radius = ref('0px')

function measure() {
  const p = panel.value?.getBoundingClientRect()
  const b = button.value?.getBoundingClientRect()
  if (!p || !b) return

  radius.value = `${discRadius(p, b)}px`
}

// The page as it is, not the front page: reloading is for getting out of a
// state, and being sent home would lose where you were.
function reload() {
  window.location.reload()
}

// A way somewhere, so it closes on arrival — otherwise it would sit over the
// page it was asked for.
watch(() => route.fullPath, () => {
  open.value = false
})

function onKey(event: KeyboardEvent) {
  if (event.key === 'Escape') open.value = false
}

// The panel is laid out even while the menu is shut — it is only transparent —
// so the radius is known before the circle is asked to grow to it, and the
// first open animates to the right size rather than settling into it.
//
// Two watchers because they answer different things: the observer catches the
// panel changing size, which is the areas arriving after the session loads;
// the resize listener catches the window changing without the panel doing so,
// which is a rotation moving the button under a different safe area.
let observer: ResizeObserver | undefined

onMounted(() => {
  window.addEventListener('keydown', onKey)
  window.addEventListener('resize', measure)
  measure()
  if (panel.value) {
    observer = new ResizeObserver(measure)
    observer.observe(panel.value)
  }
})

onBeforeUnmount(() => {
  window.removeEventListener('keydown', onKey)
  window.removeEventListener('resize', measure)
  observer?.disconnect()
})

// The page behind must not scroll while the menu is over it.
watch(open, (isOpen) => {
  document.body.style.overflow = isOpen ? 'hidden' : ''
})

onBeforeUnmount(() => {
  document.body.style.overflow = ''
})
</script>
