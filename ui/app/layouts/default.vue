<template>
  <div>
    <!-- The bar and the content, together as tall as the screen: the content
         takes whatever the bar leaves. Nothing subtracts the bar's height from
         the viewport, so nothing has to be told what that height is, and a page
         that hides the bar does not reserve room for one never rendered.

         The footer is deliberately outside this column. Inside it, it would be
         part of the screenful and sit at the bottom of the viewport on a short
         page; outside, the column fills the screen on its own and the footer
         begins exactly at the fold — reached by scrolling, not on arrival. -->
    <!-- --content-shell-offset is for what is embedded rather than for anything
         drawn here: the CMS's content module frames itself against the window
         and can only subtract what this shell has already spent above it. -->
    <div
      class="flex min-h-screen flex-col"
      :style="{
        '--page-bg': shell.background,
        '--band-to': shell.bandTo ?? shell.background,
        '--content-shell-offset': offset,
      }"
    >
      <AppBar v-if="shell.bar" :pinned="pinned" />

      <!-- The join between the strip and the page, where a page asks for one.
           A gradient rather than a gap: a page whose content begins hard against
           the strip — the front page's wall of tiles does — reads as one block
           of colour with the join lost somewhere inside it, and a plain band of
           either colour only moves the line rather than removing it.

           Part of the column and not of the strip, so a pinned strip slides
           over it as the page scrolls, which is what it means for the band to
           belong to the page.

           It arrives at whatever is directly under it, which is the page's own
           background unless the page says otherwise: a page opening with a
           block of its own colour — the calendar's green band — fades to that
           instead, or the line would simply move down by the height of the
           band.

           `--ui-bg` is what Nuxt UI's `bg-default` resolves to — the colour the
           strip is actually painted — so this follows light and dark without
           being told, and without the layout holding a second opinion about
           what the strip looks like.

           aria-hidden and empty: it is a colour, and there is nothing in it to
           read or to land on. -->
      <div
        v-if="shell.bar && band"
        class="shrink-0"
        :style="{ height: band, background: fade }"
        aria-hidden="true"
      />

      <!-- A plain <main>, beside the bar rather than around it: this is the
           document's dominant content, and the bar is a banner next to it. -->
      <main
        class="flex grow bg-(--page-bg)"
        :style="{ paddingBottom: reserved }"
      >
        <slot />
      </main>
    </div>

    <Footer v-if="shell.footer" />

    <!-- Last, so its button is above everything and its panel above everything
         but the button.

         A page that supplies its own way out — the reader does — can turn this
         off, or the round button would sit over it. -->
    <AppMenu
      v-if="shell.menu"
      :pinned="menuPinned"
    />
  </div>
</template>

<script setup lang="ts">
// The default shell: a strip at the top, the footer, and the player over
// whatever is playing. Navigation is the menu, at every width, plus the areas
// written into the top strip where there is room for them.
//
// A page keeps all of it unless it says otherwise. What it can turn off is in
// ShellParts; what it cannot is anything depending on who is signed in, because
// definePageMeta is resolved at build time — that belongs here, or in a layout
// of the module's own, as geheimtipp's footer does.
const route = useRoute()

const shell = computed(() => ({
  bar: true,
  footer: true,
  menu: true,
  // The colour the pages have always been. Named here rather than left to the
  // <main> element's class, because the band has to fade to exactly it and two
  // places saying what the page's colour is would eventually disagree.
  background: 'var(--color-ral-7004)',
  band: '2rem' as string | false,
  // The strip keeps its own colour for the first part of the band. Fading from
  // the very first pixel leaves the strip without an edge — the colour is
  // already going where the eye is still looking for the join.
  bandHold: '40%',
  ...(route.meta.shell ?? {}),
}))

// A height, or nothing. '0' is spelled out as often as false is, and a band of
// no height with a gradient in it is a thing the browser will happily lay out.
const band = computed(() => {
  const asked = shell.value.band

  return !asked || asked === '0' ? undefined : asked
})

// The gradient itself, built rather than written, because how many stops it has
// depends on what the page asked for — and a rule that covers two stops and
// three and a held first colour with vars alone needs a no-op value for every
// one of them, which there isn't an honest one of.
//
// `--ui-bg` is what Nuxt UI's `bg-default` resolves to, so the top of the band
// is the colour the strip is actually painted and follows light and dark
// without being told.
const fade = computed(() => {
  const { bandHold: hold, bandVia: via } = shell.value
  const stops = ['var(--ui-bg) 0']

  if (hold && hold !== '0') stops.push(`var(--ui-bg) ${hold}`)
  if (via) stops.push(via)
  stops.push('var(--band-to) 100%')

  return `linear-gradient(to bottom, ${stops.join(', ')})`
})

// Whether the band at the top stays put while the page scrolls.
//
// One decision for the strip and the round button, because they sit in the same
// band: one of them staying while the other left was the only thing that made
// it look like two decisions. A page can still part them — the reader has no
// strip and its way out leaves with the page — by saying so for the menu alone.
const pinned = computed(() => shell.value.bar === 'fixed')
const menuPinned = computed(() =>
  shell.value.menu === true ? pinned.value : shell.value.menu === 'fixed',
)

// What a page filling the window has to leave for this shell: the strip, and
// the band under it where there is one. Worked out here because it is this
// layout that decides whether either is drawn — a page cannot know, and the one
// that needs the answer is the CMS's content module, which comes from a layer
// and knows nothing about this app's strip at all.
//
// Ordinary pages ignore it. They grow to what is in them and the column handles
// the rest; this is for the one kind of page that is a frame rather than a
// document.
const offset = computed(() => {
  if (!shell.value.bar) return '0px'

  return band.value ? `calc(var(--bar-height) + ${band.value})` : 'var(--bar-height)'
})

// Room at the foot for the player, when something is playing. Reserved here
// rather than by each page, because the player outlives any of them.
const { current } = usePlayer()

const reserved = computed(() => (current.value ? '4rem' : undefined))
</script>
