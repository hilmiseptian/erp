<script setup>
import { useAuthStore } from "@/stores/auth";

const auth = useAuthStore();

async function handleLogout() {
  await auth.logout();
  window.location.href = "/login";
}
</script>

<template>
  <main>
    <h1>Inventory Dashboard</h1>

    <p>Welcome, {{ auth.user?.name }}</p>
    <p>{{ auth.user?.email }}</p>

    <button @click="handleLogout">Logout</button>
  </main>
</template>
