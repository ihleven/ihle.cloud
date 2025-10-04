<template>
  <main class="min-h-screen bg-white py-12">
    <button @click="authorize">asd</button>
    <MDC :doc="doc" />
    <NuxtLink to="/">root</NuxtLink>
  </main>
</template>

<script setup lang="ts">
definePageMeta({
  middleware: [
    function (to, from) {
      const hitoken = useState('hitoken', () => useCookie('hitoken'))
      // const hitoken = useCookie('hitoken')

      if (!hitoken.value) {
        console.log('hitoken undefined', process.env)
        return navigateTo('/tokenlogin')
      }

      // counter.value = counter.value || Math.round(Math.random() * 1000)
      console.log('middleware', hitoken.value, to.fullPath, from.fullPath)
    },

  ],
})
const route = useRoute()
const path = typeof route.params.slug === 'string' ? route.params.slug : route.params.slug.join('/')
const dir = dirname(path)

async function authorize() {
  const formData = new FormData()
  formData.append('token', 'asdfasdfasdf')

  const data = await $fetch('http://localhost:8000/tokenauth', { method: 'POST', body: formData })

  console.log('data', data, formData)
}

provide('basePath', `/api/home`)
provide('dirPath', dir)

function dirname(input: string): string {
  const i = input.lastIndexOf('/')
  return input.slice(0, i)
}
</script>
