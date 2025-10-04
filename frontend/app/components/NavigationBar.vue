<template>
  <header class="flex items-center justify-start bg-transparent backdrop-blur">
    <button v-if="target" class="flex items-center justify-start py-2 focus:outline-none" @click="goback">
      <svg
        viewBox="0 0 24 24"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
        class="h-8 w-8 fill-none stroke-current stroke-2"
      >
        <polyline points="15 18 9 12 15 6"></polyline>
      </svg>
      <span class="text-md -m-1 font-medium tracking-tighter">{{ target }}</span>
    </button>

    <Logo v-else />
  </header>
</template>

<script setup>
  const props = defineProps({
    title: String,
    target: String,
  })

  function goback() {
    if (props.target) {
      useRouter().push({ path: '/' + props.target })
    } else {
      history.back()
    }
  }
</script>

<style lang="css" scoped>
  .kreis {
    width: 180vh;
    height: 180vh;
  }
  .navigation {
    position: absolute;
    top: 0;
    bottom: 0;
    left: 0;
    right: 0;
    color: white;
    background: black;
    width: 100vw;
    height: 100vh;
    text-align: left;
    opacity: 0;
    pointer-events: none;
    /* z-index: 1000; */
    line-height: 1;
    transition: all 0.3s ease-in-out 0s;
    overflow-y: scroll;
  }

  .navigation.open {
    opacity: 1;
    pointer-events: auto;
  }
</style>
