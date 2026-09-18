<template>
  <article class="min-h-full w-full bg-elevated">
    <!-- The shelf's own bar, under the app's. Everything that changes what is
         on screen lives here rather than scattered over the wall, so the wall
         itself stays nothing but covers. -->
    <!-- Wraps rather than squeezing. Everything in here but the filter refuses
         to shrink, so on a phone the filter was the only thing that could — and
         it shrank to nothing, leaving a control that was present, focusable and
         invisible. A second line is the honest answer to not enough width. -->
    <nav class="sticky top-0 z-10 flex flex-wrap items-center gap-x-2 gap-y-1 border-b border-accented bg-default px-2 py-1">
      <NuxtLink to="/" class="flex shrink-0 items-center gap-0.5 text-sm text-muted hover:text-highlighted">
        <UIcon name="i-lucide-chevron-left" class="size-5" />
        <span class="hidden sm:inline">Zurück</span>
      </NuxtLink>

      <h1 class="shrink-0 text-sm font-semibold text-highlighted">Musik</h1>
      <span v-if="shown.length" class="shrink-0 text-xs text-muted tabular-nums">{{ shown.length }}</span>

      <UInput
        v-model="query"
        size="xs"
        variant="outline"
        icon="i-lucide-search"
        placeholder="Filtern"
        class="w-28 shrink-0 sm:w-40"
        :ui="{ base: 'h-6' }"
      />

      <!-- Only offered where it would do something: with one year on the shelf
           there is nothing to narrow.

           No year chosen is the empty value, which is the select's own way of
           saying nothing is selected — so "every year" is the placeholder
           rather than an entry in the list. It has to be got back to somehow,
           and the component offers no way of its own, hence the button beside
           it: it appears only once there is a choice to undo. -->
      <div v-if="years.length > 1" class="flex shrink-0 items-center">
        <USelect
          v-model="year"
          size="xs"
          variant="outline"
          placeholder="Alle Jahre"
          :items="yearItems"
          class="w-24"
        />
        <UButton
          v-if="year"
          size="xs"
          variant="ghost"
          color="neutral"
          icon="i-lucide-x"
          aria-label="Alle Jahre zeigen"
          title="Alle Jahre zeigen"
          @click="year = ''"
        />
      </div>

      <div class="grow" />

      <!-- How big a cover is, rather than how many fit: the wall reflows to the
           window, so a count would mean a different size on every screen. -->
      <div class="hidden shrink-0 items-center gap-1 sm:flex">
        <UIcon name="i-lucide-grid-3x3" class="size-3.5 text-dimmed" />
        <USlider
          v-model="tile" :min="80" :max="320"
          :step="20" class="w-24"
          aria-label="Kachelgröße"
        />
      </div>

      <!-- The comparison this page exists to make. Both buttons read the same
           shelf and must produce the same albums; what differs is how many
           requests it takes and how long they take, which is printed beside
           them because that is the whole question. -->
      <div class="flex shrink-0 items-center gap-1">
        <UButton
          v-for="option in strategies"
          :key="option.value"
          size="xs"
          :variant="strategy === option.value ? 'solid' : 'ghost'"
          :color="strategy === option.value ? 'primary' : 'neutral'"
          :title="option.hint"
          @click="strategy = option.value"
        >
          {{ option.label }}
        </UButton>
        <!-- Room kept for the longest reading either way can produce, so that
             the bar does not slide sideways each time the number changes. -->
        <span v-if="cost" class="w-24 shrink-0 text-right text-xs text-dimmed tabular-nums">{{ cost }}</span>
      </div>
    </nav>

    <p v-if="pending && !shelf" class="px-3 py-8 text-sm text-muted">Wird geladen…</p>

    <div v-else-if="error && !shelf" class="px-3 py-8 text-sm">
      <p class="text-error">{{ problem.what }}</p>
      <p v-if="problem.why" class="mt-1 text-muted">{{ problem.why }}</p>
    </div>

    <p v-else-if="!shelf?.albums.length" class="px-3 py-8 text-sm text-muted">
      Hier steht noch nichts.
    </p>

    <p v-else-if="!shown.length" class="px-3 py-8 text-sm text-muted">
      Nichts gefunden.
    </p>

    <!-- The wall. No titles under the covers and no artists beside them: a
         album is recognised by its sleeve, and the name of everything at once
         is noise. What an album is called is in its tooltip and in the window
         that opens on it. -->
    <ul
      v-else
      ref="wall"
      class="grid gap-2 p-2"
      :style="{ gridTemplateColumns: `repeat(auto-fill, minmax(${tile}px, 1fr))` }"
    >
      <li
        v-for="album in shown"
        :key="album.path"
        class="group relative aspect-square overflow-hidden rounded-sm bg-black"
      >
        <img
          v-if="album.cover"
          :data-src="covers(album.cover, coverWidth)"
          :alt="album.title"
          src="data:image/gif;base64,R0lGODlhAQABAAAAACH5BAEKAAEALAAAAAABAAEAAAICTAEAOw=="
          class="size-full object-cover transition duration-300 group-hover:scale-105"
        >
        <div v-else class="flex size-full items-center justify-center p-2 text-center text-xs text-dimmed">
          {{ album.title }}
        </div>

        <!-- Both handles appear together on hover, and on a touch screen there
             is no hover at all — so the whole tile opens the album, and the
             play button is the only thing that does something else. -->
        <button
          type="button"
          class="absolute inset-0 cursor-pointer"
          :title="album.year ? `${album.title} (${album.year})` : album.title"
          @click="opened = album"
        >
          <span class="sr-only">{{ album.title }}</span>
        </button>

        <div
          class="pointer-events-none absolute inset-x-0 bottom-0 flex items-end justify-between gap-1 bg-gradient-to-t from-black/80 to-transparent p-1.5 opacity-0 transition group-hover:opacity-100 focus-within:opacity-100"
        >
          <UButton
            v-if="album.tracks.length"
            size="xs"
            color="neutral"
            variant="solid"
            icon="i-lucide-play"
            class="pointer-events-auto rounded-full"
            :aria-label="`${album.title} abspielen`"
            :title="`${album.title} abspielen`"
            @click.stop="playAlbum(album)"
          />
          <span v-else class="text-xs text-white/70">keine Titel</span>

          <UButton
            size="xs"
            color="neutral"
            variant="ghost"
            icon="i-lucide-info"
            class="pointer-events-auto rounded-full text-white hover:bg-white/20"
            :aria-label="`${album.title} — Details`"
            :title="`${album.title} — Details`"
            @click.stop="opened = album"
          />
        </div>
      </li>
    </ul>

    <!-- The album itself: its sleeve, and under it what is on it. No page of
         its own, because an album is a detail of the shelf and not a place —
         closing it should put somebody back where they were looking. -->
    <UModal v-model:open="detailsOpen" :title="opened?.title ?? ''" :ui="{ content: 'max-w-lg' }">
      <template #body>
        <div v-if="opened" class="space-y-3">
          <div class="flex gap-3">
            <img
              v-if="opened.cover"
              :src="covers(opened.cover, 320)"
              :alt="opened.title"
              class="size-32 shrink-0 rounded-sm bg-black object-cover"
            >
            <div class="min-w-0">
              <!-- Not the title: the window's own header carries that, and
                   printing it again underneath it merely said it twice. The
                   artist is what belongs here — it is the one thing the shelf
                   itself cannot say, since no folder and no filename holds it. -->
              <p v-if="albumArtist" class="truncate text-base font-semibold text-highlighted">
                {{ albumArtist }}
              </p>
              <p class="text-sm text-muted">
                <span v-if="openedYear">{{ openedYear }} · </span>
                {{ opened.tracks.length }} Titel
                <span v-if="playtime(opened)"> · {{ playtime(opened) }}</span>
                <span v-if="albumGenre"> · {{ albumGenre }}</span>
              </p>
              <!-- The folder, spelled out: it is the only name the storage
                   knows this album by, and the thing to look for when a cover
                   or a track is missing. -->
              <p class="mt-1 text-xs break-all text-dimmed">{{ opened.path }}</p>
            </div>
          </div>

          <!-- Where the titles come from. Offered rather than decided,
               because the two disagree in both directions: the tags spell
               things the way whoever ripped the album did, the filenames the
               way whoever filed it did, and neither is reliably the better. The
               switch is only there when there is something to switch to. -->
          <div v-if="opened.tracks.length" class="flex items-center justify-between border-t border-accented pt-2">
            <span class="text-xs text-dimmed">
              <template v-if="tagsPending">Tags werden gelesen…</template>
              <template v-else-if="tagged">{{ tagFormat }}</template>
              <template v-else>Keine Tags in den Dateien</template>
            </span>
            <div v-if="tagged" class="flex items-center gap-1">
              <UButton
                v-for="source in sources"
                :key="source.value"
                size="xs"
                :variant="fromTags === source.value ? 'solid' : 'ghost'"
                :color="fromTags === source.value ? 'primary' : 'neutral'"
                @click="fromTags = source.value"
              >
                {{ source.label }}
              </UButton>
            </div>
          </div>

          <ol v-if="opened.tracks.length" class="divide-y divide-accented border-t border-accented">
            <li
              v-for="(item, index) in listed"
              :key="item.path"
              class="flex items-baseline gap-2 py-1.5"
              :class="{ 'text-primary': item.path === current?.id }"
            >
              <span class="w-5 shrink-0 text-right text-xs text-muted tabular-nums">
                {{ item.number ?? index + 1 }}
              </span>
              <button
                type="button"
                class="min-w-0 grow cursor-pointer text-left hover:text-primary"
                @click="playFrom(index)"
              >
                <span class="block truncate text-sm text-highlighted">{{ item.title }}</span>
                <!-- What the file says about itself, under what it is called —
                     the shelf's own two-line shape. Only where the tags are
                     what is being shown: the filenames carry none of this, and
                     an empty line under each of them would say so twelve
                     times. -->
                <span v-if="item.meta" class="block truncate text-xs font-light text-dimmed">{{ item.meta }}</span>
              </button>
              <span class="shrink-0 text-xs text-dimmed tabular-nums">
                {{ item.length || driveSize(item.track.size) }}
              </span>
            </li>
          </ol>

          <p v-else class="text-sm text-muted">
            Von dieser Platte liegt noch nichts da — nur das Cover.
          </p>
        </div>
      </template>
    </UModal>

  </article>
