<template>
  <article ref="page" class="min-h-full w-full bg-elevated">
    <header class="flex flex-wrap items-baseline gap-2 px-3 py-2">
      <h1 class="text-sm font-semibold text-highlighted">Zeitschriften</h1>
      <span v-if="issueCount" class="text-xs text-muted">{{ issueCount }} Ausgaben</span>

      <div class="grow" />

      <!-- Which reader opens an issue. Switchable rather than chosen, because
           the two behave differently enough on a phone to be worth comparing
           against real scans: the app's own keeps you inside it, the system's
           is the one iOS draws and — installed — the one you cannot come back
           from. -->
      <div class="flex items-center gap-1">
        <UButton
          v-for="choice in readers"
          :key="choice.value"
          size="xs"
          :variant="reader === choice.value ? 'solid' : 'ghost'"
          :color="reader === choice.value ? 'primary' : 'neutral'"
          :title="choice.hint"
          @click="reader = choice.value"
        >
          {{ choice.label }}
        </UButton>
      </div>
    </header>

    <p v-if="pending" class="px-3 py-8 text-sm text-muted">Wird geladen…</p>

    <div v-else-if="error" class="px-3 py-8 text-sm">
      <p class="text-error">{{ problem.what }}</p>
      <p v-if="problem.why" class="mt-1 text-muted">{{ problem.why }}</p>
    </div>

    <p v-else-if="!shelf?.length" class="px-3 py-8 text-sm text-muted">
      Hier steht noch nichts.
    </p>

    <section v-for="magazine in shelf" v-else :key="magazine.dir" class="pb-4">
      <h2 class="px-3 pt-3 pb-1 text-sm font-semibold text-highlighted">{{ magazine.title }}</h2>

      <section v-for="group in magazine.groups" :key="group.title" class="pb-2">
        <h3 class="flex items-baseline gap-2 px-3 py-1">
          <span class="text-xs font-medium text-default">{{ group.title }}</span>
          <span class="text-xs text-muted">{{ group.issues.length }}</span>
        </h3>

        <!-- One row per year, scrolled sideways, and it stops on a cover rather
             than between two: scroll-snap gives each tile a line the row comes
             to rest on, so a flick always leaves a whole issue at the left edge.
             scroll-px matches the row's own padding, or the first tile would
             snap under it. -->
        <ul
          class="flex snap-x snap-mandatory scroll-px-3 gap-1 overflow-x-auto overscroll-x-contain px-3 pb-2"
        >
          <li
            v-for="issue in group.issues"
            :key="issue.name"
            class="w-32 shrink-0 snap-start sm:w-36"
          >
            <!-- Written as a plain link to the file, which is what the
                 system reader wants and what every browser already knows how
                 to do. The app's reader takes the click instead. -->
            <a
              :href="issue.href"
              target="_blank"
              rel="noopener"
              class="group block"
              :title="issue.title"
              @click="openIssue($event, issue)"
            >
              <div class="relative aspect-[3/4] overflow-hidden rounded-sm bg-black">
                <img
                  :data-src="issue.cover"
                  :alt="issue.title"
                  src="data:image/gif;base64,R0lGODlhAQABAAAAACH5BAEKAAEALAAAAAABAAEAAAICTAEAOw=="
                  class="absolute inset-0 size-full object-cover transition duration-300 group-hover:scale-105"
                >
              </div>
              <p class="flex items-baseline gap-1 pt-1 text-xs">
                <span class="truncate text-muted">{{ issue.label }}</span>
                <span class="shrink-0 text-dimmed tabular-nums">{{ driveSize(issue.size) }}</span>
              </p>
            </a>
          </li>
        </ul>
      </section>
    </section>

    <!-- Up from the foot of the screen, over everything including the app's own
         bar: reading a magazine is the whole screen's job, and the bar above it
         would only offer somewhere else to be. Its own header carries what is
         being read and the way out. -->
    <USlideover
      v-model:open="reading"
      side="bottom"
      :title="opened?.title ?? ''"
      :ui="{
        content: 'h-dvh max-h-dvh',
        // No padding at any width. p-0 on its own loses to the sm: variant the
        // slideover sets, which is how a phone came out right and a tablet did
        // not.
        body: 'overflow-y-auto p-0 sm:p-0',
        // The title bar as short as it can be and still hold a title: it is
        // sixty-four pixels by default, and every one of them is a strip of the
        // scan somebody is trying to read.
        header: 'min-h-0 p-2 sm:px-3',
        close: 'top-2 end-2',
      }"
    >
      <template #body>
        <RetroReader v-if="opened" :key="opened.path" :path="opened.path" />
      </template>
    </USlideover>
  </article>
