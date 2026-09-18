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
</script>