</template>

<script setup lang="ts">
// The music shelf: a wall of covers, and what is on each.
//
// Two things here are deliberate and would otherwise look like omissions. There
// are no artist names on the wall — the shelf does not record them, and reading
// them out of the files would mean opening every one of them to draw a page of
// pictures. And there are no subpages: an album opens over the wall and closes
// back onto it, because it is a detail of the shelf rather than a place in it.
//
// The two ways of reading the shelf are switchable rather than chosen, because
// which one is better depends on how the shelf grows. See useMusik.
const { shelf: load, cover: covers, track: tracks, tags: readTags } = useMusik()

// The player is the app's, not this page's: an album goes on playing while you
// walk off to the magazines. See usePlayer.
const { current, play } = usePlayer()

const strategies: { value: Strategy, label: string, hint: string }[] = [
  { value: 'dir', label: 'Ordner', hint: 'Ein Verzeichnis pro Anfrage, alle gleichzeitig — schnell, aber viele' },
  { value: 'search', label: 'Suche', hint: 'Das ganze Regal in drei Anfragen — wenige, aber langsam' },
]

const strategy = ref<Strategy>('dir')

// One key rather than one per strategy, which is what keeps the wall on screen
// while the other reading loads.
//
// Keyed by strategy, switching moved to a piece of state that had never been
// filled, so the covers vanished and were replaced by the word "loading" for
// the eleven seconds a search takes — and the two readings produce the same
// albums, so there was nothing to redraw at the end of it. Under one key the
// previous answer stays until the new one arrives.
//
// It also means each switch measures again rather than reporting a remembered
// number, which for a page whose purpose is the measurement is the right way
// round.
const { data: shelf, pending, error } = await useAsyncData(
  'musik-shelf',
  () => load(strategy.value),
  { watch: [strategy] },
)

