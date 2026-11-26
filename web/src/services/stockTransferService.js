import api from '@/services/api'

const stockTransferService = {
  branches() { return api.get('/stock-transfers/branches') },
  items(branchId, params = {}) { return api.get('/stock-transfers/items', { params: { branch_id: branchId, ...params } }) },
  list(params) { return api.get('/stock-transfers', { params }) },
  get(id) { return api.get(`/stock-transfers/${id}`) },
  create(payload) { return api.post('/stock-transfers', payload) },
  update(id, payload) { return api.put(`/stock-transfers/${id}`, payload) },
  dispatch(id, payload) { return api.post(`/stock-transfers/${id}/dispatches`, payload) },
  receive(id, payload) { return api.post(`/stock-transfers/${id}/receipts`, payload) },
  cancel(id) { return api.patch(`/stock-transfers/${id}/cancel`) },
}

export default stockTransferService
