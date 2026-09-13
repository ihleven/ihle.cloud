<template>
  <div class="pb-32">
    <nav class="flex items-center justify-between border-b bg-white">
      <UButton
        variant="ghost" color="neutral" icon="i-heroicons-chevron-left"
        :disabled="spieltag <= 1"
        :to="`/geheimtipp/spieltag-${saison}-${spieltag - 1}`"
      />
      <h1 class="text-base">Spieltag {{ spieltag }}</h1>
      <UButton
        variant="ghost" color="neutral" icon="i-heroicons-chevron-right"
        :disabled="spieltag >= lastSpieltag"
        :to="`/geheimtipp/spieltag-${saison}-${spieltag + 1}`"
      />
    </nav>

    <div v-if="status === 'pending'" class="flex justify-center py-24">
      <UIcon name="i-heroicons-arrow-path" class="h-10 w-10 animate-spin text-blue-600" />
    </div>

    <div v-else-if="error || !round" class="flex flex-col items-center gap-4 py-24">
      <UAlert
        class="max-w-md" color="error" variant="soft"
        title="Dieser Spieltag ist gerade nicht erreichbar"
        :description="String(error ?? '')"
      />
      <NuxtLink to="/geheimtipp" class="text-sky-600 underline">zurück</NuxtLink>
    </div>

    <template v-else>
      <section class="overflow-x-auto py-8">
        <GhtSpieltag
          :round="round"
          :tipper="tipperByLogin"
          :login="login"
          :results="keepsResults(login)"
          @tipps="save"
          @results="saveResults"
        />
      </section>

      <UAlert
        v-if="resultState === 'error'"
        class="mx-auto max-w-screen-md" color="error" variant="soft"
        title="Die Ergebnisse konnten nicht gespeichert werden"
      />

      <UAlert
        v-if="Object.keys(refusals).length"
        class="mx-auto max-w-screen-md" color="error" variant="soft"
        title="Nicht alle Tipps wurden angenommen"
        :description="Object.values(refusals).join(' · ')"
      />

      <section class="flex flex-col items-center gap-8 bg-gray-100 md:flex-row md:items-start md:justify-center">
        <GhtTipperRanking :round="round" :tipper="tipperByLogin" />
        <GhtTabelle :round="round" />
      </section>
    </template>

    <GhtRegistrationDialog v-model:open="askToRegister" />
  </div>
</template>

<script setup lang="ts">
import type { CompRound } from '../../types'

definePageMeta({ public: true, layout: 'geheimtipp' })

const route = useRoute()
const saison = route.params.saison as string
const spieltag = Number(route.params.spieltag)

const { login, registration } = useGhtSession()
const { tipperByLogin, comp, load } = useGhtEdition()
const { refusals, place } = useGhtTipps()
const { state: resultState, save: putResults } = useGhtResults()

// The edition first: it carries the competition this matchday belongs to, and
// the names and avatars the matrix heads its columns with. Everything else in
// the screen — rows, both footer rows, the league table — arrives in the single
// round request below.
await load(saison)

const { data: round, status, error, refresh } = await useGht<CompRound>(
  () => `/comprounds/${comp.value}/${saison}/${spieltag}`,
)

const askToRegister = ref(false)

async function save(tipps: Record<string, string>) {
  // Registering is what earns a place in the edition; signing in alone does not.
  // Asked here rather than left to the backend, because the answer is a thing to
  // do and not an error.
  if (!registration.value) {
    askToRegister.value = true
    return
  }
  await place(comp.value, saison, spieltag, tipps).catch(() => {})
  // The save answers with the whole round, but re-reading keeps one path into
  // this page's state instead of two that have to agree.
  await refresh()
}

async function saveResults(results: Record<string, string>) {
  // Every tipper's points change when a result does, so the whole page is
  // re-read rather than the matrix patched: the two standings tables below it
  // are derived from the same payload.
  await putResults(round.value!, results).catch(() => {})
  await refresh()
}

// Bound from the season rather than assumed: the pool's own frontend hardcodes
// 34, which is right for the Bundesliga and wrong for everything else it has run.
const lastSpieltag = computed(() => round.value?.compSeason?.nrounds ?? 34)
</script>
