<template>
  <figure>
    <ContainerScreenWidth><slot /></ContainerScreenWidth>

    <figcaption>
      <slot name="namedSlot" />
      <!-- <ContentSlot :use="$slots.caption" unwrap="p">
        <slot></slot>
      </ContentSlot> -->
    </figcaption>
  </figure>
</template>

<script setup>
  defineProps(['caption'])
</script>
