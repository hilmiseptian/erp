import api from '@/services/api'

const goodDeliveryService = {
  list(params) { return api.get('/good-deliveries', { params }) },
  sources(refType = 'sales_order') { return api.get('/good-deliveries/sources', { params: { ref_type: refType } }) },
  source(refType, refNum) { return api.get(`/good-deliveries/sources/${encodeURIComponent(refType)}/${encodeURIComponent(refNum)}`) },
  get(id) { return api.get(`/good-deliveries/${id}`) },
  create(payload) { return api.post('/good-deliveries', payload) },
}

export default goodDeliveryService
