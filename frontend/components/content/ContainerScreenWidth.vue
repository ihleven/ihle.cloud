<template>
  <div class="container-margin w-screen">
    <slot />
  </div>
</template>

<style scoped>
  .container-margin {
    margin-left: calc((1rem) / 2 * -1);
  }

  @screen sm {
    .container-margin {
      margin-left: calc((100vw - max(theme('screens.sm'), 100vw -2rem)) / 2 * -1);
    }
  }

  @screen md {
    .container-margin {
      margin-left: calc((100vw - min(theme('screens.md'), 100vw - 2rem)) / 2 * -1);
    }
  }
</style>
