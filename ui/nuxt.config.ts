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
  // The Go reference for this app's own packages, rendered natively by the SPA
  // from the JSON the backend serves. The layer is the CMS's, extended rather
  // than copied; app.config below points it at this module instead of the CMS's.
  //
  // The geheimtipp pool, which is a separate site with its own backend and its
  // own sign-in rather than an area of this app — see ui/layers/geheimtipp.
  extends: ['../../cms/ui/layers/entry', '../../cms/ui/layers/godoc', './layers/geheimtipp', './layers/art'],

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
    // Components a markdown body may use — MDC resolves them by name while
    // rendering, so the same rule applies: they have to be global.
    { path: '~/components/content', pathPrefix: false, global: true },
    { path: '~/components', pathPrefix: false },
  ],

  devtools: { enabled: true },

  // Added to the home screen on an iPhone, the app opens without Safari's
  // address bar and toolbar. This is a manifest and a few meta tags and nothing
  // else — deliberately no service worker: iOS does not need one for standalone
  // mode, and a precaching worker is what makes a deployed update arrive late.
  app: {
    head: {
      // viewport-fit=cover lays the page out across the whole screen and, with
      // that, makes env(safe-area-inset-*) report real numbers. Installed, the
      // page is under the status bar either way; this is what says by how much,
      // and the strip at the top of the header is sized from it.
      viewport: 'width=device-width, initial-scale=1, viewport-fit=cover',
      link: [
        { rel: 'manifest', href: '/manifest.webmanifest' },
        { rel: 'apple-touch-icon', href: '/apple-touch-icon.png' },
      ],
      meta: [
        // The modern name, and the one iOS has always read.
        { name: 'mobile-web-app-capable', content: 'yes' },
        { name: 'apple-mobile-web-app-capable', content: 'yes' },
        { name: 'apple-mobile-web-app-title', content: 'ihleven' },
        // Who owns the status bar. "black" is the value that lets the page
        // reach under it on both devices here — measured, not assumed: under
        // "default" an iPhone still reported an inset of 50 while an iPad
        // reported 0 and had no bleed at all, which is not the documented
        // split. /debug/haze is where that was read off.
        //
        // Neither of the other two colour tags reaches the bar: the manifest's
        // theme_color stops at the splash screen, and the theme-color meta
        // below is not read for the bar at all.
        //
        // Installed, the value is taken from the page the app was added from,
        // so a change needs a build and the icon removed and added again.
        { name: 'apple-mobile-web-app-status-bar-style', content: 'black' },
        // A starting value only: app.vue replaces it with the colour of the
        // page you are actually on. Not for the status bar — that is the tag
        // above and the pixels the page paints — but for whatever browser
        // chrome elsewhere reads this, where one written colour cannot match
        // pages that choose their own.
        //
        // The value is the colour the pages have always been, written out
        // because this is read before any of ours runs — a var would resolve to
        // nothing. Signalgrau, --color-ral-7004; the manifest says black, which
        // is the splash screen's and nothing else's.
        { name: 'theme-color', content: '#9b9b9b' },
      ],
    },
  },
  // ssr: false,
  css: ['~/assets/css/main.css'],
  ui: {
    colorMode: false,
  },

  runtimeConfig: {
    public: {
      api: {
        // Relative on purpose: the app calls whatever origin served it. An
        // absolute URL here would be baked into the build, so the app would call
        // one host while being served from another — which is cross-origin, and
        // then cookies are withheld and CORS has to be configured to match.
        //
        // In development the API is genuinely on another port; devProxy below
        // bridges that instead, so requests stay same-origin there too.
        base: '',
      },
      // This app's own API. Relative like everything else here, so the app
      // calls whichever origin served it. The CMS's layers do not read this —
      // they have a key of their own, contentApiBase, which is the difference
      // between a host configuring a layer and a host giving up a name.
      apiBaseURL: '/api/v1',

      // Where the geheimtipp pages fetch from. It is this app's proxy today,
      // because the pool's own API is on another origin and cannot be called
      // across it. Configuration rather than a constant, so moving to the same
      // origin as the pool later is this one value, not a rewrite.
      geheimtippBase: '/ght',

      // Who keeps the pool's results. Entering a match result is the one thing
      // the pool offers to a single person, and its own frontend writes that
      // login into a component. Stated here instead: which account runs a
      // deployment is a property of the deployment. The backend decides whether
      // to accept the write either way — this only decides whether to offer it.
      geheimtippAdmin: 'matt',
    },
  },

  // A production build can be sent to its own directory, so that a `nuxt dev`
  // server left running — which keeps rewriting the build directory in
  // development mode — cannot leak its own asset URLs into the generated HTML.
  // Unset, this is the usual .nuxt, so `nuxt dev` is unaffected.
  buildDir: process.env.NUXT_BUILD_DIR || '.nuxt',

  compatibilityDate: '2025-07-15',

  // Development only: Nuxt serves the app on :3000 while the Go server holds the
  // API. Proxying keeps the browser talking to one origin, which is what makes
  // session and passkey cookies behave the same in development as in production.
  nitro: {
    devProxy: {
      '/auth': { target: devAPI + '/auth', changeOrigin: true },
      '/api': { target: devAPI + '/api', changeOrigin: true },
      // Named down to the one route the server has under it, for the same
      // reason as /media/videos below: a prefix is matched as a string and not
      // as a path, so '/hi' also claims /hidrive and the whole area is answered
      // by the API instead of by the app.
      '/hi/auth': { target: devAPI + '/hi/auth', changeOrigin: true },
      // Named down to the one route the server has under /media, because a
      // prefix is matched as a string and not as a path: `/media` also claims
      // /mediathek, and the whole area is then answered by the API instead of
      // by the app. A trailing slash does not help — the match ignores it.
      '/media/videos': { target: devAPI + '/media/videos', changeOrigin: true },
      '/apihle': { target: devAPI + '/apihle', changeOrigin: true },
      // The CMS's API, which the Go server mounts whole under its own
      // prefix. Every layer this app takes from the CMS fetches there — the
      // entry editor, the package docs, the accounts and the old screens —
      // so without this none of them has a backend to talk to.
      '/content-api': { target: devAPI + '/content-api', changeOrigin: true },
      // The pool's backend, which the Go server proxies onward. Without this the
      // dev server's catch-all answers /ght/* with the SPA shell — HTML where a
      // page expects JSON, which surfaces as a fetch failure rather than a 404.
      '/ght': { target: devAPI + '/ght', changeOrigin: true },
    },
  },

  // The band under the strip fades to whatever is directly beneath it, and a
  // page says what that is. /content cannot: it is the CMS layer's page, not
  // this app's, so there is no definePageMeta of ours to put it in.
  //
  // What is beneath it there is the module's own chrome — the content bar and
  // the entry header, which share the elevated tone so that the crumb trail,
  // the entry's name and its toolbar read as one slab. Fading to the page's
  // colour instead would put the join back exactly where the band exists to
  // remove it.
  hooks: {
    'pages:extend': (pages) => {
      const content = pages.find(p => p.path.startsWith('/content'))
      if (content) content.meta = { ...content.meta, shell: { bandTo: 'var(--ui-bg-elevated)' } }
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
