<template>
  <main class="w-screen grid grid-cols-2">
    <section
      class="bg-ral-7035 col-span-2 aspect-[2/1] flex items-center justify-center"
    >
      <Logo />
    </section>

    <nuxt-link to="/videos" class="bg-blue-500 bg-opacity-90 p-8 aspect-square">
      <h1 class="text-white text-lg font-black hover:text-outline">Videos</h1>
    </nuxt-link>

    <section class="bg-green-500 bg-opacity-80 p-8 aspect-square">
      <!-- <dl class="text-white text-md font-extralight hover:font-normal">
        <nuxt-link :to="'/videos'" class="text-lg font-black hover:text-shadow"
          >Arbeit</nuxt-link
        >
      </dl> -->
    </section>
  </main>
</template>
