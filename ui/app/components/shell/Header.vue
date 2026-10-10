<template>
  <!-- The top of the page: one element, two behaviours, and the width decides
       which.

       Below lg it is fixed, and on an installed iPhone it is the status bar.
       iOS takes the bar's colour from the pixels the page paints beneath it —
       not from the theme-color meta, which it ignores, and not from the
       manifest, which reaches the splash screen and nothing after it. So this
       decides what the top of the screen looks like, and it is as tall as the
       inset plus a row, or the bar samples the page instead of it.

       It is also what the haze falls on. iOS puts a system material over the
       top of scrolling content and nothing a page can write turns it off — six
       attempts established that. What a page can do is give it one flat colour
       to blur, which is the inset's: nothing is put there, because it is hazed
       and on a phone it is out of reach as well.

       From lg it is the same thing in a different colour, and it stays put for
       the same reason it does on a phone: it used to leave with the page, for
       the iOS reason above — which is a phone matter and does not reach this
       width — and since the areas moved into the menu the button in the corner
       is the whole of the navigation, which should not be able to scroll away.

       Fixed at both widths rather than sticky at one, so this can be written
       last in the layout. The app sets no z-index and stacks by document order,
       and a page's own positioned content — the album tiles are `relative` —
       beats anything written before it however it is positioned. Sticky has to
       sit in the flow at the top to work, which is exactly where it loses. What
       the page leaves clear for it is a padding on the column, from the same
       --top-height. -->
  <!-- The band under the status bar, which is a separate thing from the row.
       The row may leave with the page and this may not: the system
       draws its clock and battery over the top and takes their colour from what
       is painted there, so something has to be.

       In a browser tab --status-bar is zero and this is a strip of no
       height. -->
  <div
    class="fixed inset-x-0 top-0 h-(--status-bar) bg-(--shell-top) lg:bg-default"
    aria-hidden="true"
  />

  <header
    class="inset-x-0 top-(--status-bar) bg-(--shell-top) lg:bg-default"
    :class="shell.top.scrolls ? 'absolute' : 'fixed'"
  >
    <!-- From lg the row is taller than it needs to be, and that is the point:
         the extra height is whitespace above the name, and it carries the line
         down with it. The row's contents then sit at its foot rather than in its
         middle, which puts the extra height above them instead of splitting it
         either side — what is left under them is the padding and the type's own
         half-leading, a gap the eye reads as the line belonging to the row
         rather than running beneath it at a distance.

         The row, and the line under it, stop at --shell-width. The colour runs
         to the edges of the window — the strip is still the strip — but what is
         in it does not, so on a wide screen the name and the round button are
         not a foot apart in opposite corners.

         Below lg there is no room for any of that: one row, its contents
         centred, as tall as a thing worth pressing. -->
    <div
      class="relative mx-auto flex h-14 max-w-(--shell-width) items-center gap-3 px-2 sm:px-6 lg:h-(--bar-height) lg:items-end lg:pb-1"
    >
      <!-- The line under the row, drawn rather than bordered, because it does
           not run the width of anything: it begins where the name begins and
           stops short of the round button, and a border is one line across a
           whole element.

           Where it stops is the button's own room: its width, 1rem from the
           edge, and half a rem of gap. That used to be written out here as
           right-20 with a comment pointing at the menu — a dependency nothing
           could check, and wrong the first time the button changed size. Both
           now come from the same table.

           From lg only. Below it the header is a block of flat colour and a
           line across the bottom of it would be a second edge where the page
           already provides one. -->
      <div
        class="pointer-events-none absolute bottom-0 left-6 h-0.5 bg-inverted max-lg:hidden"
        :class="metric.rule"
        aria-hidden="true"
      />

      <!-- The way back or the way home, at the left. -->
      <div class="flex min-w-0 items-center gap-2">
        <!-- The way back, where there is one. Labelled where there is room for
             words; a plain chevron where there is not — the round, bordered,
             shadowed button existed to be legible over a photograph, and over
             flat colour it is a button wearing a coat indoors. -->
        <UButton
          v-if="back"
          :to="back.to"
          :aria-label="`Zurück zu ${back.label}`"
          :title="back.label"
          icon="i-lucide-chevron-left"
          size="lg"
          variant="ghost"
          color="neutral"
          class="shrink-0 text-(--shell-ink) hover:text-(--shell-ink) lg:hidden"
        />
        <UButton
          v-if="back"
          :to="back.to"
          :label="back.label"
          variant="ghost"
          color="neutral"
          icon="i-lucide-chevron-left"
          class="shrink-0 cursor-pointer max-lg:hidden"
        />

        <!-- The way home, where there is no way back — which is the front page
             and the areas. Only the cloud, not the whole name: the name is the
             size of a heading and this corner holds a control, and the cloud is
             the last glyph of the word rather than an ornament beside it, so it
             stands for the word on its own. -->
        <NuxtLink
          v-if="!back"
          to="/"
          aria-label="Startseite"
          class="shrink-0 text-(--shell-ink) lg:text-default"
        >
          <UIcon name="i-lucide-cloud" class="block size-6" />
        </NuxtLink>
      </div>

      <!-- What the page says it is. A teleport target rather than a prop,
           because a page is not this component's parent and what it has to give
           is markup — a title that knows which episode is playing, a count
           after a name.

           Where no page has claimed it, the area's own name, so the middle of
           the row is never empty. PageTitle says which of the two it is.

           In the middle of the window rather than in the middle of what the
           sides leave, which means taking it out of the row's flow: as a third
           flex item it was centred between two boxes of different widths, and
           a long title then ran under the buttons on the right rather than
           shortening. The gutters are what the sides may grow to — a way back
           on one, a set of buttons and the round button's hole on the other —
           and the title gives up room before either of them does. -->
      <div class="pointer-events-none absolute inset-x-0 flex justify-center px-28 text-sm font-semibold text-(--shell-ink) lg:px-40 lg:text-highlighted">
        <div v-if="!claimed" class="truncate">{{ area }}</div>
        <div id="header-title" class="truncate" />
      </div>

      <!-- What the page offers, and the room the round button needs. -->
      <div class="ml-auto flex min-w-0 items-center gap-1 text-(--shell-ink) lg:text-default">
        <div
          id="header-actions"
          class="flex shrink-0 items-center gap-1"
        />

        <!-- The corner is the round menu button's, and that button is not in
             this row — what it opens is teleported to the end of the document
             and would paint over a button sitting in the flow here, leaving
             nothing to press to close it. So it is positioned against the page,
             in Menu, and the row leaves a hole its width.

             No button, no hole: signed out there is nowhere the menu could take
             anyone, so neither is drawn. -->
        <div v-if="session" class="shrink-0" :class="metric.hole" aria-hidden="true" />
      </div>
    </div>
  </header>

  <!-- The button that opens the menu, and the menu. Written here because the
       button belongs to this row; what it opens teleports itself to the end of
       the document, so neither has to be the last thing in any file.

       Only for somebody signed in. The strip itself is drawn either way — on a
       phone it is the status bar, and a page reached without a session has one
       like any other — but what the menu offers is areas, an account and a way
       out of them, and none of that exists yet. -->
  <Menu v-if="session" />
</template>

<script setup lang="ts">
// The top of the page, at both widths. One element rather than two, because
// what is in it is the same and only its behaviour differs — fixed under the
// status bar below lg, sticky in the flow above it. Two of them would also mean
// two teleport targets with one id between them.

// Whether this page keeps the header at the top of the window or lets it leave
// with the page. Absolute rather than fixed is the whole of the difference: out
// of the flow either way, so the room reserved for it does not change, but one
// of them scrolls away and the other does not.
const { shell } = useShell()

// The row leaves a hole for the round button and stops its rule short of it, so
// it needs the button's size as much as the button does.
const { metric } = useMenuSize()

// Whether there is a menu to offer. The strip is drawn either way; its contents
// are what a session decides.
const { session } = useAuth()

// Where this page goes back to, if anywhere.
const { back } = useBackLink()

// What to put in the middle when the page has not said: the area this path is
// in. A page's own title is better — it can name the episode — but an area's
// name is better than nothing, and nothing is what the front page and the area
// roots had.
const route = useRoute()
const claimed = useState('page-title-claimed', () => false)
const area = computed(() => moduleAt(route.path)?.label ?? '')
</script>
