import axios from 'axios'

const laravel = axios.create({
  baseURL: import.meta.env.VITE_LARAVEL_API_URL,
  headers: {
    Accept: 'application/json',
  },
})

export default laravel