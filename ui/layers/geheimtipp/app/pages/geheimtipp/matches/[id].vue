<template>
  <main class="bg-gray-100 pb-16">
    <p v-if="status === 'pending'" class="p-8 text-center text-sm text-gray-500">Wird geladen …</p>

    <UAlert
      v-else-if="error || !match"
      class="m-8" color="error" variant="soft"
      title="Dieses Spiel ist gerade nicht erreichbar"
    />

    <template v-else>
      <GhtMatchHeader :match="match" />

      <section v-if="open" class="py-16">
        <GhtTippForm
          :tipp="myTipp"
          :state="state"
          :refusal="refusals[String(match.id)]"
          @save="save"
        />
      </section>

      <!-- Past the kick-off there is nothing to offer, and saying so is better
           than a form that refuses on submit. -->
      <p v-else-if="signedIn" class="py-16 text-center text-sm text-gray-500">
        Für dieses Spiel kann nicht mehr getippt werden.
      </p>

      <section>
        <h4 class="mx-auto max-w-screen-md px-4 py-2 text-xl font-medium">Alle Tipps:</h4>
        <GhtTippList :tipps="match.odds?.tipps ?? null" :tipper="tipperByLogin" />
      </section>
    </template>

    <GhtRegistrationDialog v-model:open="askToRegister" />
  </main>
</template>

<script setup lang="ts">
import type { MatchDetail } from '../../../types'

definePageMeta({ public: true, layout: 'geheimtipp' })

const id = useRoute().params.id as string

const { login, signedIn, edition, registration } = useGhtSession()
const { tipperByLogin, load } = useGhtEdition()
const { state, refusals, place } = useGhtTipps()

await load(edition.value)

const { data: match, status, error, refresh } = await useGht<MatchDetail>(() => `/matches/${id}`)

const myTipp = computed(() => (login.value && match.value?.odds?.tipps?.[login.value]) || null)

/** Whether to offer the form at all: signed in, and before kick-off. */
const open = computed(() => !!match.value && mayTip(login.value, login.value, match.value.kickoff))

const askToRegister = ref(false)

async function save(ergebnis: string) {
  // Registering is what earns a place in the edition; signing in alone does not.
  // Checked here rather than by letting the backend refuse, because the answer
  // is a thing to do, not an error.
  if (!registration.value) {
    askToRegister.value = true
    return
  }
  const m = match.value!
  await place(m.comp, m.season, m.round.num, { [String(m.id)]: ergebnis }).catch(() => {})
  // The round that comes back is the whole matchday; this page shows one match
  // of it, so it re-reads its own rather than picking the match out.
  await refresh()
}
</script>
