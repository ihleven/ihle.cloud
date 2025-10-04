<template>
  <div>
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
    <slot></slot>

    <footer class="h-96 w-full bg-gray-500"></footer>
  </div>
</template>

<style>
  .page-left-enter-active,
  .page-right-enter-active,
  .page-left-leave-active,
  .page-right-leave-active {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    transition: all 0.2s linear;
  }

  .page-left-enter-from,
  .page-right-leave-to {
    transform: translateX(100%);
  }
  .page-left-leave-to,
  .page-right-enter-from {
    transform: translateX(-100%);
  }
  .page-left-enter-to,
  .page-right-enter-to {
    transform: translateX(0);
  }
</style>
