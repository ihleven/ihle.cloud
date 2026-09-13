<template>
  <section class="grow bg-slate-100 px-2 py-16">
    <p v-if="!ausgabe" class="text-sm text-gray-500">Wird geladen …</p>

    <ul v-else class="mx-auto max-w-screen-md">
      <li
        v-for="t in ausgabe.tipper" :key="t.id"
        class="mx-auto mt-4 flex max-w-sm items-start space-x-4 rounded-lg bg-white p-3 shadow-md"
      >
        <div class="shrink-0">
          <img class="h-20 w-20 rounded shadow" :src="ghtAvatar(t.avatar, 96)" :alt="t.name">
        </div>

        <div class="w-full">
          <div class="text-base leading-none font-medium text-black">
            {{ t.name }}
            <small class="text-sm text-gray-500">({{ t.ght }})</small>
          </div>
          <div class="text-sm font-light text-gray-600">&bdquo;{{ t.motto }}&ldquo;</div>

          <!-- One line per competition of the edition. Empty for everyone but
               yourself: the pool only fills this in for the caller it knows. -->
          <p
            v-for="(tn, comp) in t.teilnahmen" :key="comp"
            class="mt-2 flex w-full justify-between text-sm text-gray-500"
          >
            <span>{{ tn.platz }}.</span>
            <span>{{ tn.tipps }} Tipps</span>
            <b>{{ tn.punkte }} Pkt.</b>
          </p>

          <div class="text-right">{{ t.platz }}. {{ t.pkt }} Pkt.</div>
        </div>
      </li>
    </ul>
  </section>
</template>

<script setup lang="ts">
definePageMeta({ public: true, layout: 'geheimtipp' })

const { edition } = useGhtSession()
const { ausgabe, load } = useGhtEdition()

await load(edition.value)
</script>
