<template>
  <main class="bg-cyan-100 min-h-screen">
    <h1>Tag: {{ params.tag }}</h1>
  </main>
</template>
<script setup>
const { params } = useRoute();
</script>
