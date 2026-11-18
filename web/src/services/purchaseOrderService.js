import api from "@/services/api";

const purchaseOrderService = {
  list(params) {
    return api.get("/purchase-orders", { params });
  },
  sources() {
    return api.get("/purchase-orders/purchase-requests");
  },
  source(prNumber) {
    return api.get(`/purchase-orders/purchase-requests/${encodeURIComponent(prNumber)}`);
  },
  get(id) {
    return api.get(`/purchase-orders/${id}`);
  },
  create(payload) {
    return api.post("/purchase-orders", payload);
  },
  cancel(id) {
    return api.patch(`/purchase-orders/${id}/cancel`);
  },
};

export default purchaseOrderService;
