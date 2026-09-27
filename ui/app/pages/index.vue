<template>
  <!-- An account that exists only for the pool gets the pool's own front door,
       reproduced from ihleven.de: no chrome from this app around it, because
       none of it would lead anywhere they may go. -->
  <main
    v-if="confined"
    class="ght-door grid min-h-screen place-content-center overflow-hidden font-sans antialiased"
  >
    <!-- The glow is fixed and so painted over unpositioned content whatever
         the order; giving the content a position of its own puts the two on the
         same footing, where document order decides and the later one wins. That
         is why this needs no z-index — see main.css. -->
    <div class="spotlight fixed right-0 left-0" />
    <div class="relative max-w-[520px] text-center">
      <div class="flex w-full flex-col items-center justify-center">
        <NuxtLink
          to="/geheimtipp"
          class="gradient-border text-md cursor-pointer px-4 py-2 sm:px-6 sm:py-3 sm:text-xl"
        >
          zum Geheimtipp
        </NuxtLink>
      </div>
    </div>
  </main>

  <!-- Everyone else: the areas this account has, inside the app's own chrome. -->
  <NuxtLayout v-else name="default">
    <!-- A wall of mixed blocks rather than a table of equal squares. The cell
         is 11rem, which on a 360px phone is two columns — small enough that a
         wide tile is the screen and not more than it — and on a laptop is
         seven or so.

         auto-fill and not auto-fit: auto-fit collapses the columns nobody is
         using, which on a page with three tiles on it stretches those three
         across the whole window and undoes the sizes entirely.

         Deliberately not `dense`. A block that will not fit at the end of a row
         moves to the next and leaves the remainder empty, and that hole is the
         point of the exercise — dense packing closes every one of them and
         gives back the slab.

         w-full, not w-screen: this sits inside the layout's <main>, a flex
         container, and 100vw is the viewport including any scrollbar — wider
         than the space it has been given. The overflow shows as a pale strip
         down the right and a page that scrolls sideways. -->
    <main
      ref="grid"
      class="grid w-full auto-rows-[14rem] grid-cols-[repeat(auto-fill,_minmax(14rem,_1fr))] content-start"
    >
      <section class="col-span-1 row-span-1 flex items-center justify-between bg-ral-7035">
        <Logo />
      </section>

      <!-- Only the areas this account is entitled to. Someone who may see one
           thing gets one tile, not a wall of doors that refuse to open. -->
      <template v-for="{ tile, class: span, key } in laid" :key="key">
        <!-- A hole: the page's own background, showing through. Rendered rather
             than left out, because a gap in a packed grid is not an absence of
             an element — the next tile simply moves up into it. -->
        <div v-if="!tile" :class="span" />

        <!-- The calendar tile is the calendar's own date block rather than a
             label: it says what day it is, and each part of the date is a way
             into the page for that part. Which is why it cannot be one link
             like the others — an anchor inside an anchor is not a thing a
             browser will render. -->
        <div
          v-else-if="tile.module === 'kalender'"
          class="p-4 sm:p-6"
          :class="[tile.class, span]"
        >
          <KalenderHero variant="tile" />
        </div>

        <!-- Three layers, and no z-index between them: each is positioned and
             each comes after the last, which is all CSS needs to stack them —
             see main.css. The picture, a scrim dark enough to carry the label
             over whatever the picture happens to be doing in that corner, and
             the label. -->
        <NuxtLink
          v-else
          :to="tile.to"
          class="tile relative block overflow-hidden p-8"
          :class="[tile.class, span]"
        >
          <!-- Two nested elements for one picture, because two things move it
               and a transform is one property. The frame drifts with the page;
               the picture inside it grows under the pointer. Written on one
               element they would overwrite each other, and the scroll one —
               being an animation rather than a transition — would win. -->
          <span v-if="tile.photo" class="tile-frame pointer-events-none absolute inset-x-0">
            <img
              :src="tile.photo"
              alt=""
              loading="lazy"
              decoding="async"
              class="tile-photo absolute inset-0 size-full object-cover"
            >
          </span>
          <span v-if="tile.photo" class="tile-scrim pointer-events-none absolute inset-0" />

          <!-- On a photograph the outline is worn always, not on hover. A
               picture is not one colour: the scrim holds the corner down but
               the film reels are near-white exactly where the label sits, and
               white on near-white is legible in the way a thing can be legible
               and still not be worth reading. The calendar tile settled this
               same question the same way. A flat tile keeps the outline for
               hover, where it is an answer to the pointer rather than a
               repair. -->
          <h1
            class="tile-label relative text-lg font-black text-white"
            :class="tile.photo ? 'text-outline' : 'hover:text-outline'"
          >{{ tile.label }}</h1>
        </NuxtLink>
      </template>

      <p v-if="!laid.length" class="col-span-full p-8 text-muted">
        Für dieses Konto ist noch nichts freigeschaltet.
      </p>

      <!-- The application's own Go documentation. A page of this app, from the
           CMS's godoc layer, reading JSON the backend renders out of the source
           embedded in the binary.

           A tile like the others in every respect but its colour: it is not an
           area an account is entitled to, so it is not in the list above, but it
           sits in the same grid and should not announce itself as something
           else. -->
      <NuxtLink
        to="/godoc"
        class="col-span-1 row-span-1 block bg-yellow-500/80 p-8"
      >
        <h1 class="tile-label text-lg font-black text-white hover:text-outline">Dokumentation</h1>
      </NuxtLink>
    </main>
  </NuxtLayout>
