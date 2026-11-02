import axios from 'axios'

const inventory = axios.create({
  baseURL: import.meta.env.VITE_INVENTORY_API_URL,
  headers: {
    Accept: 'application/json',
  },
})

inventory.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')

  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }

  return config
})

export default inventory