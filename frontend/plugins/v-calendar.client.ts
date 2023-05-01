import { defineNuxtPlugin } from "#app";
import DatePicker from "v-calendar";

export default defineNuxtPlugin((nuxtApp) => {
  nuxtApp.vueApp.use(DatePicker);
});