</template>

<script setup lang="ts">
// The front page, and the sign-in surface.
//
// It used to be public: two audiences share this domain, only one of them had
// an account, and a sign-in shown to the other was a dead end — so the page
// asked for nothing and offered the pool, with a quiet way back for a browser
// that had been here before.
//
// One form now accepts both, so the dead end is gone and with it the reason to
// be public. Signed out, app.vue raises the sign-in dialog over this page like
// any other gated route. Signed in, what someone sees follows from who they
// are: the tiles they are entitled to, or — for an account that exists only for
// the pool — the pool's own door.
//
// The chrome is chosen by rendering it, not by setPageLayout. setPageLayout
// writes the name onto route meta, which for "/" persists for the rest of the
// visit — so it would stamp this app's bar onto the pool's pages opened from
// here. `layout: false` plus an explicit NuxtLayout is what keeps each branch
// to its own.
definePageMeta({ layout: false })

const { session } = useAuth()
const { allowed } = useModules()

// The labels are given a little weight, so scrolling swings them and stopping
// settles them back. On the grid rather than on each label: one loop, one style
// write, and every label reads the result by inheritance.
const grid = ref<HTMLElement | null>(null)
useScrollGravity(grid)

// A confined account is reported as entitled to the pool and nothing else, by
// its type rather than by its permissions; see Service.offered.
const confined = computed(() =>
  session.value?.modules.length === 1 && session.value.modules[0] === 'geheimtipp',
)

// Each tile names the area it belongs to; the entitlement decides whether it is
// rendered at all.
// The tiles this account may see. Which modules have one, what each says and
// how it looks is utils/modules — the same list the navigation and the route
// guard read, so a new area appears in all three or in none.
const tiles = computed(() => allowed(
  tileModules().map(m => ({
    module: m.id,
    label: m.tile!.label,
    to: m.at,
    class: m.tile!.class,
    photo: m.tile!.photo,
    shapes: m.tile!.shapes,
  })),
))

// The arrangement: a shuffled order and a shape apiece, dealt from one number.
//
// Held in Nuxt state rather than worked out on every mount, so it is rolled
// once when the app is opened and stays put while it is being used. Re-rolled
// per mount, coming back from a tile would find the page rearranged — which
// does not read as variety, it reads as having pressed the wrong thing.
const arrangement = useState('tile-mosaic', () => Math.random())
const laid = computed(() => mosaic(tiles.value, arrangement.value))
</script>

<!-- Carried over from ihleven.de's own front page so the door looks the way it
     always has: a blurred gradient behind, and a pill whose border is a gradient
     masked to the edge, sliding on hover. Not scoped — the spotlight is fixed
     and full-bleed, and scoping buys nothing for two classes used here only. -->
<style>
/* The surface follows prefers-color-scheme rather than Tailwind's dark:
   variant, which never fires here — @nuxt/ui's colorMode is off, so no .dark
   class is ever set. Keying the page off one signal and the pill off another
   is how you get a white page holding a pill styled for a dark one. */
.ght-door {
  background-color: #fff;
  color: #000;
}

@media (prefers-color-scheme: dark) {
  .ght-door {
    background-color: #000;
    color: #fff;
  }
}

/* The calendar tile is the whole green ramp, pale at the top and deep at the
   foot, so the square reads as one colour with depth rather than a flat fill.

   Every green main.css defines is in it. Green 500 is not one of them — the
   file runs 50 to 950 and leaves 500 to Tailwind's own palette — so ral-5000
   stands in that slot, which is where it belongs: #00C16A falls exactly between
   green-400 and green-600, and is plainly the 500 the ramp was written around.

   The date sits in the middle of the square and so on the middle of the ramp,
   but the ends run from near-white to near-black, which is why everything on
   this tile is drawn with the outline: it is the one treatment that survives
   both.

   The ramp drifts down the tile and back, a full sweep taking twenty-four
   seconds — slower than the DJ tile, because a calendar is a quieter thing to
   look at and the two sitting in the same grid should not appear to be keeping
   time with each other.

   It travels from 30% rather than from the top, and the gradient box is a
   little over twice the tile rather than three times. Both are the same
   constraint: the palest greens have to stay up in the corner. Swept under the
   date they leave white letters on near-white, held together by half a pixel of
   outline, which is legible in the way a thing can be legible and still not be
   worth reading. */
