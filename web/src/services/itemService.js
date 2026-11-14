import api from './api'

const itemService = {
  list(params) { return api.get('/items', { params }) },
  categories() { return api.get('/items/categories') },
  get(id) { return api.get(`/items/${id}`) },
  create(data) { return api.post('/items', data) },
  update(id, data) { return api.put(`/items/${id}`, data) },
  setActive(id, is_active) { return api.patch(`/items/${id}/status`, { is_active }) },
}

export default itemService
