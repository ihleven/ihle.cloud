<template>
  <div class="relative min-h-full w-full bg-elevated">
    <p v-if="failed" class="px-3 py-8 text-sm text-error">
      Diese Ausgabe lässt sich nicht öffnen.
    </p>

    <p v-else-if="!pageCount" class="px-3 py-8 text-sm text-muted">
      Wird geöffnet…
    </p>

    <!-- A placeholder per page from the start, each at its own page's shape.
         Without a height the whole document would sit at one scroll position
         and every page would count as visible at once, which is the opposite of
         drawing them as they are reached.

         Edge to edge, with no margin of its own: whoever shows this decides
         how much room it gets, and over the whole screen the answer is all of
         it. The page is the thing being read; anything around it is width the
         scan does not get.

         Light, and each sheet told apart by a shadow rather than by an edge
         against a dark field. A scan carries the margin the scanner left around
         the paper; against dark that margin meets a hard line and reads as a
         white frame around the ink, which is not something in the file and not
         something any viewer can remove. -->
    <div v-else ref="reel" class="flex flex-col gap-2">
      <div
        v-for="n in pageCount"
        :key="n"
        :data-page="n"
        class="relative w-full bg-default shadow-sm ring-1 ring-default"
        :style="{ aspectRatio: shapes[n - 1] ?? shapes[0] }"
      >
        <canvas :ref="el => keep(n, el as HTMLCanvasElement | null)" class="absolute inset-0 h-full w-full" />
      </div>
    </div>

    <!-- Where you are, out of the way. Inside the reader rather than in either
         of the two headers that host it, so both get it without being told. -->
    <p
      v-if="pageCount"
      class="pointer-events-none sticky bottom-2 mx-auto w-fit rounded-full bg-inverted/80 px-2 py-0.5 text-xs text-inverted tabular-nums"
    >
      {{ visible }} / {{ pageCount }}
    </p>
  </div>
</template>

<script setup lang="ts">
import * as pdfjs from 'pdfjs-dist'
import type { PDFDocumentProxy } from 'pdfjs-dist'
import workerSrc from 'pdfjs-dist/build/pdf.worker.min.mjs?url'

// Reading a scan inside the app rather than handing it to the browser.
//
// The archive is the one part of this app whose whole purpose is a document, so
// it owns the view of one. Handing the file over instead replaced the app with
// it, and an installed app has nothing to come back with — see useStandalone.
// An <iframe> is not the cheaper version of this: iOS renders only the first
// page of a PDF in a frame and will not scroll it.
//
// An issue is fifty to a hundred and twenty megabytes, so nothing here reads
// the whole file. pdf.js asks for byte ranges, the media route proxies them to
// the storage rather than redirecting, and a reverse proxy passes Range through
// — which is what makes opening page one of a ninety-megabyte scan cost a page
// rather than a magazine.
//
// The pages and nothing else: whoever shows this supplies the frame around it,
// because a page of the app and a sheet sliding up over it want different ones.
pdfjs.GlobalWorkerOptions.workerSrc = workerSrc

const props = defineProps<{ path: string }>()

const { issue, encodePath } = useRetro()

const doc = shallowRef<PDFDocumentProxy | null>(null)
const pageCount = ref(0)
const failed = ref(false)

/**
 * Each page's proportions, in order.
 *
 * Per page rather than one shape for the document: a scanned magazine is not
 * uniform — a cover, a fold-out or a page fed in the other way round comes out
 * a different size — and holding every page to the first one's box stretches
 * those, which shows as the picture not meeting its own edges.
 *
 * Measured before anything is drawn. Reading a page's size does not render it,
 * so this is the page dictionaries and no more, and doing it up front means the
 * reel does not resize under the reader as pages arrive.
 */
const shapes = ref<string[]>([])

const canvases = new Map<number, HTMLCanvasElement>()
const drawn = new Set<number>()
const visible = ref(1)

let watching: IntersectionObserver | undefined

/**
 * How many pages stay drawn.
 *
 * A rendered page is a canvas of a few megabytes, so a hundred-page magazine
 * left fully drawn is hundreds of megabytes of them — which a phone answers by
 * discarding the tab. Pages are drawn as they are reached and the ones furthest
 * from the reader are cleared again.
 */
const keepDrawn = 6

function keep(n: number, el: HTMLCanvasElement | null) {
  if (el) canvases.set(n, el)
  else canvases.delete(n)
}

async function draw(n: number) {
  const pdf = doc.value
  const canvas = canvases.get(n)
  if (!pdf || !canvas || drawn.has(n)) return

  drawn.add(n)
  const page = await pdf.getPage(n)

  // Drawn at the pixels the screen actually has, not at the CSS size: a phone
  // has two or three of them to the CSS pixel, and a scan drawn at the smaller
  // number is a photograph of print that has been through a blur.
  const width = canvas.clientWidth * Math.min(window.devicePixelRatio || 1, 2)
  const unscaled = page.getViewport({ scale: 1 })
  const viewport = page.getViewport({ scale: width / unscaled.width })

  canvas.width = Math.floor(viewport.width)
  canvas.height = Math.floor(viewport.height)

  // The canvas itself, not its context: pdf.js takes the context only for
  // backwards compatibility, and then only when no canvas is given.
  await page.render({ canvas, viewport }).promise
  forget(n)
}

/** Clear whatever is furthest from what is being read. */
function forget(around: number) {
  if (drawn.size <= keepDrawn) return

  const far = [...drawn].sort((a, b) => Math.abs(b - around) - Math.abs(a - around))
  for (const n of far.slice(0, drawn.size - keepDrawn)) {
    const canvas = canvases.get(n)
    if (canvas) {
      canvas.width = 0
      canvas.height = 0
    }
    drawn.delete(n)
  }
}

const reel = useTemplateRef<HTMLElement>('reel')

function watchPages() {
  watching?.disconnect()
  if (!reel.value) return

  watching = new IntersectionObserver((entries) => {
    for (const entry of entries) {
      const n = Number((entry.target as HTMLElement).dataset.page)
      if (!entry.isIntersecting) continue

      visible.value = n
      draw(n)
    }
    // A page ahead and a page behind, so turning the page finds it drawn.
  }, { rootMargin: '200% 0px' })

  for (const page of reel.value.querySelectorAll<HTMLElement>('[data-page]')) {
    watching.observe(page)
  }
}

async function open() {
  try {
    const task = pdfjs.getDocument({ url: issue(encodePath(props.path)), withCredentials: true })
    const pdf = await task.promise

    doc.value = pdf
    pageCount.value = pdf.numPages

    shapes.value = await Promise.all(
      Array.from({ length: pdf.numPages }, async (_, i) => {
        const { width, height } = (await pdf.getPage(i + 1)).getViewport({ scale: 1 })

        return `${width} / ${height}`
      }),
    )

    await nextTick()
    watchPages()
  }
  catch (err) {
    console.warn('could not open the issue:', props.path, err)
    failed.value = true
  }
}

onMounted(open)

onBeforeUnmount(() => {
  watching?.disconnect()
  doc.value?.destroy()
})
</script>
