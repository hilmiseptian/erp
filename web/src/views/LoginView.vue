
<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { Boxes, LockKeyhole, UserRound } from 'lucide-vue-next'

const router = useRouter()
const auth = useAuthStore()

const login = ref('')
const password = ref('')
const errorMessage = ref('')

async function handleLogin() {
  errorMessage.value = ''

  try {
    await auth.login({
      login: login.value,
      password: password.value,
    })

    await router.replace('/dashboard')
  } catch (error) {
    errorMessage.value =
      error.response?.data?.message || 'Login failed. Please try again.'
  }
}
</script>

<template>
  <div class="flex min-h-screen items-center justify-center bg-slate-100 px-4 py-8">

    <div class="w-full max-w-md rounded-2xl border border-slate-200 bg-white p-8 shadow-lg">

      <!-- Logo -->
      <div class="mb-8 text-center">
        <div class="mx-auto flex h-14 w-14 items-center justify-center rounded-xl bg-blue-600 text-white">
          <Boxes :size="30" />
        </div>

        <h1 class="mt-4 text-2xl font-bold text-slate-800">
          ERP Application
        </h1>

        <p class="mt-2 text-sm text-slate-500">
          Sign in to manage your business operations.
        </p>
      </div>

      <!-- Login Form -->
      <form @submit.prevent="handleLogin" class="space-y-5">

        <div>
          <label
            for="login"
            class="mb-2 block text-sm font-medium text-slate-700"
          >
            Email / Username
          </label>

          <div class="relative">
            <UserRound
              :size="18"
              class="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400"
            />

            <input
              id="login"
              v-model="login"
              type="text"
              required
              autocomplete="username"
              placeholder="Enter your email or username"
              class="w-full rounded-lg border border-slate-300 py-3 pl-10 pr-3 text-sm outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-100"
            />
          </div>
        </div>

        <div>
          <label
            for="password"
            class="mb-2 block text-sm font-medium text-slate-700"
          >
            Password
          </label>

          <div class="relative">
            <LockKeyhole
              :size="18"
              class="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400"
            />

            <input
              id="password"
              v-model="password"
              type="password"
              required
              autocomplete="current-password"
              placeholder="Enter your password"
              class="w-full rounded-lg border border-slate-300 py-3 pl-10 pr-3 text-sm outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-100"
            />
          </div>
        </div>

        <div
          v-if="errorMessage"
          class="rounded-lg bg-red-50 p-3 text-sm text-red-600"
        >
          {{ errorMessage }}
        </div>

        <button
          type="submit"
          :disabled="auth.loading"
          class="w-full rounded-lg bg-blue-600 py-3 text-sm font-semibold text-white transition hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-60"
        >
          {{ auth.loading ? 'Signing in...' : 'Sign In' }}
        </button>

      </form>

      <p class="mt-8 text-center text-xs text-slate-400">
        ERP Application
      </p>

    </div>

  </div>
</template>
