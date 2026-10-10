<template>
  <!-- Teleported to the end of whatever box the app is, which is how this
       app's rule about stacking is kept rather than worked around: no z-index
       anywhere, and what is later paints over what is earlier (see main.css).
       It is also how Nuxt UI's own overlays manage without a number. Written inside the header,
       where the button belongs; it ends up last in the body either way, and a
       dialog opened afterwards still comes out above it.

       Four layers, and their order here is the whole of the stacking: the scrim
       over the page, the disc over the scrim, the menu over the disc, the button
       over everything — including the header it was written in. -->
  <Teleport to="body">

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
      class="fixed inset-0 flex flex-col justify-start overflow-y-auto transition-opacity duration-200 motion-reduce:transition-none"
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
        class="relative flex w-full max-w-[min(42rem,70vh)] gap-6 self-end p-6 max-lg:flex-col lg:items-start lg:justify-between"
        :style="clearance"
        @click.self="open = false"
      >
        <!-- From lg: two columns, hanging from the top the menu grows out of,
           the areas at one edge and the account at the other.

           Below lg they stack, and the areas go last — which is to say
           nearest. The menu grows out of the bottom corner and is read upward
           from the thumb, so the thing you came to press should be the first
           thing the thumb reaches rather than a column's worth away. Side by
           side on a phone gave each of them about 180px, which is not enough
           for an email address and more than enough for the areas; stacked,
           both get the width.

           The name, level with the button so the two bracket the end of the
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
          class="absolute top-[calc(1rem+var(--status-bar))] left-6 flex items-center"
          :class="metric.name"
          aria-label="Startseite"
          @click="open = false"
        >
          <Wordmark />
        </NuxtLink>

        <!-- Where you can go, against the panel's left edge, under the name.
           Written out rather than taken from the navigation component: a list
           of links is structure, and this one has to sit a particular way.

           Under their concerns, which is where that grouping now shows. It used
           to be in the row along the top, one dropdown per concern, because
           twelve areas do not fit a line — and the panel, which has a column
           and no shortage of room, listed them flat. The row has given the
           areas up; this is the one place they are, so this is where the
           grouping belongs.

           The column is as wide as its longest label and every row fills it —
           which is what the stretch of a flex column gives for free — so the
           icons line up and each row is the same thing to aim at. -->
        <nav class="flex shrink-0 flex-col gap-4">
          <div
            v-for="group in grouped"
            :key="group.concern || group.areas[0]?.id"
            class="flex flex-col gap-1"
          >
            <!-- An area belonging to no concern stands as itself, with no
                 heading over it: a heading of one is a label pretending to be
                 a group. -->
            <h2
              v-if="group.concern"
              class="text-xs font-medium tracking-wide text-dimmed uppercase"
            >
              {{ group.concern }}
            </h2>

            <NuxtLink
              v-for="area in group.areas"
              :key="area.id"
              :to="area.at"
              class="flex items-center gap-2 py-1 text-lg text-muted hover:text-highlighted"
              active-class="text-highlighted font-medium"
            >
              <UIcon :name="area.icon!" class="size-5 shrink-0" />
              {{ area.label }}
            </NuxtLink>
          </div>
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

          <!-- The way to the measuring page, which belongs to no area and is
               therefore in none of the lists. Installed there is no address
               bar, so without a link it cannot be reached at all. -->
          <UButton
            to="/debug/haze"
            variant="ghost"
            color="neutral"
            icon="i-lucide-ruler"
            label="Haze messen"
            class="-mr-2 block"
            @click="open = false"
          />
        </div>
      </div>
    </div>

    <!-- A frame the width of the strip's own row, so the button sits at the end of
       that row rather than in the corner of the window. Below --shell-width the
       two are the same thing and this changes nothing; above it, the button
       comes in with the wordmark instead of drifting a foot away from it.

       It carries the positioning the button used to carry, because a fixed
       button inside a static frame would ignore the frame entirely and go back
       to the window. No height and no pointer events: it is a ruler, and
       everything under it has to stay clickable.

       The frame has no height, so the button hangs from its top edge.

       Fixed, so it is always where it was: a button that scrolls away is a
       button you have to go looking for, and since the areas moved into the
       menu it is the whole of the navigation.

       One corner at both widths, so there is one of everything: one origin for
       the circle, one padding for the panel, one rule for where the button is.

       It leaves with the header where a page sends the header away: a film page
       is one picture and a list, and a white circle sitting on the picture is
       the noise that page is trying not to have. What it opens stays fixed —
       an overlay that scrolled with the page behind it would be no overlay.
       It used to be the bottom corner below lg, within reach of a thumb, and the
       cost of giving that up is stated rather than argued away — the top right
       is the farthest point on a phone from the hand holding it. -->
    <div
      class="pointer-events-none inset-x-0 top-[calc(0.25rem+var(--status-bar))] mx-auto max-w-(--shell-width) lg:top-[calc(1rem+var(--status-bar))]"
      :class="shell.top.scrolls ? 'absolute' : 'fixed'"
    >
      <!-- Three bars and nothing behind them. It used to be a white disc with
           a ring and a shadow, which is what a control needs to be legible over
           a photograph and not what it needs over flat colour — the same
           reasoning the way-back chevron in the header already follows.

           It is also what the haze was showing. The system blurs the top
           `inset + 30px` of the screen and cannot be stopped; a white disc on a
           dark band is the strongest edge up there, so it was the thing the
           blur had to work on, and the top of the button came out washed grey
           on the iPad. Bars take the ink of whatever is behind them and give it
           far less to smear.

           Open, the grown circle is behind it and that is bg-default at every
           width, so the ink is the one that reads on that rather than the one
           the page chose for its band. -->
      <button
        ref="button"
        type="button"
        class="pointer-events-auto absolute top-0 right-1 flex cursor-pointer items-center justify-center rounded-full transition-transform duration-300 ease-out motion-reduce:transition-none lg:right-4"
        :class="[metric.button, open ? 'text-highlighted' : 'text-(--shell-ink) lg:text-highlighted']"
        :aria-expanded="open"
        :aria-label="open ? 'Menü schließen' : 'Menü öffnen'"
        @click="open = !open"
      >
        <!-- Three bars rather than two icons swapped over: the middle one fades
         away and the outer two meet in the middle and cross.

         Each turns about its own centre, not its end. Turning about an end is
         fewer moving parts but the crossing then falls well short of half way
         along the bar, which leaves one arm of the cross longer than the other.
         About the centre the four arms come out equal, at every size.

         One transform apiece and nothing else moves. Drawn here rather than
         taken from an icon set because an icon is a picture and cannot come
         apart.

         aria-hidden: the button already says what it is and what it does. -->
        <span class="relative block" :class="metric.glyph" aria-hidden="true">
          <span
            class="absolute top-1/2 left-0 w-full rounded-full bg-current transition-transform duration-300 ease-out motion-reduce:transition-none"
            :class="metric.bar"
            :style="{ transform: open ? 'translateY(-50%) rotate(45deg)' : `translateY(calc(-50% - ${metric.off}px))` }"
          />
          <span
            class="absolute top-1/2 left-0 w-full -translate-y-1/2 rounded-full bg-current transition-opacity duration-300 ease-out motion-reduce:transition-none"
            :class="[metric.bar, open ? 'opacity-0' : '']"
          />
          <span
            class="absolute top-1/2 left-0 w-full rounded-full bg-current transition-transform duration-300 ease-out motion-reduce:transition-none"
            :class="metric.bar"
            :style="{ transform: open ? 'translateY(-50%) rotate(-45deg)' : `translateY(calc(-50% + ${metric.off}px))` }"
          />
        </span>
      </button>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
