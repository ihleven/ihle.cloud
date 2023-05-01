<template>
  <Dialog :open="open" class="relative z-50" @close="setIsOpen">
    <!-- The backdrop, rendered as a fixed sibling to the panel container -->
    <div class="fixed inset-0 bg-black/30" aria-hidden="true" />

    <!-- Full-screen scrollable container -->
    <div class="fixed inset-0 overflow-y-auto">
      <!-- Container to center the panel -->
      <div class="flex h-screen min-h-full items-stretch justify-center p-4">
        <!-- The actual dialog panel -->
        <DialogPanel class="flex w-full max-w-screen-md flex-col justify-stretch rounded bg-white">
          <DialogTitle>Complete your order</DialogTitle><button @click="close">close</button>
          <div class="flex h-full items-stretch p-1">
            <slot />
          </div>
        </DialogPanel>
      </div>
    </div>
  </Dialog>
</template>

<script setup>
  import { Dialog, DialogPanel, DialogTitle, DialogDescription } from '@headlessui/vue'

  defineProps({ open: Boolean })
  const emit = defineEmits(['close'])

  function close() {
    // isOpen.value = value
    emit('close')
  }
</script>
