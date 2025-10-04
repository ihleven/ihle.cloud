<template>
  <main class="flex min-h-screen items-start justify-between space-x-4 bg-gray-100 p-8">
    <section class="w-1/2 rounded border border-gray-200 bg-white px-4 py-2">
      <h1 class="mb-6">Websocket Test-Seite</h1>
      <button @click="startJob">START</button>
      <button @click="ListJob">LIST</button>
      <br />
      <br />
      <br />
      <br />
      <br />
      <br />
    </section>
    <ul class="w-1/2">
      <li v-for="job in queue" :key="job.id" class="mb-2 rounded border border-gray-300 bg-white p-2">
        {{ job }}
      </li>
    </ul>
  </main>
</template>

<script setup>
  const socket = new WebSocket('ws://localhost:8081/ws')

  socket.onopen = () => {
    console.log('Successfully Connected')
    // socket.send('Hi From the Client!')
  }

  socket.onclose = event => {
    console.log('Socket Closed Connection: ', event)
    socket.send('Client Closed!')
  }

  socket.onerror = error => {
    console.log('Socket Error: ', error)
  }
  // socket.addEventListener('message', event => {})

  function startJob() {
    socket.send('START')
  }

  function ListJob() {
    socket.send('LIST')
  }

  const queue = ref([])

  socket.onmessage = event => {
    // const f = document.getElementById('chatbox').contentDocument
    // let text = ''
    const msg = JSON.parse(event.data)
    // const time = new Date(msg.date)
    // const timeStr = time.toLocaleTimeString()
    console.log('MESSAGE =>', msg)

    switch (msg.type) {
      case 'list':
        queue.value = msg.payload
        break
      // case 'username':
      //   text = `User <em>${msg.name}</em> signed in at ${timeStr}<br>`
      //   break
      // case 'message':
      //   text = `(${timeStr}) ${msg.name} : ${msg.text} <br>`
      //   break
      // case 'rejectusername':
      //   text = `Your username has been set to <em>${msg.name}</em> because the name you chose is in use.<br>`
      //   break
      // case 'userlist':
      //   document.getElementById('userlistbox').innerHTML = msg.users.join('<br>')
      //   break
    }
  }
</script>
