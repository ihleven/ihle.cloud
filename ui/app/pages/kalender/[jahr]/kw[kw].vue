<template>
  <h1 class="bg-orange-100 px-2 py-1 text-sm text-neutral-700">KW: {{ route.params.kw }}</h1>
</template>

<script setup lang="ts">
// An ISO week, beside the months rather than inside one: a week can straddle
// two months and two years, which is why it hangs off the year and not off a
// month.
const route = useRoute()
</script>
