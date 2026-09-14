<template>
  <!-- An account that exists only for the pool gets the pool's own front door,
       reproduced from ihleven.de: no chrome from this app around it, because
       none of it would lead anywhere they may go. -->
  <main
    v-if="confined"
    class="ght-door grid min-h-screen place-content-center overflow-hidden font-sans antialiased"
  >
    <div class="spotlight fixed right-0 left-0 z-10" />
    <div class="z-20 max-w-[520px] text-center">
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
    <!-- Two columns until there is room to auto-fit 200px tiles. A phone is
         360–390px wide, so auto-fit on its own drops to one column and the page
         becomes a long scroll of a single tile at a time.

         w-full, not w-screen: this sits inside UMain, which is a flex
         container, and 100vw is the viewport including any scrollbar — wider
         than the space it has been given. The overflow shows as a pale strip
         down the right and a page that scrolls sideways. -->
    <main class="grid w-full grid-cols-2 content-start sm:grid-cols-[repeat(auto-fit,_minmax(200px,_1fr))]">
      <section class="flex items-center justify-between bg-ral-7035">
        <Logo />
      </section>

      <!-- Only the areas this account is entitled to. Someone who may see one
           thing gets one tile, not a wall of doors that refuse to open. -->
      <NuxtLink
        v-for="tile in tiles"
        :key="tile.label"
        :to="tile.to"
        class="block aspect-square p-8"
        :class="tile.class"
      >
        <h1 class="text-lg font-black text-white hover:text-outline">{{ tile.label }}</h1>
      </NuxtLink>

      <p v-if="!tiles.length" class="col-span-full p-8 text-muted">
        Für dieses Konto ist noch nichts freigeschaltet.
      </p>
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

// A confined account is reported as entitled to the pool and nothing else, by
// its type rather than by its permissions; see Service.offered.
const confined = computed(() =>
  session.value?.modules.length === 1 && session.value.modules[0] === 'geheimtipp',
)

// Each tile names the area it belongs to; the entitlement decides whether it is
// rendered at all.
const tiles = computed(() => allowed([
  { module: 'content', label: 'CMS', to: '/entries', class: 'bg-blue-500/80' },
  { module: 'filme', label: 'Super 8', to: '/filme', class: 'bg-rose-500/80' },
  { module: 'familie', label: 'Familie', to: '/famihlie', class: 'bg-violet-500/80' },
  { module: 'hidrive', label: 'Dateien', to: '/hidrive', class: 'bg-amber-500/80' },
  { module: 'kalender', label: 'Kalender', to: '/kalender', class: 'bg-green-500/90' },
  { module: 'mediathek', label: 'Mediathek', to: '/mediathek', class: 'bg-sky-500/80' },
  { module: 'musik', label: 'Musik', to: '/musik', class: 'bg-cyan-300/80' },
  { module: 'geheimtipp', label: 'Geheimtipp', to: '/geheimtipp', class: 'bg-sky-400/80' },
]))
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
