export default defineNuxtConfig({
  modules: [
    '@nuxt/ui',
    // '@nuxtjs/mdc',
    '@nuxt/eslint',
    '@nuxt/content',
  ],
  ssr: false,
  // components: [
  //   { path: '~/components/prose', global: true },
  //   { path: '~/components/content', global: true },
  //   { path: '~/components', pathPrefix: false },
  // ],
  devtools: { enabled: false },

  app: {
    rootAttrs: { class: 'h-full' },
    head: {
      htmlAttrs: {
        class: 'h-full',
      },
      bodyAttrs: {
        class: 'h-full',
      },
    },
  },

  css: ['~/assets/css/main.css'],

  // css: [
  //   // 'v-calendar/dist/style.css',
  //   '@/assets/iawriter.css',
  //   '@fontsource/inter/variable.css',
  //   '@fontsource/raleway/variable.css',
  // ],
  // routeRules: { '/hi/**': { proxy: { to: 'http://localhost:8000/hi/**' } } },
  runtimeConfig: {
    apiBaseUrl: 'http://localhost:10815',

    public: {
      hidrive: {
        api: 'http://localhost:8000/hi',
      },
    },

    cookie: {
      key: '',
      options: {
        // maxAge: 30,
        // expires: new Date(response.expires * 1000),
        sameSite: 'lax', // 'strict' prevents direct auth in case of new single sign login
        httpOnly: true,
        secure: process.env.NODE_ENV !== 'development',
      },
    },
  },

  // srcDir: 'frontend/',

  // ssr: false,
  // devServer: {
  //   port: 3000,
  // },

  // compatibilityDate: '2024-12-17',

  // content: {
  //   documentDriven: false,
  // },

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

  // tailwindcss: {
  //   exposeConfig: true,
  //   viewer: true,
  //   configPath: '@/tailwind.config.js',
  //   cssPath: '@/tailwind.css',
  // },

  eslint: {
    config: {
      stylistic: true,
      nuxt: {
        sortConfigKeys: true,
      },
    },
  },
})
