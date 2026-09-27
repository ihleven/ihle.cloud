<template>
  <!-- The strip along the top. In the flow by default, so it scrolls away with
       the page: iOS blurs the top of scrolling content and nothing a page can
       write prevents it, so a strip anchored there would be permanently hazed —
       better the haze falls on ordinary content.

       A page that would rather keep it in reach says so, and it is then sticky
       rather than fixed. Sticky stays in the flow, so nothing has to reserve
       room for it and the page does not jump by the strip's height at the
       moment it leaves; it is still a positioned element, so it paints over the
       page without a z-index, which is the rule this app keeps (see main.css).
       The round button in the corner follows the same decision — the layout
       makes it once for the whole band. -->
  <header
    class="relative bg-default"
    :class="pinned ? 'sticky top-0' : ''"
  >
    <!-- Taller than the row needs to be, and that is the point: the extra height
         is whitespace above the wordmark, and it carries the line down with it.
         At 3.375rem the line crosses the round button about two thirds of the
         way down it — the button is 3.5rem tall and hangs 1rem from the top, so
         the line and the button are related to each other rather than each
         sitting where it happened to land.

         The row's contents sit at its foot rather than in its middle, which is
         what puts the extra height above them instead of splitting it either
         side. The name and the areas then keep the line close, and what is left
         under them is the padding and the type's own half-leading — a gap the
         eye reads as the line belonging to the row rather than running beneath
         it at a distance.

         The row, and the line under it, stop at --shell-width. The colour runs
         to the edges of the window — the strip is still the strip — but what is
         in it does not, so on a wide screen the wordmark and the round button
         are not a foot apart in opposite corners.

         The line goes with the row rather than with the strip, which is what
         leaves it out to the left of the wordmark and to the right of the
         button. It can afford to stop there now: with the band below holding
         the strip's own colour for its first part, the line is no longer what
         separates the two, and a rule running to both edges of a wide window
         only underlines emptiness. -->
    <div
      class="relative mx-auto flex h-(--bar-height) max-w-(--shell-width) items-end justify-between gap-3 px-4 pb-1 sm:px-6"
    >
      <!-- The line under the row, drawn rather than bordered, because it does
           not run the width of anything: it begins where the wordmark begins and
           stops short of the round button, and a border is one line across a
           whole element.

           This used to be a border with the bar's own colour painted over it
           where the button crossed — a hole punched in a line that went too far.
           Ending the line is what was wanted all along; it only became possible
           once the line stopped being the separator between the strip and the
           page, which is the band's job now.

           right-20 is the button's own room: 3.5rem across (size-14) and 1rem
           from the edge (right-4), leaving half a rem of gap. Both numbers live
           in AppMenu; if the button moves, this follows. -->
      <div
        class="pointer-events-none absolute right-20 bottom-0 left-4 h-0.5 bg-inverted sm:left-6"
        aria-hidden="true"
      />
      <!-- The name, except on a single film, where the space is better spent on
         the way back to the films — the only place you can have come from. -->
      <UButton
        v-if="film"
        to="/filme"
        variant="ghost"
        color="neutral"
        icon="i-lucide-chevron-left"
        label="Filme"
        class="cursor-pointer"
      />
      <Wordmark v-else />

      <!-- Where there is room for the words. Below that the areas are in the bar
         along the foot of the screen, within reach of a thumb.

         One entry per concern rather than one per area: eleven of them in a row
         is a list to read before you can aim at it, and the strip has a wordmark
         at one end and the round menu button at the other. -->
      <nav class="hidden items-center gap-0.5 lg:flex" aria-label="Bereiche">
        <template v-for="entry in menus" :key="entry.label">
          <!-- A concern, with its areas behind it. The list is a dropdown rather
             than a navigation menu's own panel for one reason: this app sets no
             z-index and stacks by document order (see main.css), and the bar is
             the first thing in the document — so a panel anchored inside it is
             painted under the page. A dropdown's content is portalled to the
             end of the document, which is how the menu and every other overlay
             here already gets above the page without a number. -->
          <UDropdownMenu
            v-if="entry.children"
            :items="[entry.children]"
            :content="{ align: 'start', sideOffset: 8 }"
          >
            <UButton
              :icon="entry.icon"
              :label="entry.label"
              :color="entry.active ? 'primary' : 'neutral'"
              variant="ghost"
              trailing-icon="i-lucide-chevron-down"
            />
          </UDropdownMenu>

          <!-- An area that belongs to no concern stands as itself. -->
          <UButton
            v-else
            :to="entry.to"
            :icon="entry.icon"
            :label="entry.label"
            :color="entry.active ? 'primary' : 'neutral'"
            variant="ghost"
          />
        </template>
      </nav>

      <!-- The corner is the round menu button's, and that button is not in this
         row — it is positioned against the page, in AppMenu. Without a hole its
         size, the areas run underneath it. Its width (size-14); the separation
         is the row's own gap. -->
      <div class="w-14 shrink-0" aria-hidden="true" />
    </div>
  </header>
</template>

<script setup lang="ts">
// The strip along the top, and the areas where there is room to write them out.
//
// There is one navigation, in two renderings, and never both at once: from lg
// it is here, as a concern per dropdown; below that it is the bar along the
// foot of the screen and the menu its "Mehr" opens. Which one you get is the
// width, so neither has to know about the other beyond the breakpoint they
// share.
defineProps<{
  /** Whether the strip stays put while the page scrolls. The layout decides. */
  pinned?: boolean
}>()

const route = useRoute()
const { menus } = useAreas()

// A film's own page, as opposed to the overview at /filme.
const film = computed(() => /^\/filme\/.+/.test(route.path))
</script>
