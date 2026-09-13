<template>
  <section class="mx-auto max-w-screen-lg px-4 py-8">
    <header class="flex items-baseline justify-between gap-4">
      <div>
        <h1 class="text-xl font-semibold">Avatare</h1>
        <p class="mt-1 text-sm text-gray-500">{{ avatare?.length ?? 0 }} Bilder</p>
      </div>

      <!-- The grid is for choosing a picture, the list for reading about one.
           Both answer a real question, so neither replaces the other. -->
      <div class="flex overflow-hidden rounded-md border border-gray-300">
        <button
          v-for="option in views" :key="option.value"
          class="flex items-center gap-1.5 px-3 py-1.5 text-sm"
          :class="view === option.value ? 'bg-gray-800 text-white' : 'bg-white text-gray-600 hover:bg-gray-100'"
          @click="view = option.value"
        >
          <UIcon :name="option.icon" class="h-4 w-4" />
          {{ option.label }}
        </button>
      </div>
    </header>

    <p v-if="status === 'pending'" class="mt-6 text-sm text-gray-500">Wird geladen …</p>

    <UAlert
      v-else-if="error"
      class="mt-6" color="error" variant="soft"
      title="Die Avatare sind gerade nicht erreichbar"
    />

    <ul v-else-if="view === 'grid'" class="mt-6 grid grid-cols-3 gap-4 sm:grid-cols-6">
      <li v-for="a in avatare" :key="a.id" class="text-center">
        <!-- Addressed by id, which is also what a tipper's `avatar` field is.
             The name is a filename and the backend does not route by it. -->
        <img
          :src="ghtAvatar(a.id, 96)"
          :alt="a.ght || a.name"
          class="aspect-square w-full rounded bg-gray-100 object-cover"
          loading="lazy"
        >
        <p class="mt-1 truncate text-xs text-gray-500">{{ a.ght || '—' }}</p>
      </li>
    </ul>

    <div v-else class="mt-6 overflow-x-auto">
      <table class="table-auto text-sm">
        <tbody>
          <tr
            v-for="a in avatare" :key="a.id"
            class="divide-y divide-gray-300 odd:bg-white even:bg-gray-200"
          >
            <td>
              <img
                class="h-16 w-16 max-w-none rounded-sm shadow-lg active:shadow"
                :src="ghtAvatar(a.id, 64)"
                :alt="a.ght || a.name"
                loading="lazy"
              >
            </td>
            <td class="px-2">{{ a.id }}</td>
            <td class="px-2">{{ a.account_id || '' }}</td>
            <td class="px-2">{{ a.ght }}</td>
            <td class="px-2">{{ a.name }}<br>{{ a.file }}</td>
            <td class="px-2 text-right">{{ a.size }}</td>
            <td class="px-2">{{ a.mimetype }}<br>{{ a.format }}, {{ a.mode }}</td>
            <td class="px-2 whitespace-nowrap">{{ a.w }}x{{ a.h }}</td>
            <td class="px-2">{{ a.labels.join(', ') }}</td>
            <td class="px-2">{{ a.topic }}</td>
            <td class="px-2">{{ a.public }}</td>
            <td class="px-2">{{ a.active }}</td>
            <td class="px-2 whitespace-nowrap">{{ ghtDate(a.created) }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </section>
</template>

<script setup lang="ts">
import type { AvatarFile } from '../../types'

// Moved under the pool's prefix: the old site had this at /avatare, and the
// root belongs to the family app.
definePageMeta({ public: true, layout: 'geheimtipp' })

const { data: avatare, status, error } = await useGht<AvatarFile[]>('/avatars')

// Kept for the session rather than in the URL: it is how someone prefers to
// look at the page, not what the page is about.
const view = useState<'grid' | 'list'>('ght:avatare:view', () => 'grid')

const views = [
  { value: 'grid' as const, label: 'Kacheln', icon: 'i-heroicons-squares-2x2' },
  { value: 'list' as const, label: 'Liste', icon: 'i-heroicons-bars-3' },
]
</script>
