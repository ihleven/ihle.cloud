<template>
  <article class="mx-auto max-w-2xl p-6">
    <header class="mb-6">
      <h1 class="text-2xl font-bold text-highlighted">
        {{ title }}
      </h1>
      <p
        v-if="c.year"
        class="text-muted"
      >
        {{ c.year }}
      </p>

      <!-- The old database's DEL, carried through as an inactive entry. Said
           plainly rather than hidden: the work is in the archive on purpose,
           and somebody reading the page should know which shelf it is on. -->
      <UBadge
        v-if="entry.status === 'inactive'"
        color="neutral"
        variant="subtle"
        label="Nicht mehr im Bestand"
        class="mt-2"
      />
    </header>

    <dl class="grid grid-cols-3 gap-x-4 gap-y-2 text-sm">
      <template
        v-for="f in facts"
        :key="f.label"
      >
        <dt class="col-span-1 text-muted">
          {{ f.label }}
        </dt>
        <dd class="col-span-2 text-highlighted">
          {{ f.value }}
        </dd>
      </template>
    </dl>

    <section
      v-if="c.remark"
      class="mt-6"
    >
      <h2 class="text-sm text-muted">
        Anmerkungen
      </h2>
      <p class="whitespace-pre-line">
        {{ c.remark }}
      </p>
    </section>

    <!-- `commentary` is deliberately absent. The content type marks it
         "nicht für die Öffentlichkeit gedacht", and this is the public view of
         an entry — it is also why the search index does not carry it. -->
  </article>
</template>

<script setup lang="ts">
const props = defineProps<{
  entry: ArtworkEntry
}>()

const c = computed(() => props.entry.content)

// The entry's name is "Bild <titel>" as the import wrote it; the content's own
// title is the work's. Prefer the latter, and fall back for an untitled work.
const title = computed(() => c.value.title?.trim() || props.entry.name || 'Ohne Titel')

// The archive stores the Gattung as a single letter, which is what the index is
// faceted on. Spelled out here, since a reader is not looking at a facet.
const forms: Record<string, string> = { M: 'Malerei', Z: 'Zeichnung', P: 'Plastik' }

// Only what this work actually has: an empty row in a definition list reads as
// missing data rather than as a work that simply has no depth.
const facts = computed(() => {
  const f: Array<{ label: string, value: string }> = []
  const add = (label: string, value?: string | number) => {
    if (value !== undefined && value !== null && String(value).trim() !== '' && value !== 0) {
      f.push({ label, value: String(value) })
    }
  }

  add('Gattung', forms[c.value.form] ?? c.value.form)
  add('Technik', c.value.medium)
  add('Träger', c.value.support)
  add('Maße', dimensions.value)
  add('Fläche', c.value.area ? `${c.value.area.toFixed(2)} m²` : undefined)
  add('Phase', c.value.phase)
  add('Teile', c.value.teile > 1 ? c.value.teile : undefined)

  return f
})

// Height × width, and depth only where there is one — a painting has none, and
// printing "× 0" would say it does.
const dimensions = computed(() => {
  const { height, width, depth } = c.value
  if (!height && !width) return undefined

  return depth ? `${height} × ${width} × ${depth} cm` : `${height} × ${width} cm`
})
</script>
