<template>
  <section class="bg-green-500 p-2">
    <div class="flex justify-between items-start">
      <div class="left">
        <div class="flex text-outline">
          <span class="p-2 flex flex-col justify-between items-end">
            <nuxt-link
              class="font-medium text-2xl"
              :to="{
                name: 'kalender-jahr-monat',
                params: { jahr: jahr, monat: monat },
              }"
            >
              {{ monatsname }}
            </nuxt-link>
            <nuxt-link class="kw" :to="'/kalender/' + kw.jahr + '/KW' + kw.nr">
              <span v-if="kw.nr">KW{{ kw.nr }}</span>
            </nuxt-link>
            <nuxt-link class="text-3xl font-black" :to="'/kalender/' + jahr">{{
              jahr
            }}</nuxt-link>
          </span>
          <span v-show="tag" class="text-8xl font-black leading-[6rem]">{{
            tag
          }}</span>
        </div>
        <div class="wochentag">{{ wochentag }}</div>
      </div>
      <div class="right">
        <ClientOnly>
          <!-- <v-calendar v-model="date" @dayclick="changeDate" /> -->
          <v-date-picker
            show-iso-weeknumbers
            :modelValue="date"
            @update:modelValue="changeDate"
            @update:to-page="changePage"
            @weeknumberclick="onWN"
          />
        </ClientOnly>
        <!-- <no-ssr> -->
        <!-- <v-calendar
          :value="selectedDate"
          is-inline
          color="teal"
          is-dark
          @dayclick="selectDate"
        /> -->
        <!-- </no-ssr> -->
      </div>
    </div>
  </section>
</template>

<script setup>
// import { DatePicker } from "v-calendar";
const tag = computed(() => date.value.getDate());
const monat = computed(() => date.value.getMonth());
const jahr = computed(() => date.value.getFullYear());

const kw = "tag";
const monatsname = computed(() =>
  date.value.toLocaleDateString("de-DE", { month: "long" })
);
const wochentag = computed(() =>
  date.value.toLocaleDateString("de-DE", { weekday: "long" })
);

const date = ref(new Date());
function changeDate(d) {
  const year = d.getFullYear();
  const month = d.getMonth() + 1;
  const day = d.getDate();

  date.value = d;
  navigateTo({ path: `/kalender/${year}/${month}/${day}` });
}
function changePage(page) {
  date.value.setFullYear(page.year);
  date.value.setMonth(page.month - 1);
  console.log("changePage:", page);
  navigateTo({ path: `/kalender/${page.year}/${page.month}` });
}

function onWN(week) {
  console.log("weeknumber:", week);
  navigateTo({
    path: `/kalender/${date.value.getFullYear()}/kw${week.weeknumber}`,
  });
}
</script>

<script>
// import { mapState } from 'vuex'

// import FeatherIcon from '@/components/FeatherIcon.vue'

export default {
  name: "ArbeitHero",
  //   components: {
  //     FeatherIcon,
  //   },
  data() {
    return {
      week: "KW23",
      selectedDate: new Date(),
    };
  },

  // computed: {
  //   ...mapState('datum', ['jahr', 'monat', 'monatsname', 'tag', 'wochentag', 'kw']),
  // },

  methods: {
    selectDate(d) {
      console.log("event:", d);
      if (d.date === this.selectedDate.getDate()) return;
      this.$router.push({
        name: "kalender-jahr-monat-tag",
        params: {
          jahr: d.date.getFullYear(),
          monat: d.date.getMonth() + 1,
          tag: d.date.getDate(),
        },
      });
    },
  },
};
</script>

<style scoped>
.day {
  background-color: rgba(100, 30, 50, 0);
  font: ultra-condensed normal 900 6rem / 1 "Raleway";
}

.wrapper {
  display: inline-block;
  background-color: rgba(100, 30, 50, 0);
  position: relative;
}

.monat {
  position: relative;
  top: -2rem;
  left: -1.4rem;
  background-color: rgba(0, 150, 20, 0);
  font: 600 2rem / 1 "Raleway";
}

.kw {
  position: relative;
  top: -1rem;
  left: 1rem;
  background-color: rgba(0, 150, 20, 0);
  font: 300 1.5rem / 1 "Raleway";
}

.jahr {
  position: absolute;
  top: 1.8rem;
  left: 2rem;
  background-color: rgba(0, 150, 20, 0);
  font: 700 2.5rem / 1 "Raleway";
}

.wochentag {
  font: 600 2rem / 1 "Raleway";
}
</style>
