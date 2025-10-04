// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({

  modules: [
    '@nuxt/eslint',
    '@nuxt/ui',
    '@nuxtjs/mdc',
    '@vueuse/nuxt',
  ],

  ssr: false,
  components: [
    // { path: '~/components/prose', global: true },
    // { path: '~/components/content', global: true },
    { path: '~/components', pathPrefix: false },
  ],

  devtools: { enabled: true },
  // ssr: false,
  // app: {
  //   rootAttrs: { class: 'h-full' },
  //   head: {
  //     htmlAttrs: {
  //       class: 'h-full',
  //     },
  //     bodyAttrs: {
  //       class: 'h-full',
  //     },
  //   },
  // },
  css: ['~/assets/css/main.css'],

  runtimeConfig: {
    public: {
      api: {
        origin: '',
        base: '',
      },
    },
  },

  compatibilityDate: '2025-07-15',

  eslint: {
    config: {
      stylistic: true,
      nuxt: {
        sortConfigKeys: true,
      },
    },
  },
})
