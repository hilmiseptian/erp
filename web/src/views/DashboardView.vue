<script setup>
import { useAuthStore } from "@/stores/auth";

const auth = useAuthStore();

async function handleLogout() {
  await auth.logout();
  window.location.href = "/login";
}
</script>

<template>
  <div class="p-8">
    <h1 class="text-3xl font-bold text-blue-600">
      Tailwind CSS is working!
    </h1>

    <p class="mt-2 text-slate-500">
      ERP frontend is ready.
    </p>
  </div>
</template>
