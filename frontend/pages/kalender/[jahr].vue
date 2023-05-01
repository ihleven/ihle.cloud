<template>
  <h1 class="bg-cyan-100">Jahr: {{ params.jahr }}</h1>
  <NuxtPage />
</template>
<script setup>
const { params } = useRoute();
</script>
