import api from './api'

const customerService = {
  list(params = {}) { return api.get('/customers', { params }) },
  get(id) { return api.get(`/customers/${id}`) },
  create(payload) { return api.post('/customers', payload) },
  update(id, payload) { return api.put(`/customers/${id}`, payload) },
  setActive(id, is_active) { return api.patch(`/customers/${id}/status`, { is_active }) },
}

export default customerService
