<template>
  <section class="flex flex-col gap-6 border-t border-gray-200 pt-6">
    <div>
      <h2 class="text-base font-medium">Anmeldung</h2>
      <p class="mt-1 text-sm text-gray-600">
        {{ account.has_password ? 'Passwort gesetzt' : 'Noch kein Passwort' }} ·
        {{ account.passkeys === 1 ? '1 Passkey' : `${account.passkeys} Passkeys` }}
      </p>
    </div>

    <!-- The link is the way to onboard someone: they choose their own password
         as they register a device, so nothing has to be conveyed. -->
    <div>
      <UButton variant="subtle" :loading="linking" @click="issueLink">
        Einladungslink erzeugen
      </UButton>
      <div v-if="link" class="mt-3 rounded border border-gray-200 bg-gray-50 p-3">
        <p class="font-mono text-xs break-all">{{ link.url }}</p>
        <p class="mt-2 text-xs text-gray-500">
          Einmal gültig, bis {{ new Date(link.expires_at).toLocaleTimeString() }}.
          {{ account.has_password
            ? 'Das bestehende Passwort wird beim Registrieren abgefragt.'
            : 'Das Passwort wird dabei selbst gewählt.' }}
        </p>
      </div>
    </div>

    <form class="flex flex-wrap items-end gap-3" @submit.prevent="submitPassword()">
      <UFormField label="Passwort setzen" class="grow">
        <UInput v-model="password" type="password" autocomplete="new-password" class="w-full" />
      </UFormField>
      <UButton type="submit" variant="subtle" :loading="settingPassword">Setzen</UButton>
    </form>
    <p class="-mt-4 text-xs text-gray-500">
      Für den Notfall. Ein Passwort, das hier gesetzt wird, kennt die
      Administration — der Einladungslink ist der bessere Weg.
    </p>

    <!-- Advice, not a verdict: the server reports what it found and sets
         nothing, and using it anyway takes a second, deliberate click. -->
    <UAlert v-if="advice" color="warning" variant="soft">
      <template #description>
        <ul class="list-disc pl-4">
          <li v-if="advice.breaches">
            Dieses Passwort taucht {{ advice.breaches }}× in bekannten Datenlecks auf.
          </li>
          <li v-if="advice.too_short">
            {{ advice.length }} Zeichen; üblich sind 12 oder mehr.
          </li>
          <li v-if="advice.unchecked">
            Konnte nicht gegen bekannte Datenlecks geprüft werden.
          </li>
        </ul>
        <UButton class="mt-2" size="xs" color="warning" :loading="settingPassword" @click="submitPassword(true)">
          Trotzdem setzen
        </UButton>
      </template>
    </UAlert>

    <div v-if="keys.length">
      <h3 class="text-sm font-medium">Geräte</h3>
      <ul class="mt-2 divide-y divide-gray-200 border-y border-gray-200">
        <li v-for="key in keys" :key="key.id" class="flex items-center gap-4 py-2">
          <div class="grow">
            <p class="text-sm">{{ key.name || 'Unbenanntes Gerät' }}</p>
            <p class="text-xs text-gray-500">
              {{ new Date(key.created_at).toLocaleDateString() }} ·
              {{ key.last_used_at ? `zuletzt ${new Date(key.last_used_at).toLocaleDateString()}` : 'nie benutzt' }} ·
              {{ key.syncable ? 'geräteübergreifend' : 'nur dieses Gerät' }}
            </p>
          </div>
          <UButton color="error" variant="ghost" size="xs" :loading="removing === key.id" @click="remove(key.id)">
            Entfernen
          </UButton>
        </li>
      </ul>
    </div>

    <div class="flex flex-wrap gap-3">
      <UButton variant="ghost" size="xs" :loading="signingOut" @click="signOut">
        Überall abmelden
      </UButton>
      <UButton color="error" variant="ghost" size="xs" :loading="revoking" @click="revokeAll">
        Alle Passkeys entziehen
      </UButton>
    </div>

    <UAlert v-if="error" color="error" variant="soft" :description="error" />
    <UAlert v-if="note" color="success" variant="soft" :description="note" />
  </section>
</template>

<script setup lang="ts">
const props = defineProps<{ account: AdminAccount }>()
const emit = defineEmits<{ changed: [] }>()

const admin = useAdminAccounts()

const keys = ref<AdminPasskey[]>([])
const link = ref<Enrollment | null>(null)
const password = ref('')
const advice = ref<PasswordAdvice | null>(null)

const linking = ref(false)
const settingPassword = ref(false)
const removing = ref('')
const revoking = ref(false)
const signingOut = ref(false)
const error = ref('')
const note = ref('')

async function refresh() {
  try {
    keys.value = await admin.passkeys(props.account.name)
  }
  catch (e) {
    error.value = admin.describe(e)
  }
}

async function run(flag: Ref<boolean>, work: () => Promise<string>) {
  flag.value = true
  error.value = ''
  note.value = ''
  try {
    note.value = await work()
    emit('changed')
  }
  catch (e) {
    error.value = admin.describe(e)
  }
  finally {
    flag.value = false
  }
}

function issueLink() {
  return run(linking, async () => {
    link.value = await admin.enroll(props.account.name)
    return ''
  })
}

// confirm is the operator overriding the screening, which is why it is a second
// call rather than something sent optimistically the first time.
function submitPassword(confirm = false) {
  return run(settingPassword, async () => {
    const result = await admin.setPassword(props.account.name, password.value, confirm)
    if (!result.set) {
      advice.value = result.advice
      return ''
    }
    advice.value = null
    password.value = ''
    return 'Passwort gesetzt.'
  })
}

async function remove(id: string) {
  removing.value = id
  error.value = ''
  note.value = ''
  try {
    await admin.removePasskey(props.account.name, id)
    await refresh()
    note.value = 'Gerät entfernt.'
    emit('changed')
  }
  catch (e) {
    error.value = admin.describe(e)
  }
  finally {
    removing.value = ''
  }
}

function revokeAll() {
  return run(revoking, async () => {
    const { passkeys, sessions } = await admin.revoke(props.account.name)
    await refresh()
    return `${passkeys} Passkey(s) entzogen, ${sessions} Sitzung(en) beendet.`
  })
}

function signOut() {
  return run(signingOut, async () => {
    const { sessions } = await admin.signOutEverywhere(props.account.name)
    return `${sessions} Sitzung(en) beendet.`
  })
}

onMounted(refresh)
</script>
