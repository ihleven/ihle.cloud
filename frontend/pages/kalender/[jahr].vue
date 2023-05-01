<template>
  <div>
    <h1 class="bg-cyan-100">Jahr: {{ params.jahr }}</h1>
    <NuxtPage />
  </div>
</template>
<script setup>
  const { params } = useRoute()
</script>
