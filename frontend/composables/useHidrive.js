// import { useState } from '#app'

export default async function () {
  const route = useRoute();

  const path = computed(() => route.params.slug.join("/"));

  console.log("slug", `${path.value}`);

  const { data, refresh } = await useFetch(`/api/${path.value}`, {
    server: false,
    key: path.value,
  });
  const files = useState("files", () => data);

  files.value = data;

  watch(
    () => path,
    () => refresh()
  );

  return {
    path,
    files,
    refresh,
  };
}
