<template>
  <main class="mx-auto max-w-sm px-4 py-8">
    <h1 class="text-lg font-semibold">{{ registration ? 'Meine Registrierung' : 'Registrieren' }}</h1>
    <p class="mt-1 text-sm text-gray-500">
      Name, Motto und Avatar für {{ edition }}. Ohne Registrierung kann nicht getippt werden.
    </p>

    <label class="mt-6 block text-base font-semibold text-gray-700">Avatar:</label>
    <button
      type="button"
      class="block w-full rounded border border-sky-300 bg-white p-1 focus:border-sky-500 focus:ring-2 focus:ring-sky-500"
      @click="picking = true"
    >
      <img class="mx-auto rounded" :src="ghtAvatar(form.avatar, 64)" :alt="String(form.avatar)">
    </button>

    <label for="name" class="mt-6 block text-base font-semibold text-gray-700">Name:</label>
    <input
      id="name" v-model="form.name" name="name"
      type="text" :placeholder="login"
      class="mt-1 block w-full rounded border border-sky-300 px-3 py-2 focus:border-sky-500 focus:ring-sky-500"
    >

    <label for="motto" class="mt-6 block text-base font-semibold text-gray-700">Motto:</label>
    <textarea
      id="motto" v-model="form.motto" rows="3"
      placeholder="Dein Motto für die Tipprunde"
      class="mt-1 block w-full rounded border border-sky-300 px-3 py-2 focus:border-sky-500 focus:ring-sky-500"
    />

    <button
      :disabled="!dirty || busy"
      class="mt-8 flex w-full items-center justify-center gap-2 rounded border border-sky-300 bg-white px-4 py-2 text-sky-400 enabled:hover:bg-sky-300 enabled:hover:text-zinc-50 disabled:bg-gray-100 disabled:text-sky-200"
      @click="submit"
    >
      {{ registration ? 'Speichern' : 'Registrieren' }}
      <UIcon v-if="busy" name="i-heroicons-arrow-path" class="h-5 w-5 animate-spin" />
    </button>

    <UAlert v-if="saved" class="mt-6" color="success" variant="soft" title="Gespeichert" />
    <UAlert v-if="failed" class="mt-6" color="error" variant="soft" title="Das hat nicht geklappt" />

    <UModal v-model:open="picking" title="Avatar wählen" description="Ein Bild für die Tipprunde">
      <template #content>
        <div class="max-h-[70vh] overflow-y-auto p-4">
          <input type="file" multiple class="mb-4 block w-full text-sm" @change="onFileSelect">
          <p class="mb-2 text-sm text-gray-500">{{ avatars?.length ?? 0 }} Avatare</p>
          <div class="grid gap-2" style="grid-template-columns: repeat(auto-fill, minmax(96px, 1fr))">
            <button v-for="a in avatars" :key="a.id" type="button" @click="choose(a.id)">
              <img class="h-24 w-24 rounded shadow" :src="ghtAvatar(a.id, 96)" :alt="String(a.id)" loading="lazy">
            </button>
          </div>
        </div>
      </template>
    </UModal>
  </main>
</template>

<script setup lang="ts">
import type { AvatarFile, EditionTipper } from '../../types'

definePageMeta({ public: true, layout: 'geheimtipp' })

const { login, edition, registration, load: reloadSession } = useGhtSession()
const { load: reloadEdition } = useGhtEdition()

// 148 is the pool's own stand-in picture, which is what an unregistered form
// starts from — the same default its site uses.
const DEFAULT_AVATAR = 148

const form = reactive({
  name: registration.value?.name ?? '',
  motto: registration.value?.motto ?? '',
  avatar: registration.value?.avatar ?? DEFAULT_AVATAR,
})

const dirty = computed(() =>
  form.name !== (registration.value?.name ?? '')
  || form.motto !== (registration.value?.motto ?? '')
  || form.avatar !== (registration.value?.avatar ?? DEFAULT_AVATAR),
)

const picking = ref(false)
const busy = ref(false)
const saved = ref(false)
const failed = ref(false)

const { data: avatars, refresh: reloadAvatars } = await useGht<AvatarFile[]>('/avatars')

function choose(id: number) {
  form.avatar = id
  picking.value = false
}

async function onFileSelect(event: Event) {
  const files = (event.target as HTMLInputElement).files
  for (const file of files ?? []) {
    const body = new FormData()
    body.append('name', file.name)
    body.append('modtime', String(file.lastModified / 1000))
    body.append('file', file, file.name)
    await ghtFetch('/avatars', { method: 'POST', body }).catch(() => {})
  }
  await reloadAvatars()
}

async function submit() {
  busy.value = true
  saved.value = false
  failed.value = false
  try {
    await ghtFetch<EditionTipper>('/registrations', {
      method: 'PUT',
      body: new URLSearchParams({ name: form.name, motto: form.motto, avatar: String(form.avatar) }),
    })
    // Both copies, because the same person is described twice: /aktuell carries
    // the registration the menus read, and the edition carries the tipper the
    // ranking and the matrix headers read. Refreshing only the first is why a
    // changed motto used to need a page reload before it showed up.
    await reloadSession(true)
    await reloadEdition(edition.value, true)
    saved.value = true
  }
  catch {
    failed.value = true
  }
  finally {
    busy.value = false
  }
}
</script>
