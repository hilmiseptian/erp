import api from "@/services/api";

const stockService = {
  branches() { return api.get("/stocks/branches"); },
  list(params) { return api.get("/stocks", { params }); },
  cards(branchId, itemId, params) { return api.get(`/stocks/${branchId}/${itemId}/cards`, { params }); },
};

export default stockService;
