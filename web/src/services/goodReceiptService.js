import api from "@/services/api";

const goodReceiptService = {
  list(params) {
    return api.get("/good-receipts", { params });
  },
  sources() {
    return api.get("/good-receipts/purchase-orders");
  },
  source(poNumber) {
    return api.get(`/good-receipts/purchase-orders/${encodeURIComponent(poNumber)}`);
  },
  get(id) {
    return api.get(`/good-receipts/${id}`);
  },
  create(payload) {
    return api.post("/good-receipts", payload);
  },
};

export default goodReceiptService;