// Blank while a reading is in flight: the number belongs to the reading that
// produced it, and leaving the old one up next to the newly-pressed button
// credits one way of reading the shelf with the other's speed.
const cost = computed(() => {
  if (pending.value) return '…'
  if (!shelf.value) return ''
  const { requests, took } = shelf.value

  return `${requests} × ${Math.round(took)} ms`
})

const query = ref('')

// No year chosen is the empty string, because that is what the select means by
// it — an item carrying that value is refused outright, since it could not be
// told apart from having cleared the selection.
const year = ref('')
const tile = ref(160)

// How large a sleeve to ask the storage for.
//
// Not the tile's own size: a screen may have two pixels to the tile's one, so
// asking for exactly the displayed width gives a soft picture on half the
// devices there are. Twice it, rounded up to a step, because the alternative —
// a width per slider position — would fetch the whole wall again on every
// nudge, and because a browser can then reuse a sleeve across sizes.
//
// It is worth the arithmetic. The storage honours the width and the sizes are
// far apart: measured on one cover, 3 kB at 100 wide, 8 kB at 200, 53 kB at
// 480. Asking for 480 to draw a 160-pixel tile is six times the bytes for a
// picture nobody can see the detail of, which over a wall of two hundred is ten
// megabytes against under two.
const coverWidth = computed(() =>
  [200, 320, 480, 640].find(width => width >= tile.value * 2) ?? 640)
