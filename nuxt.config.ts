export default defineNuxtConfig({
  ssr: false,
  srcDir: "frontend/",
  // app: {
  //     pageTransition: { name: 'page', mode: 'out-in' }
  //   },
  modules: [
    "@nuxt/content",
    "@nuxtjs/tailwindcss",
    "@kevinmarrec/nuxt-pwa",
    "nuxt-icon",
    "nuxt-icons", //https://github.com/gitFoxCode/nuxt-icons
  ],

  content: {
    documentDriven: false,
  },
  pwa: {
    meta: {
      theme_color: "#e5e5e5",
    },
    manifest: {
      lang: "de",
      name: "Ihle.Cloud",
      short_name: "iCloud",
      description: "Die Ihle Cloud",
      display: "standalone",
    },
    workbox: {
      enabled: false,
    },
  },
  tailwindcss: {
    configPath: "@/tailwind.config.js",
    cssPath: "@/tailwind.css",
  },

  runtimeConfig: {
    apiBaseUrl: "",
  },
});
