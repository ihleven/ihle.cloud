// The geheimtipp pool, as a layer of its own.
//
// It is a separate site that happens to be served by this binary: public, with
// its own sign-in against its own backend, and with nothing of the family app's
// chrome or entitlements around it. Keeping it in a layer is what stops those
// two ideas leaking into each other.
export default defineNuxtConfig({})
