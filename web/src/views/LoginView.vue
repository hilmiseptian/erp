<script setup>
import { ref } from "vue";
import { useRouter } from "vue-router";
import { useAuthStore } from "@/stores/auth";

const router = useRouter();
const auth = useAuthStore();

const login = ref("");
const password = ref("");
const errorMessage = ref("");

async function handleLogin() {
  errorMessage.value = "";

  try {
    await auth.login({
      login: login.value,
      password: password.value,
    });

    router.push("/dashboard");
  } catch (error) {
    errorMessage.value = error.response?.data?.message || "Login failed";
  }
}
</script>

<template>
  <div class="login-container">
    <h1>Inventory ERP</h1>
    <p>Sign in to continue</p>

    <form @submit.prevent="handleLogin">
      <div>
        <label>Email/Username</label>
        <input v-model="login" type="test" required placeholder="Enter your email or password" />
      </div>

      <div>
        <label>Password</label>
        <input v-model="password" type="password" required placeholder="Enter your password" />
      </div>

      <p v-if="errorMessage" class="error">
        {{ errorMessage }}
      </p>

      <button type="submit" :disabled="auth.loading">
        {{ auth.loading ? "Signing in..." : "Login" }}
      </button>
    </form>
  </div>
</template>

<style scoped>
.login-container {
  max-width: 400px;
  margin: 100px auto;
  padding: 24px;
}

form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

input {
  display: block;
  width: 100%;
  padding: 10px;
  margin-top: 6px;
}

button {
  padding: 12px;
  cursor: pointer;
}

.error {
  color: red;
}
</style>
