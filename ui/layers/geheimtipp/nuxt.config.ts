import { fileURLToPath } from 'node:url'
import { dirname, join } from 'node:path'

// The geheimtipp pool, as a layer of its own.
//
// It is a separate site that happens to be served by this binary: its own users
// in its own database, and none of the family app's chrome around it — save the
// footer, which appears only for an account entitled to more than the pool and
// is the one way back out. Keeping it in a layer is what stops the two leaking
// into each other.
//
// Components are registered from WITHIN the layer using an absolute path derived
// from import.meta.url: a layer's own ~/ alias does not resolve to its app/
// directory. Mirrors the CMS entry layer.
const dir = dirname(fileURLToPath(import.meta.url))

export default defineNuxtConfig({
  // The pool's own classes, the ones Tailwind has no utility for — named grid
  // areas, its purple, the outlined text. Absolute for the same reason the
  // components are: a layer's ~/ does not point at its own app/.

  components: [
    // Prefixed, because both this site and the family app have a Header and a
    // Footer and the app registers its own without a prefix — two <Header>s
    // would leave one of them silently unregistered.
    //
    // pathPrefix keeps directories meaningful as well as tidy: components/match/
    // Row.vue is <GhtMatchRow>, so the same basename can appear under several
    // folders. Not global — these are found by name at build time, so they stay
    // code-split with the page that uses them.
    { path: join(dir, 'app/components'), prefix: 'Ght', pathPrefix: true },
  ],
  css: [join(dir, 'app/assets/css/geheimtipp.css')],
})
