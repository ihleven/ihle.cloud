export default defineAppConfig({

  // This app's contribution to the CMS entry layer's registries: content type ->
  // editor component name, and content type -> preview card in the directory
  // browser. Both are deep-merged with the layer's own, which carries only the
  // fallback and the directory itself.
  contentTypes: {
    Person: 'Person',
    Work: 'Artwork',
    Film: 'Film',
  },

  contentTypePanels: {
    Person: 'PersonPanel',
    Film: 'FilmPanel',
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
