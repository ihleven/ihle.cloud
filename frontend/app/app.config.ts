export default defineAppConfig({
  title: 'Hello Nuxt',
  api: {
    origin: 'process.env.NUXT_PUBLIC_API_ORIGIN',
    basePath: 'process.env.NUXT_PUBLIC_API_BASE_PATH',
    baseURL: 'process.env.NUXT_PUBLIC_API_BASE_URL',
  },

  toaster: {
    duration: 5000,
    position: 'top-right',
  },

  ui: {
    colors: {
    //   cta: 'orange',
      primary: 'sky',
      //   secondary: 'rose',
      //   success: 'emerald',
      //   info: 'cyan',
      //   warning: 'amber',
      //   error: 'red',
      neutral: 'zinc', // zink
    },

    //   button: {
    //     slots: {
    //       base: 'font-bold',
    //     },
    //   },
    modal: {
      slots: {
        overlay: 'fixed inset-0 bg-inverted/75',
        // content: 'fixed w-full h-dvh bg-[var(--ui-bg)] divide-y divide-[var(--ui-border)] flex flex-col focus:outline-none',
        // header: 'px-4 py-5 sm:px-6',
        // body: 'flex-1 p-4 sm:p-6',
        // footer: 'flex items-center gap-1.5 p-2 sm:px-6',
        // title: 'text-[var(--ui-text-highlighted)] font-semibold',
        // description: 'mt-1 text-[var(--ui-text-muted)] text-sm',
        // close: 'absolute top-4 right-4',
      },
      // variants: {
      //   transition: {
      //     true: {
      //       overlay: 'data-[state=open]:animate-[fade-in_200ms_ease-out] data-[state=closed]:animate-[fade-out_200ms_ease-in]',
      //       content: 'data-[state=open]:animate-[scale-in_200ms_ease-out] data-[state=closed]:animate-[scale-out_200ms_ease-in]',
      //     },
      //   },
      //   fullscreen: {
      //     true: {
      //       content: 'inset-0',
      //     },
      //     false: {
      //       content: 'top-[50%] left-[50%] translate-x-[-50%] translate-y-[-50%] sm:max-w-lg sm:h-auto sm:my-8 sm:rounded-[calc(var(--ui-radius)*2)] sm:shadow-lg sm:ring ring-[var(--ui-border)]',
      //     },
      //   },
      // },
    },
  },

})
