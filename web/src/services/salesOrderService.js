import api from '@/services/api'

const salesOrderService = {
  list(params) { return api.get('/sales-orders', { params }) },
  branches() { return api.get('/sales-orders/branches') },
  get(id) { return api.get(`/sales-orders/${id}`) },
  create(payload) { return api.post('/sales-orders', payload) },
  update(id, payload) { return api.put(`/sales-orders/${id}`, payload) },
  cancel(id) { return api.patch(`/sales-orders/${id}/cancel`) },
}

export default salesOrderService