.tile-kalender {
  background-image: linear-gradient(
    160deg,
    var(--color-green-50) 0%,
    var(--color-green-100) 10%,
    var(--color-green-200) 20%,
    var(--color-green-300) 30%,
    var(--color-green-400) 40%,
    var(--color-ral-5000) 50%,
    var(--color-green-600) 60%,
    var(--color-green-700) 70%,
    var(--color-green-800) 80%,
    var(--color-green-900) 90%,
    var(--color-green-950) 100%
  );
  background-size: 100% 220%;

  /* Stated rather than left at the default, because it is also what shows when
     the animation is turned off below: 0% 0% would park the tile on exactly the
     pale end the sweep is arranged to avoid. */
  background-position: 50% 65%;
  animation: tile-kalender 24s ease-in-out infinite alternate;
}

@keyframes tile-kalender {
  from {
    background-position: 50% 30%;
  }

  to {
    background-position: 50% 100%;
  }
}

@media (prefers-reduced-motion: reduce) {
  .tile-kalender {
    animation: none;
  }
}

/* What the gravity loop spends its number on.
   Nothing but a translate, so it stays on the compositor and costs no layout —
   twelve labels moving four pixels must not reflow a grid. The fallback of 0px
   is what shows before the first frame, and the whole of what shows when the
   loop never starts: reduced motion, or a browser with no JavaScript running
   yet. The label is simply where it was drawn. */
.tile-label {
  transform: translateY(var(--gravity, 0px));
}

/* A tile with a photograph, and the three things that move it.
   Nothing here is meant to be noticed on its own: a grid of squares that all
   breathe at once is a screensaver. Each movement is the smallest one that
   still reads as a response.

   The picture is drawn a little larger than the square and offset upwards, so
   there is something to slide: the frame drifts by 4% of its own height in each
   direction and 12% of spare height is exactly enough that no edge is ever
   pulled into view. */
.tile-frame {
  top: -6%;
  height: 112%;
}

/* Starting a touch over 1 leaves room for the press below to go *down* without
   uncovering a corner, and hides the half-pixel seam a scaled image can show
   against the tile edge. */
.tile-photo {
  transform: scale(1.02);
  transition:
    transform 700ms cubic-bezier(0.2, 0.6, 0.2, 1),
    filter 700ms ease;
}

/* Hover: the picture grows, slowly, and lifts a little out of the scrim.
   700ms in and the same out — a zoom that snaps back is a flinch. */
.tile:hover .tile-photo {
  transform: scale(1.08);
  filter: saturate(1.15) brightness(1.06);
}

/* Click: the whole square presses in and comes back. Short, because a press is
   an answer rather than a transition, and on the tile rather than the picture
   so the label goes down with it — a picture that shrinks under a label that
   does not is a picture coming loose.

   The photo's own transition is overridden to match, or the press would take
   700ms to arrive and the pointer would be gone before it did. */
.tile {
  transition: transform 260ms cubic-bezier(0.2, 0.6, 0.2, 1);
}

.tile:active {
  transform: scale(0.975);
  transition-duration: 110ms;
}

.tile:active .tile-photo {
  transform: scale(1.04);
  transition-duration: 110ms;
}

/* The scrim exists for the label, not for the picture. Darkest in the top-left
   corner where the label sits and gone by the middle, so the photograph is
   mostly itself — a flat wash over the whole square would make four grey
   squares out of four photographs.

   The second stop is the insurance: a hand-picked corner gradient is only right
   until somebody swaps a photograph for a paler one, and a tenth of black
   across the whole square costs nothing to look at and keeps white legible on
   anything. */
.tile-scrim {
  background:
    linear-gradient(135deg, rgb(0 0 0 / 55%) 0%, rgb(0 0 0 / 25%) 35%, rgb(0 0 0 / 0%) 65%),
    linear-gradient(rgb(0 0 0 / 10%) 0%, rgb(0 0 0 / 10%) 100%);
}

/* Scroll: the picture drifts against the page as the tile crosses the viewport.
   A scroll-driven animation rather than a scroll listener — the browser runs it
   off the same timeline it scrolls with, so there is no work on the main thread
   and nothing to throttle.

   Behind @supports on purpose. Where view() is not understood the whole rule is
   skipped and the tile is simply a tile with a photograph in it, which is the
   right thing to degrade to. background-attachment: fixed would have been the
   old way to get this and is the reason not to: iOS has never run it properly,
   and this app is mostly read on a phone. */
