<template>
  <div class="grid grid-cols-2 gap-4">
    <div class="ring-sky-300 hover:ring-0">
      <slot mdc-unwrap="p" />
    </div>
    <aside class="ring-sky-300 hover:ring-0">
      <slot name="aside" />
    </aside>
  </div>
</template>

<script setup></script>
