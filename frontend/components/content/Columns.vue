<template>
  <div class="grid grid-cols-2 gap-4 ring-0 ring-sky-300 ring-offset-4 hover:ring-2">
    <div class="ring-sky-300 hover:ring-2">
      <slot />
    </div>
    <aside class="ring-sky-300 hover:ring-2">
      <slot name="aside" />
    </aside>
  </div>
</template>

<script setup></script>
