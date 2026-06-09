import axios from 'axios'
import { snakeToCamel, camelToSnake } from './transform'

const api = axios.create({
  baseURL: '/api',
  headers: { 'Content-Type': 'application/json' },
})

// 응답: snake_case → camelCase
api.interceptors.response.use((response) => {
  if (response.data && String(response.headers['content-type'] ?? '').includes('application/json')) {
    response.data = snakeToCamel(response.data)
  }
  return response
})

// 요청: camelCase → snake_case
api.interceptors.request.use((config) => {
  if (config.data && typeof config.data === 'object') {
    config.data = camelToSnake(config.data)
  }
  return config
})

export default api
