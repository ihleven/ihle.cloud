<template>
  <div class="min-h-full w-full bg-elevated">
    <!-- The green band with today in it, the same on every page below this one:
         the calendar always says what day it is, whatever you have navigated
         to. -->
    <section class="bg-green-500 p-2">
      <KalenderHero />
    </section>

    <NuxtPage />
  </div>
</template>

<script setup lang="ts">
// The calendar, carried over from a migration that stopped here.
//
// What exists is the frame: the band that says today, and routes for a year, a
// month, a day and an ISO week that nest inside it. What each of those shows is
// still a placeholder naming the part of the date it was reached by, exactly as
// the previous attempt left them — they mark where the work resumes rather than
// pretending to be a calendar.
//
// The chrome is the app's own. The old page drew its own navigation bar because
// it had no layout to put it in.

// The page is bg-elevated, but nothing of it is visible at the top: the green
// band saying what day it is sits directly under the strip, so that is what the
// fade has to arrive at. Fading to the grey instead would not remove the line —
// it would move it down by the height of the band.
definePageMeta({
  shell: {
    background: 'var(--ui-bg-elevated)',
    bandTo: 'var(--color-green-500)',
    // Through a pale green rather than straight from white. White to a
    // saturated green in one step passes through a grey-green nothing in the
    // palette is, and it is the part of the ramp the eye reads as dirty; going
    // by way of the ramp's own 200 keeps the whole band in the greens. Held
    // late, so the colour arrives near the block rather than filling the band.
    bandVia: 'var(--color-green-200) 72%',
  },
})
</script>
