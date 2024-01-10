<!-- <template>
  <span :class="{ 'fill-current': !filled }" v-html="icon" />
</template> -->

<template>
  <span
    class="micon block"
    :class="{ 'nuxt-icon--fill': !filled, 'nuxt-icon--stroke': hasStroke && !filled }"
    v-html="icon"
  />
</template>

<script setup lang="ts">
  import { ref, watchEffect } from '#imports'

  const props = withDefaults(
    defineProps<{
      name: string
      filled?: boolean
    }>(),
    { filled: false }
  )

  const icon = ref<string | Record<string, any>>('')
  // let hasStroke = false

  async function getIcon() {
    try {
      const iconsImport = import.meta.glob('assets/icons/**/**.svg', {
        as: 'raw',
        eager: false,
      })
      const rawIcon = await iconsImport[`/assets/icons/${props.name}.svg`]()
      // if (rawIcon.includes('stroke')) { hasStroke = true }
      icon.value = rawIcon
    } catch {
      console.error(`[nuxt-icons] Icon '${props.name}' doesn't exist in 'assets/icons'`)
    }
  }

  await getIcon()

  watchEffect(getIcon)
</script>

<style>
  .micon svg {
    width: 100%;
    height: 100%;
    /*  margin-bottom: 0.125em;
  vertical-align: middle; */
  }
  .micon.nuxt-icon--fill,
  .micon.nuxt-icon--fill * {
    fill: currentColor !important;
  }

  .micon--stroke,
  .micon.nuxt-icon--stroke * {
    stroke: currentColor !important;
  }
</style>
