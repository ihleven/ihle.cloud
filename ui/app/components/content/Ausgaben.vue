<template>
  <div class="w-64 rounded bg-elevated p-1 shadow">
    <h4 class="pt-2 pl-4 text-sm font-bold">
      Tagesausgaben:
    </h4>
    <slot />
  </div>
</template>

<script setup lang="ts">
// Used from inside a journey's markdown as `::ausgaben`, so it is found by name
// at runtime — which is why components/content is registered global in
// nuxt.config, like the entry layer's content-type editors.
</script>
