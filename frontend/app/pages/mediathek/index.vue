<template>
  <main class="grid w-screen grid-cols-2">
    <NavigationBar /><i />
    <nuxt-link
      v-for="m in medien"
      :key="m.dir"
      :to="`/mediathek/${m.dir}`"
      :style="{ 'background-image': `url(${media('/public/mediathek/' + m.dir + '/cover.jpg')})` }"
      class="relative flex aspect-square h-full w-full items-center justify-center overflow-hidden bg-sky-300 bg-cover bg-top bg-no-repeat p-8 text-lg font-black text-white hover:text-outline"
    >
      <!-- <img :src="media('/public/mediathek/' + m.dir + '/cover.jpg')" class="absolute inset-0" /> -->
      <!-- <h4>{{ m.name || m.dir }}</h4> -->
    </nuxt-link>

    <!-- <section class="aspect-square bg-green-500 bg-opacity-90">
      <nuxt-link v-if="dev" to="/kalender" class="block h-full w-full p-8">
        <h1 class="hover:text-outline text-lg font-black text-white">Kalender</h1>
      </nuxt-link>
    </section> -->

    <!-- <section class="aspect-square bg-sky-500 bg-opacity-80">
      <nuxt-link
        v-if="dev"
        to="/mediathek"
        class="hover:text-outline block h-full w-full p-8 text-lg font-extralight text-white hover:font-normal"
        >Mediathek
      </nuxt-link>

    </section> -->

    <!-- <section class="aspect-square bg-cyan-300 bg-opacity-80"></section> -->
  </main>
</template>

<script setup lang="ts">
import { useCookie } from 'nuxt/app'

definePageMeta({
  layout: 'token',
  middleware: [
    function (to) {
      console.log('to', to, import.meta.client)
      if (import.meta.server) {
        const route = useRoute()
        const token = route.query.token

        const s = useState('token', () => token)
        // const foo = useCookie('foo')
        console.log('foo', s)
      }
      console.log('custom middleware', to.path)
      if (import.meta.client) {
        const token = useState('token')

        console.log('token client', token.value)
      }
    },
  ],
})

const { media } = useHidrive()
const medien = [
  { dir: 'captain-future' },
  // {
  //   name: 'Die Erben der Saurier',
  //   dir: 'Die-Erben-der-Saurier',
  //   link: '/mediathek/Die-Erben-der-Saurier/1.1-Eine-neue-Zeit',
  //   img: '/public/mediathek/Die-Erben-der-Saurier/cover.jpg',
  //   class: "bg-[url('http://localhost:8000/hi/media/public/mediathek/Die-Erben-der-Saurier/cover.jpg')]",
  // },
  { dir: '2001-walking-with-beasts', titel: 'Die Erben der Saurier', orig: 'Walking with Beasts', jahr: 2001 },
  {
    dir: '2003-walking-with-cavemen',
    titel: 'Im Reich der Urmenschen',
    orig: 'Walking with Cavemen',
    jahr: 2003,
  },
  {
    dir: '2003-sea-monsters',
  },
  {
    dir: '2003-monsters-we-met', titel: 'Menschen gegen Monster', orig: 'Monsters We Met', jahr: 2003,
  },
  {
    dir: '2005-walking-with-monsters', titel: 'Die Ahnen der Saurier', orig: 'Walking with Monsters', jahr: 2005,
  },
  {
    dir: '2011-planet-of-the-apemen',
    titel: 'Kampf der Menschenaffen',
    orig: 'Planet of the Apemen: Battle for Earth',
    jahr: 2011,
  },
  { dir: '2013-ice-age-giants', orig: 'Ice Age Giants', jahr: 2013 },
]
</script>
