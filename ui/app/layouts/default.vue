<template>
  <div>
    <!-- The bar and the content, together as tall as the screen: the column is
         a screenful and the bar's height is taken out of it as padding, which
         box-sizing keeps inside the hundred per cent rather than adding to it.
         The bar is fixed and reserves nothing itself, so this is where the room
         for it is made.

         The footer is after the column rather than inside it. Inside, it would
         be part of the screenful and sit at the bottom of the viewport on a
         short page; after it, the column fills the screen on its own and the
         footer begins exactly at the fold — reached by scrolling, not on
         arrival. -->
    <!-- --content-shell-offset is for what is embedded rather than for anything
         drawn here: the CMS's content module frames itself against the window
         and can only subtract what this shell has already spent above it.

         Built from the one number the header publishes rather than from a
         second idea of the same thing. It used to be forced to 0px below lg
         while <main> was padded by the strip's height, and a page framed
         against the window was then that much too tall: 900px of window,
         948px of page.

         The band is the only part that is not in --top-height, because it
         hangs under the wide rendering alone. Nothing below lg, therefore, and
         its own height above. -->
    <div
      class="flex min-h-screen flex-col pt-(--top-height) [--band-spent:0px] [--content-shell-offset:calc(var(--top-height)+var(--band-spent))] lg:[--band-spent:var(--band-height)]"
      :style="{
        '--page-bg': shell.background,
        '--band-to': shell.bandTo ?? shell.background,
        '--band-height': band ?? '0px',
        '--sticky-top': stickyTop,
      }"
    >

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
        v-if="band"
        class="shrink-0 max-lg:hidden"
        :style="{ height: band, background: fade }"
        aria-hidden="true"
      />

      <!-- A plain <main>, beside the bar rather than around it: this is the
           document's dominant content, and the bar is a banner next to it. -->
      <main class="flex grow bg-(--page-bg)">
        <slot />
      </main>
    </div>

    <!-- Not on a narrow screen. It repeats the areas the menu already carries
         and the account the menu already shows, offers a label that is not a
         link, and ends in a table of dashes that reports nothing by design. On
         a page you scroll with a thumb that is a long way to the bottom for
         nothing. A wide screen has room for a page to end somewhere. -->
    <Footer v-if="shell.footer" class="max-lg:hidden" />

    <!-- Last, which is the whole of why it paints over the page: this app sets
         no z-index and stacks by document order, and a page's own positioned
         content — Musik's album tiles are `relative` — beats anything written
         before it. Written at the top of the column it was scrolled over; the
         room it needs is the column's padding, not its place in the flow. -->
    <Header />
  </div>
</template>

<script setup lang="ts">
// The page's own furniture: the strip at the top, the band under it, the
// background, and the footer. Not the navigation — that is the shell, in
// app.vue, where it survives the way from one layout to the next.
//
// What a page may refuse is `shell` meta, read through useShell so that this
// and the shell itself cannot hold two opinions about it.
const { shell } = useShell()

/**
 * What a page's own sticky things have to clear.
 *
 * The header is over the page, so they clear what it covers: all of it, or only
 * the band where the page has sent the row away.
 *
 * One number rather than six pages each choosing, which is how two of them came
 * to stick underneath the header and hide their own first inch.
 */
const stickyTop = computed(() =>
  (shell.value.top.scrolls ? 'var(--status-bar)' : 'var(--top-height)'))

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
</script>
