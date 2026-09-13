<template>
  <main class="min-h-screen grow bg-gray-100">
    <h1 class="p-2 text-lg underline decoration-sky-300 decoration-dashed decoration-4 underline-offset-4">
      Geheimtipp
    </h1>

    <section class="bg-white/30 pt-8">
      <h3 class="px-4 py-1 font-semibold text-sky-300">aktuelle Spieltage</h3>

      <div v-if="rows.length" class="border-y border-gray-300">
        <GhtSpieltagRow
          v-for="row in rows" :key="row.label"
          :label="row.label"
          :round="row.round"
          :season="edition"
        />
      </div>

      <p v-else class="px-4 py-2 text-sm text-gray-500">
        Für diese Ausgabe sind gerade keine Spieltage angesetzt.
      </p>
    </section>
  </main>
</template>

<script setup lang="ts">
// Public to this app: the pool's own gate is the layer's middleware, and the
// family app's sign-in overlay has no business covering this.
definePageMeta({ public: true, layout: 'geheimtipp' })

// Already loaded by the gate — /aktuell is both the session probe and this
// page's content, which is why it is one request and not two.
const { aktuell, edition } = useGhtSession()

const rows = computed(() => {
  const s = aktuell.value?.spieltage
  if (!s) return []
  return [
    { label: 'Letzter Spieltag', round: s.last },
    { label: 'Aktueller Spieltag', round: s.current },
    { label: 'Nächster Spieltag', round: s.next },
  ].filter(r => r.round).map(r => ({ ...r, round: r.round!, season: edition.value }))
})
</script>