</template>

<script setup lang="ts">
// The magazine archive: every issue as its cover, opening the scan itself.
//
// An issue is a PDF of seventy to ninety megabytes, so nothing here downloads
// one to show it — the covers are rendered by the storage, and the scan is
// opened only when somebody asks for it, from a route that serves ranges so a
// reader can jump about in it without fetching the whole thing.
const { shelf: load } = useRetro()

type Reader = 'app' | 'system'

const readers: { value: Reader, label: string, hint: string }[] = [
  { value: 'app', label: 'In der App', hint: 'Die Ausgabe öffnet sich über der Seite und lässt sich wieder schließen' },
  { value: 'system', label: 'Im Betrachter', hint: 'Die Ausgabe wird an den Browser übergeben und öffnet sich in einem eigenen Fenster' },
]

// Kept across visits, so a comparison survives walking off to look at something
// else and coming back.
const reader = useState<Reader>('retro-reader', () => 'app')

const { openOutside } = useStandalone()

const opened = ref<Issue | undefined>()
const reading = ref(false)

/**
 * Open an issue with whichever reader is chosen.
 *
 * The app's own is the sheet, and taking the click is what stops the browser
 * from navigating away underneath it.
 *
 * The system reader is the browser's, but it still must not be given this
 * window: installed, there is no address bar, no back button and no pull to
 * reload, so a scan opened in place is the last thing the app ever shows. It
 * gets a window of its own instead, which leaves this one standing behind it.
 * In an ordinary tab the link is left alone — there the browser's own chrome is
 * the way back. See useStandalone.
 */
function openIssue(event: MouseEvent, issue: Issue) {
  if (reader.value === 'app') {
    event.preventDefault()
    opened.value = issue
    reading.value = true

    return
  }

  openOutside(event, issue.href)
}

const { data: shelf, pending, error } = await useAsyncData('retro-shelf', load)

const issueCount = computed(() =>
  (shelf.value ?? []).reduce((n, m) => n + m.groups.reduce((k, g) => k + g.issues.length, 0), 0))

// Covers are fetched when they come into view, and not before.
//
// loading="lazy" is not enough, and it got worse rather than better when the
// grid became rows: a browser counts everything inside a horizontally scrolled
// row as near the viewport, so all ninety-three covers were asked for at once —
// more than the grid ever asked for. An observer judges what is actually
// visible, because the intersection it reports is clipped by the row the tile
// sits in.
const page = useTemplateRef<HTMLElement>('page')
let watching: IntersectionObserver | undefined

function watchCovers() {
  watching?.disconnect()
  if (!page.value) return

  watching = new IntersectionObserver((entries) => {
    for (const entry of entries) {
      if (!entry.isIntersecting) continue

      const cover = entry.target as HTMLImageElement
      if (cover.dataset.src) {
        cover.src = cover.dataset.src
        delete cover.dataset.src
      }
      watching?.unobserve(cover)
    }
    // A margin, so a cover is on its way by the time it is scrolled to rather
    // than starting then.
  }, { rootMargin: '300px' })

  for (const cover of page.value.querySelectorAll<HTMLImageElement>('img[data-src]')) {
    watching.observe(cover)
  }
}

onMounted(watchCovers)
watch(shelf, () => nextTick(watchCovers))
onBeforeUnmount(() => watching?.disconnect())

// The three ways this arrives as "it did not work" are worth telling apart: one
// is a setting nobody filled in, one is a permission, and only the third is a
// fault.
const problem = computed(() => {
  switch ((error.value as { statusCode?: number } | null)?.statusCode) {
    case 501:
      return {
        what: 'Für diese Installation ist kein Archiv eingerichtet.',
        why: 'Es fehlt die Einstellung RETRO_ROOT — der Ordner, in dem die Zeitschriften liegen.',
      }
    case 403:
      return {
        what: 'Dieses Konto ist nicht für das Archiv freigeschaltet.',
        why: 'Die Berechtigung heißt „retro“.',
      }
    case 401:
      return { what: 'Bitte anmelden.', why: '' }
    default:
      return { what: 'Das Archiv ist nicht erreichbar.', why: '' }
  }
})
</script>
