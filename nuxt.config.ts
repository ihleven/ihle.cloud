import { defineNuxtConfig } from 'nuxt'

// https://v3.nuxtjs.org/api/configuration/nuxt.config
export default defineNuxtConfig({
    modules: [
        '@nuxt/content',
        '@nuxtjs/tailwindcss',
        '@kevinmarrec/nuxt-pwa'
    ],
    content: {
        documentDriven: true
    },
    pwa: {
        workbox: {
            enabled: false
        }
    },
    tailwindcss: {
        // configPath: '~/tailwind.config.js',
        // cssPath: '@/tailwind.css',
      }
})
