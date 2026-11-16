<script setup>
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { SwaggerUIBundle } from 'swagger-ui-dist'
import 'swagger-ui-dist/swagger-ui.css'

const mountPoint = ref(null)
let swaggerInstance
const apiBase = (import.meta.env.VITE_INVENTORY_API_URL || 'http://localhost:8080')
  .replace(/\/$/, '')
  .replace(/\/api$/i, '')

onMounted(() => {
  swaggerInstance = SwaggerUIBundle({
    url: `${apiBase}/swagger.yaml`,
    domNode: mountPoint.value,
    deepLinking: true,
    displayRequestDuration: true,
    persistAuthorization: true,
    tryItOutEnabled: true,
  })
})

onBeforeUnmount(() => {
  swaggerInstance = undefined
})
</script>

<template>
  <section class="space-y-4">
    <header>
      <h1 class="text-2xl font-bold text-slate-800">API Documentation</h1>
      <p class="mt-1 text-sm text-slate-500">Jelajahi endpoint ERP dan coba request API. Gunakan Authorize untuk memasukkan token POS.</p>
    </header>
    <div ref="mountPoint" class="swagger-ui rounded-xl bg-white p-3 shadow-sm"></div>
  </section>
</template>

<style>
.swagger-ui .wrapper { padding: 0; }
.swagger-ui .topbar { display: none; }
.swagger-ui .scheme-container { padding: 16px 0; box-shadow: none; }
</style>