@supports (animation-timeline: view()) {
  /* The timeline is named on the tile and used inside it, rather than the
     frame simply asking for view() — and that indirection is not decoration.
     A bare view() measures the element against its nearest scrollport, and the
     tile is `overflow-hidden` to clip the zoom, which *makes* it one. The frame
     never moves inside the tile, so the timeline sat frozen at half progress
     and the drift did precisely nothing. Named on the tile, it is measured
     against the page, which is the thing the picture is supposed to drift
     against. */
  .tile {
    view-timeline-name: --tile;
  }

  .tile-frame {
    animation: tile-drift linear both;
    animation-timeline: --tile;

    /* The whole time the tile is on screen, entering to leaving, so the drift
       is one unbroken pass rather than something that finishes early and then
       sits still in the middle of the screen. */
    animation-range: cover 0% cover 100%;
  }
}

@keyframes tile-drift {
  from {
    transform: translateY(-4%);
  }

  to {
    transform: translateY(4%);
  }
}

/* Asked not to animate: the pictures stay, all three movements go. What is left
   is a grid of photographs that responds to nothing, which is what was asked
   for — the tiles still look like themselves. */
@media (prefers-reduced-motion: reduce) {
  .tile,
  .tile-photo,
  .tile:hover .tile-photo,
  .tile:active,
  .tile:active .tile-photo {
    transition: none;
    transform: none;
  }

  .tile-frame {
    animation: none;
  }
}

/* The DJ tile does not sit still: a gradient far wider than the tile, drifting
   back and forth across it. Slow on purpose — eighteen seconds one way — so it
   reads as the tile being alive rather than as something demanding attention,
   and `alternate` makes it turn round rather than snap back to the start.

   The stops run indigo to magenta and back, so the drift returns where it
   started and the turn is not a place where the colour jumps. Every one of them
   is dark enough to carry the white label at every point of the drift, which is
   the constraint that picked them.

   Twice the tile rather than three times: at 135 degrees the far stops sit in
   the corners, and a narrower window never brings them into view — a colour
   written here that never reaches the screen would be a colour in name only. */
.tile-oscillate {
  background-image: linear-gradient(135deg, #4f46e5 0%, #7c3aed 30%, #c026d3 55%, #7c3aed 80%, #4f46e5 100%);
  background-size: 200% 200%;
  animation: tile-oscillate 18s ease-in-out infinite alternate;
}

@keyframes tile-oscillate {
  from {
    background-position: 0% 50%;
  }

  to {
    background-position: 100% 50%;
  }
}

/* An animation that never ends is precisely what this setting is asked for.
   The colours stay — only the drifting stops, so the tile still looks like
   itself. */
@media (prefers-reduced-motion: reduce) {
  .tile-oscillate {
    animation: none;
  }
}

.spotlight {
  background: linear-gradient(45deg, #00dc82 0%, #36e4da 50%, #0047e1 100%);
  filter: blur(20vh);
  height: 40vh;
  bottom: 30vh;
}

.gradient-border {
  position: relative;
  border-radius: 0.5rem;
  -webkit-backdrop-filter: blur(10px);
  backdrop-filter: blur(10px);
}

@media (prefers-color-scheme: light) {
  .gradient-border {
    background-color: rgba(255, 255, 255, 0.3);
  }

  .gradient-border::before {
    background: linear-gradient(90deg, #e2e2e2 0%, #e2e2e2 25%, #00dc82 50%, #36e4da 75%, #0047e1 100%);
  }
}

@media (prefers-color-scheme: dark) {
  .gradient-border {
    background-color: rgba(20, 20, 20, 0.3);
  }

  .gradient-border::before {
    background: linear-gradient(90deg, #303030 0%, #303030 25%, #00dc82 50%, #36e4da 75%, #0047e1 100%);
  }
}

.gradient-border::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  border-radius: 0.5rem;
  padding: 2px;
  width: 100%;
  background-size: 400% auto;
  opacity: 0.5;
  transition: background-position 0.3s ease-in-out, opacity 0.2s ease-in-out;
  -webkit-mask: linear-gradient(#fff 0 0) content-box, linear-gradient(#fff 0 0);
  mask: linear-gradient(#fff 0 0) content-box, linear-gradient(#fff 0 0);
  -webkit-mask-composite: xor;
  mask-composite: exclude;
}

.gradient-border:hover::before {
  background-position: -50% 0;
  opacity: 1;
}
</style>