const opened = ref<Album | undefined>()

/**
 * The years to offer, which are the years the *folders* carry.
 *
 * Not the years in the tags, and the difference shows: Trash is filed under no
 * year at all and its files say 1989, so the window that opens on it reads 1989
 * while this list does not offer it. That is a consequence of the tags being
 * read only when an album is opened — building this from them would mean
 * reading every file on the shelf to draw a list of years, which is the same
 * cost that keeps the artists off the wall.
 *
 * Whichever album has no year is still reachable: it is simply not under a
 * year, and the filter beside this one finds it by name.
 */
const years = computed(() =>
  [...new Set((shelf.value?.albums ?? []).map(a => a.year).filter(Boolean) as string[])].sort().reverse())

const yearItems = computed(() => years.value.map(y => ({ label: y, value: y })))

const shown = computed(() =>
  (shelf.value?.albums ?? [])
    .filter(album => !year.value || album.year === year.value)
    .filter(album => albumMatches(album, query.value)))

// The modal's openness is the album, so that closing it by any means — the
// escape key included — puts the album back to none rather than leaving a
// stale one behind the next time something is opened.
const detailsOpen = computed({
  get: () => !!opened.value,
  set: (open: boolean) => {
    if (!open) opened.value = undefined
  },
})

// What the files say about themselves, fetched when an album is opened and not
// before: the wall shows no artists, so drawing it needs none of this, and
// reading it is a request per file on the server.
//
// Held per album rather than refetched, because opening the same one twice in
// a sitting is the ordinary way to use this page. The server caches too; this
// saves the round trip as well.
const tagsFor = ref<Record<string, MusikTags[]>>({})
const tagsPending = ref(false)

const sources = [
  { value: true, label: 'Tags' },
  { value: false, label: 'Dateinamen' },
]

const fromTags = ref(true)

watch(opened, async (album) => {
  if (!album || tagsFor.value[album.path]) return

  tagsPending.value = true
  try {
    tagsFor.value[album.path] = (await readTags(album.path)).tracks
  }
  catch (err) {
    // Tags are an improvement on the filenames, not a requirement: an album
    // whose tags cannot be read still lists and still plays.
    console.warn('could not read the tags of', album.path, err)
    tagsFor.value[album.path] = []
  }
  finally {
    tagsPending.value = false
  }
})