// The main menu: a round button in the corner, a circle that grows out of it to
// stand behind the menu, and the menu itself.
//
// The circle and the menu are separate things. The circle is only a background,
// so the menu is laid out against the screen and never cut by it — which is
// what a clip path did, and why the menu had to be arranged around the curve.
const route = useRoute()
const { shell } = useShell()
const { grouped } = useAreas()
const { standalone } = useStandalone()

const open = ref(false)

/** The table lives in the composable: the header depends on these numbers too. */
const { metric } = useMenuSize()

/**
 * The room the panel leaves for the button.
 *
 * One case, because the button is in one corner at both widths. It used to be
 * two, the panel clearing the button at the top on a wide screen and at the
 * foot on a narrow one.
 *
 * In a style rather than in classes because the value is a size's rather than a
 * breakpoint's, and the two have to be combined — `pt-[…]` cannot be written
 * with a number that arrives at runtime.
 */
const clearance = computed(() => ({
  paddingTop: `calc(${metric.value.clear} + var(--status-bar))`,
  paddingBottom: 'calc(2rem + env(safe-area-inset-bottom))',
}))

// Where the circle grows from: the middle of the button, read off the button
// itself rather than rebuilt from the lengths it is positioned with. Those two
// used to be written out separately and had to agree; now there is one of them,
// and it is right wherever the button has ended up.
const centreX = ref('0px')
const centreY = ref('0px')

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

  // Viewport coordinates, because that is what both of them are positioned in,
  // and because a measurement taken before the scrollbar was locked away is a
  // measurement of a slightly narrower window.
  centreX.value = `${b.left + b.width / 2}px`
  centreY.value = `${b.top + b.height / 2}px`
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
//
// Measured again on the way open, so the circle grows from where the button is
// rather than from where it was when the panel last changed size. Cheap, and it
// happens while the circle is still nothing.
//
// After the lock, not before: taking the scrollbar away widens the viewport,
// and the button is anchored to the right edge, so measuring first would put
// the circle's origin a scrollbar's width left of where the button ends up.
// Overlay scrollbars take no space and would hide this; classic ones do not.
watch(open, (isOpen) => {
  document.body.style.overflow = isOpen ? 'hidden' : ''
  if (isOpen) measure()
})

onBeforeUnmount(() => {
  document.body.style.overflow = ''
})
</script>
