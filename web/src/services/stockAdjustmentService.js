import api from "@/services/api";

const stockAdjustmentService = {
  list(params) { return api.get("/stock-adjustments", { params }); },
  get(id) { return api.get(`/stock-adjustments/${id}`); },
  create(data) { return api.post("/stock-adjustments", data); },
  update(id, data) { return api.put(`/stock-adjustments/${id}`, data); },
  post(id) { return api.post(`/stock-adjustments/${id}/post`); },
  cancel(id) { return api.patch(`/stock-adjustments/${id}/cancel`); },
};

export default stockAdjustmentService;
