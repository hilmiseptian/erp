import { defineStore } from 'pinia'
import { ref } from 'vue'
import laravel from '@/services/laravel'
import inventory from '@/services/inventory'

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('token') || '')
  const user = ref(null)
  const isAuthenticated = ref(false)
  const loading = ref(false)

  async function login(credentials) {
    loading.value = true

    try {
      const response = await laravel.post('/login', credentials)

      const accessToken = response.data.token

      localStorage.setItem('token', accessToken)
      token.value = accessToken

      await fetchUser()

      return true
    } catch (error) {
      logoutLocal()
      throw error
    } finally {
      loading.value = false
    }
  }

  async function fetchUser() {
    try {
      const response = await inventory.get('/me')

      user.value = response.data.user
      isAuthenticated.value = true

      return user.value
    } catch (error) {
      logoutLocal()
      throw error
    }
  }

  function logoutLocal() {
    localStorage.removeItem('token')
    token.value = ''
    user.value = null
    isAuthenticated.value = false
  }

  async function logout() {
    try {
      await laravel.post('/logout', {}, {
        headers: {
          Authorization: `Bearer ${token.value}`,
        },
      })
    } finally {
      logoutLocal()
    }
  }

  return {
    token,
    user,
    isAuthenticated,
    loading,
    login,
    fetchUser,
    logout,
    logoutLocal,
  }
})