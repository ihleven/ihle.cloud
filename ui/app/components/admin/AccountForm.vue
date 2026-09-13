<template>
  <form class="flex flex-col gap-4" @submit.prevent="save">
    <div class="flex flex-wrap gap-3">
      <UFormField label="Name" class="grow">
        <UInput v-model="edit.display_name" class="w-full" />
      </UFormField>
      <UFormField label="E-Mail" class="grow">
        <UInput v-model="edit.email" type="email" class="w-full" />
      </UFormField>
    </div>
    <p class="-mt-2 text-xs text-gray-500">
      Anmeldename <code>{{ account.name }}</code> — nicht änderbar, Inhalte gehören ihm.
    </p>

    <fieldset>
      <legend class="text-sm font-medium">Bereiche</legend>
      <p class="mt-1 mb-2 text-xs text-gray-500">
        Was dieses Konto zu sehen bekommt. <code>*</code> heißt alles, auch
        Bereiche, die es noch nicht gibt.
      </p>

      <div class="flex flex-wrap gap-x-6 gap-y-2">
        <UCheckbox v-model="all" label="Alles (*)" />
        <UCheckbox
          v-for="area in areas"
          :key="area"
          :model-value="all || granted.has(area)"
          :disabled="all"
          :label="area"
          @update:model-value="toggle(area, $event)"
        />
      </div>
    </fieldset>

    <!-- Stored, looks like a grant, grants nothing: a name this build does not
         define is dropped when the scope is built. Worth saying out loud. -->
    <UAlert
      v-if="account.unregistered_permissions.length"
      color="warning" variant="soft"
      :description="`Ohne Wirkung in dieser Version: ${account.unregistered_permissions.join(', ')}`"
    />

    <UCheckbox v-model="edit.disabled" label="Gesperrt" />
    <p class="-mt-2 text-xs text-gray-500">
      Ein gesperrtes Konto kann sich nicht anmelden und verliert sofort alle
      offenen Sitzungen. Konten werden nicht gelöscht.
    </p>

    <div class="flex items-center gap-3">
      <UButton type="submit" :loading="busy">Speichern</UButton>
      <span v-if="saved" class="text-sm text-gray-500">Gespeichert.</span>
    </div>

    <UAlert v-if="error" color="error" variant="soft" :description="error" />
  </form>
</template>

<script setup lang="ts">
const props = defineProps<{ account: AdminAccount }>()
const emit = defineEmits<{ saved: [] }>()

const admin = useAdminAccounts()

const areas = ref<string[]>([])
const busy = ref(false)
const saved = ref(false)
const error = ref('')

const edit = reactive({
  display_name: props.account.display_name,
  email: props.account.email,
  disabled: props.account.disabled,
})

// An area is granted by a module.<id> permission; the wildcard is stored as a
// permission of its own and stands for all of them. Both are held here as what
// they are, so saving writes back the same vocabulary the server reads.
const WILDCARD = '*'
const prefix = 'module.'

const all = ref(props.account.permissions.includes(WILDCARD))
const granted = reactive(new Set(
  props.account.permissions
    .filter(p => p.startsWith(prefix))
    .map(p => p.slice(prefix.length)),
))

// Permissions that are neither the wildcard nor an area are left untouched:
// this form is about areas, and silently dropping something it does not
// understand would be a quiet way to take a right away.
const others = props.account.permissions.filter(
  p => p !== WILDCARD && !p.startsWith(prefix),
)

function toggle(area: string, on: unknown) {
  if (on) granted.add(area)
  else granted.delete(area)
}

async function save() {
  busy.value = true
  saved.value = false
  error.value = ''
  try {
    const permissions = all.value
      ? [WILDCARD, ...others]
      : [...[...granted].map(a => prefix + a), ...others]

    await admin.update(props.account.name, {
      display_name: edit.display_name,
      email: edit.email,
      groups: props.account.groups,
      permissions,
      disabled: edit.disabled,
    })
    saved.value = true
    emit('saved')
  }
  catch (e) {
    error.value = admin.describe(e)
  }
  finally {
    busy.value = false
  }
}

onMounted(async () => {
  try {
    areas.value = await admin.areas()
  }
  catch (e) {
    error.value = admin.describe(e)
  }
})
</script>
