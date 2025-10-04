<template>
  <h1 class="bg-red-100">Monat: {{ params.monat }}</h1>
  <NuxtPage />
</template>
<script setup>
  const { params } = useRoute()
</script>
