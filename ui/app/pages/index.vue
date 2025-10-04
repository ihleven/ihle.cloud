<template>
  <div class="mx-auto my-4 w-screen max-w-screen-lg rounded bg-default p-4 shadow-md">

    INDEX
    <pre>{{ session }}</pre>

    <button @click="loadSession">Load session</button>
  </div>
</template>

<script setup lang="ts">
const { session, loadSession } = useAuth()
</script>
