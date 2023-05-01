<template>
  <div class="bg-neutral-200 min-h-screen">
    <NavigationBar target="" class="sticky top-0 w-screen" />
    <arbeit-hero class="" />

    <!-- <nav class="bg-blue-200">
      <ul class="flex space-x-4">
        <li>
          <nuxt-link :to="{ name: 'kalender' }">arbeit</nuxt-link>
        </li>
        <li>
          <nuxt-link :to="{ name: 'kalender-jahr', params: { jahr: 2015 } }"
            >2015</nuxt-link
          >
        </li>
        <li>
          <nuxt-link :to="{ name: 'kalender-jahr', params: { jahr: 2016 } }"
            >2016</nuxt-link
          >
        </li>
        <li>
          <nuxt-link :to="{ name: 'kalender-jahr', params: { jahr: 2017 } }"
            >2017</nuxt-link
          >
        </li>
        <li>
          <nuxt-link :to="{ name: 'kalender-jahr', params: { jahr: 2018 } }"
            >2018</nuxt-link
          >
          <nuxt-link
            :to="{
              name: 'kalender-jahr-monat',
              params: { jahr: 2018, monat: 5 },
            }"
            >2018/05</nuxt-link
          >
        </li>
      </ul>
    </nav> -->

    <NuxtPage />
  </div>
</template>

<script>
// import { mapState } from 'vuex'

// import ArbeitHero from '@/components/arbeit/ArbeitHero.vue'

export default {
  // components: {
  //   ArbeitHero,
  // },
  // computed: {
  //   ...mapState('datum', ['jahr', 'monat', 'monatsname', 'tag', 'wochentag', 'kw']),
  // },
};
</script>
