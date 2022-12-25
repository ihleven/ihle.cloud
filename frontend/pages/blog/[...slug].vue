<template>
  <main class="px-4">
    <NavigationBar />
    <main class="prose">
      <!-- <ContentDoc /> -->
    </main>

    <article class="border-1 border-dotted border-gray-300">
      <ContentDoc
        class="prose max-w-none prose-img:rounded-xl prose-headings:no-underline prose-a:font-semibold prose-a:underline prose-a:underline-offset-4 prose-a:decoration-sky-300 hover:prose-a:decoration-2"
      />
    </article>
  </main>
</template>

<script setup>
const route = useRoute();

const path = computed(() => {
  return route.params.slug ? route.params.slug.join("/") : "";
});
</script>