/** The open album's tags, by filename — which is what lines them up. */
const openedTags = computed(() => {
  const read = opened.value ? tagsFor.value[opened.value.path] ?? [] : []

  return new Map(read.filter(t => t.read).map(t => [t.name, t]))
})

const tagged = computed(() => openedTags.value.size > 0)

/** Which tag format the files carry, named because it is the first thing worth
 * knowing when one reads oddly. */
const tagFormat = computed(() => {
  const formats = new Set([...openedTags.value.values()].map(t => t.format).filter(Boolean))

  return formats.size === 1 ? [...formats][0] : `${openedTags.value.size} Dateien mit Tags`
})

/**
 * The album's artist.
 *
 * The album artist where the files name one, because a compilation names a
 * different artist on every track and the album still has one. Failing that
 * the first track's, and failing that nothing — this is the only place the
 * artist can come from at all.
 */
const albumArtist = computed(() => {
  const all = [...openedTags.value.values()]
  const named = all.find(t => t.album_artist)?.album_artist ?? all.find(t => t.artist)?.artist
  if (!named) return ''

  // An album whose tracks disagree is a compilation; saying one of the artists
  // would be wrong about all the others.
  const artists = new Set(all.map(t => t.artist).filter(Boolean))

  return artists.size > 1 && !all.some(t => t.album_artist) ? 'Verschiedene' : named
})

const albumGenre = computed(() => [...openedTags.value.values()].find(t => t.genre)?.genre ?? '')

// The folder's year where it has one, the files' where it has not — two of the
// albums on this shelf carry no year in the folder name at all.
const openedYear = computed(() =>
  opened.value?.year ?? [...openedTags.value.values()].find(t => t.year)?.year)

/**
 * The open album's tracks as the list shows them.
 *
 * One shape whichever source is chosen, so the list itself does not branch. The
 * order stays the filenames' — they are what the shelf is sorted by and what a
 * reader has just been looking at — and only what is *shown* changes, because
 * re-sorting the list under somebody when they flip the switch would lose their
 * place in it.
 */
const listed = computed(() => (opened.value?.tracks ?? []).map((track) => {
  // The file, as it is stored — the whole name, extension and all.
  //
  // Literally what the switch offers, and the only reading of it that shows
  // anything: the titles read *out of* the filenames are what the other side
  // already falls back to, so offering those would be a switch between two
  // identical lists on a shelf as tidily named as this one. What somebody wants
  // from this position is to see what the file is actually called.
  if (!fromTags.value) {
    return {
      path: track.path,
      number: track.number,
      title: track.name,
      meta: '',
      length: clock(openedTags.value.get(track.name)?.seconds),
      track,
    }
  }

  const tag = openedTags.value.get(track.name)
  if (!tag) {
    return { path: track.path, number: track.number, title: track.title, meta: 'ohne Tags', length: '', track }
  }

  return {
    path: track.path,
    number: tag.track || track.number,
    title: tag.title || track.title,
    meta: describe(tag),
    length: clock(tag.seconds),
    track,
  }
}))

/** Seconds as a listing shows them: 4:29, and 1:15:57 for a long album. */
function clock(seconds?: number): string {
  if (!seconds) return ''

  const hours = Math.floor(seconds / 3600)
  const minutes = Math.floor(seconds % 3600 / 60)
  const rest = seconds % 60
  if (hours) {
    return `${hours}:${String(minutes).padStart(2, '0')}:${String(rest).padStart(2, '0')}`
  }

  return `${minutes}:${String(rest).padStart(2, '0')}`
}

// What the heading has already said, which is what a track's own line leaves
// out. The album is not in the heading but belongs here all the same: a folder
// holding two albums is worth seeing, and one holding a single album is not
// worth being told about ten times.
const header = computed<AlbumHeader>(() => ({
  artist: albumArtist.value,
  album: agreed([...openedTags.value.values()].map(t => t.album)),
  year: openedYear.value ? String(openedYear.value) : '',
  genre: albumGenre.value,
}))

