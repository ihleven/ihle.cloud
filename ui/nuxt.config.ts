// https://nuxt.com/docs/api/configuration/nuxt-config

// Where `bun run dev` forwards API calls. Only used in development.
const devAPI = process.env.NUXT_DEV_API_TARGET || 'http://localhost:8000'
export default defineNuxtConfig({

  // The entry-editing shell comes from the CMS rather than being copied: the
  // metadata panel, the directory browser, the fallback editor and the
  // content-type registry. This app contributes its own editors through
  // app.config.contentTypes; the layer never needs to know about them.
  //
  // A relative path while the layer is still moving, mirroring how go.mod
  // already depends on the same checkout. Switching to a git reference later is
  // a one-line change.
  extends: ['../../cms/ui/layers/entry'],

  modules: [
    '@nuxt/eslint',
    '@nuxt/ui',
    '@nuxtjs/mdc',
    '@vueuse/nuxt',
  ],

  ssr: false,
  components: [
    // Content-type editors are resolved by name at runtime by the entry layer, so
    // they have to be global — a lazily-loaded component would not be found.
    { path: '~/components/cms/contenttypes', pathPrefix: false, global: true },
    { path: '~/components/cms/panels', pathPrefix: false, global: true },
    { path: '~/components', pathPrefix: false },
  ],

  devtools: { enabled: true },

  // Added to the home screen on an iPhone, the app opens without Safari's
  // address bar and toolbar. This is a manifest and a few meta tags and nothing
  // else — deliberately no service worker: iOS does not need one for standalone
  // mode, and a precaching worker is what makes a deployed update arrive late.
  app: {
    head: {
      link: [
        { rel: 'manifest', href: '/manifest.webmanifest' },
        { rel: 'apple-touch-icon', href: '/apple-touch-icon.png' },
      ],
      meta: [
        // The modern name, and the one iOS has always read.
        { name: 'mobile-web-app-capable', content: 'yes' },
        { name: 'apple-mobile-web-app-capable', content: 'yes' },
        { name: 'apple-mobile-web-app-title', content: 'ihleven' },
        // "default" keeps the status bar its own strip rather than putting the
        // page under it, so no safe-area handling is needed.
        { name: 'apple-mobile-web-app-status-bar-style', content: 'default' },
        { name: 'theme-color', content: '#ffffff' },
      ],
    },
  },
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
  ui: {
    colorMode: false,
  },

  runtimeConfig: {
    public: {
      api: {
        origin: '',
        // Relative on purpose: the app calls whatever origin served it. An
        // absolute URL here would be baked into the build, so the app would call
        // one host while being served from another — which is cross-origin, and
        // then cookies are withheld and CORS has to be configured to match.
        //
        // In development the API is genuinely on another port; devProxy below
        // bridges that instead, so requests stay same-origin there too.
        base: '',
      },
      // What the entry layer reads. Relative like everything else here, so the
      // app calls whichever origin served it, with this app's own API prefix.
      apiBaseURL: '/api/v1',
    },
  },

  compatibilityDate: '2025-07-15',

  // Development only: Nuxt serves the app on :3000 while the Go server holds the
  // API. Proxying keeps the browser talking to one origin, which is what makes
  // session and passkey cookies behave the same in development as in production.
  nitro: {
    devProxy: {
      '/auth': { target: devAPI + '/auth', changeOrigin: true },
      '/api': { target: devAPI + '/api', changeOrigin: true },
      '/hi': { target: devAPI + '/hi', changeOrigin: true },
      '/media': { target: devAPI + '/media', changeOrigin: true },
      '/apihle': { target: devAPI + '/apihle', changeOrigin: true },
    },
  },

  eslint: {
    config: {
      stylistic: true,
      nuxt: {
        sortConfigKeys: true,
      },
    },
  },
})
