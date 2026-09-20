<template>
  <article>
    <header class="mb-4">
      <h2 class="text-xl font-bold text-highlighted">
        {{ reise.ziel }}
      </h2>
      <p
        v-if="dates"
        class="text-sm text-muted"
      >
        {{ dates }}
      </p>
    </header>

    <!-- The body is markdown, and it carries components of its own — the daily
         expenses card, for one — which is why it goes through MDC rather than
         being rendered as prose. Those components live in components/content
         and are registered globally, because MDC resolves them by name at
         runtime and would not find a lazily-loaded one. -->
    <MDC
      v-if="reise.body"
      :value="reise.body"
      tag="div"
      class="prose prose-sm max-w-none"
    />
  </article>
</template>

<script setup lang="ts">
const props = defineProps<{
  reise: Reise
}>()

// Both ends where there are both, otherwise whichever is known — a journey
// still being written has a start and no end, and "1.8. – " reads as broken
// rather than as ongoing.
const dates = computed(() => {
  const from = format(props.reise.von)
  const to = format(props.reise.bis)
  if (from && to) return `${from} – ${to}`

  return from || to || (props.reise.jahr ? String(props.reise.jahr) : '')
})

function format(value?: string) {
  if (!value) return ''
  const d = new Date(value)

  return Number.isNaN(d.getTime()) ? value : d.toLocaleDateString('de-DE')
}
</script>