function describe(tag: MusikTags): string {
  return trackMeta(tag, header.value)
}

/**
 * An album as something to play: finished URLs, in order.
 *
 * The addresses are resolved here because the player is shared with the DJ
 * archive and knows about neither shelf. Handing over the whole album rather
 * than the chosen track is what lets it go on to the next one after this page
 * has been navigated away from.
 */
function queued(album: Album, titles?: { path: string, title: string, meta?: string }[]): PlayerTrack[] {
  const said = new Map((titles ?? []).map(t => [t.path, t]))
  const sleeve = album.cover ? covers(album.cover, 80) : undefined

  return album.tracks.map(track => ({
    id: track.path,
    src: tracks(track.path),
    // What the open window shows, where it is open: the tags are read only
    // then, and a queue built from the wall has nothing better than the
    // filename — which is what the wall itself is showing.
    title: said.get(track.path)?.title || track.title,
    subtitle: said.get(track.path)?.meta || album.title,
    cover: sleeve,
  }))
}

// Starts the album, and nothing else. The tile carries two handles because
// they do two things: opening the window as well would make the quieter of the
// two the only one that merely plays.
function playAlbum(album: Album) {
  play(queued(album))
}

/** From one track of the open album onwards. */
function playFrom(at: number) {
  if (opened.value) play(queued(opened.value, listed.value), at)
}

/**
 * How long an album runs.
 *
 * Exactly, where every one of its files says how long it is — which on this
 * shelf is all of them, the length being either written into the tag or
 * derivable from the frame count a variable-bitrate file carries.
 *
 * The estimate from file sizes is the fallback and is labelled as one, because
 * it is a guess: it assumes a bitrate the encoder was never obliged to keep to,
 * and against the exact figures it was out by twelve per cent on one album.
 */
function playtime(album: Album): string {
  const timed = album.tracks.map(t => openedTags.value.get(t.name)?.seconds ?? 0)
  if (timed.length && timed.every(Boolean)) {
    return clock(timed.reduce((sum, n) => sum + n, 0))
  }

  const bytes = album.tracks.reduce((sum, t) => sum + (t.size ?? 0), 0)
  if (!bytes) return ''

  const minutes = Math.round(bytes * 8 / (192 * 1000) / 60)

  return minutes ? `ca. ${minutes} min` : ''
}

// Covers are fetched when they come into view, and not before — the same
// reasoning as the archive's: a wall of two hundred sleeves is two hundred
// requests otherwise, most of them for something nobody scrolled to.
const wall = useTemplateRef<HTMLElement>('wall')
let watching: IntersectionObserver | undefined

function watchCovers() {
  watching?.disconnect()
  if (!wall.value) return

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
  }, { rootMargin: '300px' })

  for (const cover of wall.value.querySelectorAll<HTMLImageElement>('img[data-src]')) {
    watching.observe(cover)
  }
}

onMounted(watchCovers)
watch(shown, () => nextTick(watchCovers))
onBeforeUnmount(() => watching?.disconnect())

// The three ways this arrives as "it did not work" are worth telling apart: one
// is a setting nobody filled in, one is a permission, and only the third is a
// fault.
const problem = computed(() => {
  switch ((error.value as { statusCode?: number } | null)?.statusCode) {
    case 501:
      return {
        what: 'Für diese Installation ist kein Musikregal eingerichtet.',
        why: 'Es fehlt die Einstellung MUSIK_ROOT — der Ordner, in dem die Alben liegen.',
      }
    case 403:
      return {
        what: 'Dieses Konto ist nicht für die Musik freigeschaltet.',
        why: 'Die Berechtigung heißt „musik“.',
      }
    case 401:
      return { what: 'Bitte anmelden.', why: '' }
    default:
      return { what: 'Das Regal ist nicht erreichbar.', why: '' }
  }
})
</script>
