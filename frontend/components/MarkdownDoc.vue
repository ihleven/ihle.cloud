<template>
  <article>
    <nuxt-icon name="chevron-right" filled />
    <header class="w-full bg-violet-200">
      <p class="mb-2 text-sm font-semibold leading-6 text-sky-500">{{ doc.title }}</p>
      <h1 class="inline-block text-2xl font-extrabold tracking-tight text-slate-900 sm:text-3xl">{{ doc?.title }}</h1>
      <p class="mt-2 text-lg text-slate-700 dark:text-slate-400">{{ doc?.description }}</p>
    </header>
    <main class="container-fluid-md font-raleway">
      <ContentRenderer :value="doc" class="prose max-w-none prose-p:text-justify" />
    </main>

    <footer class="w-full bg-violet-200">footer</footer>
  </article>
</template>

<script setup>
  defineProps({ doc: Object })
</script>
