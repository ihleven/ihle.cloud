export default defineNuxtConfig({
  ssr: false,
  srcDir: 'frontend/',
  // app: {
  //     pageTransition: { name: 'page', mode: 'out-in' }
  //   },

  css: ['v-calendar/dist/style.css'],

  modules: [
    '@nuxt/content',
    '@nuxtjs/tailwindcss',
    '@nuxtjs/supabase',

    //    "@kevinmarrec/nuxt-pwa",
    'nuxt-icon', // das offizielle?
    'nuxt-icons', // https://github.com/gitFoxCode/nuxt-icons => <nuxt-icon name="mySuperIcon" filled />
  ],

  content: {
    documentDriven: false,
  },

  // pwa: {
  //   meta: {
  //     theme_color: '#e5e5e5',
  //   },
  //   manifest: {
  //     lang: 'de',
  //     name: 'Ihle.Cloud',
  //     short_name: 'iCloud',
  //     description: 'Die Ihle Cloud',
  //     display: 'standalone',
  //   },
  //   workbox: {
  //     enabled: false,
  //   },
  // },

  tailwindcss: {
    configPath: '@/tailwind.config.js',
    cssPath: '@/tailwind.css',
  },

  runtimeConfig: {
    apiBaseUrl: 'http://localhost:10815',
  },

  devtools: false,

  routeRules: {
    '/content-api/v1/**': { proxy: { to: 'http://localhost:8000/content-api/v1/**' } },
  },
})
