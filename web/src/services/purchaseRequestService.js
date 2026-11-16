import api from "@/services/api";

const purchaseRequestService = {
  list(params) {
    return api.get("/purchase-requests", { params });
  },
  branches() {
    return api.get("/purchase-requests/branches");
  },
  suppliers() {
    return api.get("/purchase-requests/suppliers");
  },
  get(id) {
    return api.get(`/purchase-requests/${id}`);
  },
  create(payload) {
    return api.post("/purchase-requests", payload);
  },
  update(id, payload) {
    return api.put(`/purchase-requests/${id}`, payload);
  },
  cancel(id) {
    return api.patch(`/purchase-requests/${id}/cancel`);
  },
};

export default purchaseRequestService;
