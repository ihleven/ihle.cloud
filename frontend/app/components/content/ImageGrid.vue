<template>
  <figure>
    <ContainerScreenWidth><slot /></ContainerScreenWidth>

    <figcaption>
      <slot name="caption" />
      <!-- <ContentSlot :use="$slots.caption" unwrap="p">
        <slot></slot>
      </ContentSlot> -->
    </figcaption>
  </figure>
</template>

<script setup>
defineProps(['caption'])
</script>
