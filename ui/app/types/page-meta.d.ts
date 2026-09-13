// A page may declare itself reachable without a session. app.vue gates the whole
// app on one, so a page that is meant to be opened by someone not signed in —
// an enrollment link, a public film — has to say so.
declare module '#app' {
  interface PageMeta {
    public?: boolean
  }
}

export {}
