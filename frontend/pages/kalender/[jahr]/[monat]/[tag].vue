<template>
  <main class="min-h-screen bg-cyan-100">
    <h1>Tag: {{ params.tag }}</h1>
  </main>
</template>
<script setup>
  const { params } = useRoute()
</script>
