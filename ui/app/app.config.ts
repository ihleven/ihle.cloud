export default defineAppConfig({

  api: {
    baseURL: 'http://localhost:8000',
  },

  toaster: {
    duration: 5000,
    position: 'top-right' as const,
  },

  ui: {
    colors: {
      primary: 'sky',
      neutral: 'zinc',
    },

    modal: {
      slots: {
        overlay: 'fixed inset-0 bg-inverted/75',
      },
    },
  },
})
