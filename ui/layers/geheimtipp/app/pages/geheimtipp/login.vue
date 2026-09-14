<template>
  <div />
</template>

<script setup lang="ts">
// The pool's sign-in form, retired.
//
// There is one form now, on the front page, and it takes a pool login and
// password exactly as this one did — someone who has always played types what
// they have always typed and is signed in, with an account created behind them
// the first time.
//
// The route stays because links to it do: the pool's own pages linked here for
// years, the gate still sends people here, and a bookmark should land somewhere
// useful rather than on a 404. It forwards, carrying where they were going so
// the sign-in still ends up there.
definePageMeta({ public: true, layout: false })

const route = useRoute()

await navigateTo({
  path: '/',
  query: route.query.weiter ? { next: route.query.weiter } : undefined,
  replace: true,
})
</script>
