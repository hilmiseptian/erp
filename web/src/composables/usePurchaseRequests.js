import { computed, onMounted, ref, watch } from "vue";
import itemService from "@/services/itemService";
import purchaseRequestService from "@/services/purchaseRequestService";
import { toDateInputValue } from "@/utils/formatters";

function todayAsDateInput() {
  const today = new Date();
  const year = today.getFullYear();
  const month = String(today.getMonth() + 1).padStart(2, "0");
  const day = String(today.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}

function emptyForm(branches, suppliers, today) {
  return {
    branch_id: branches[0] ? String(branches[0].id) : "",
    supplier_id: suppliers[0] ? String(suppliers[0].id) : "",
    pr_date: today,
    req_date: today,
    lines: [{ item_id: "", qty: "1", price: "0" }],
  };
}

export function usePurchaseRequests() {
  const requests = ref([]);
  const branches = ref([]);
  const suppliers = ref([]);
  const items = ref([]);
  const loading = ref(false);
  const saving = ref(false);
  const page = ref(1);
  const pagination = ref({ total: 0, total_pages: 0 });
  const search = ref("");
  const statusFilter = ref("");
  const branchFilter = ref("");
  const error = ref("");
  const message = ref("");
  const dialog = ref(false);
  const selected = ref(null);
  const formError = ref("");
  const today = todayAsDateInput();
  const form = ref(emptyForm([], [], today));
  const totalPages = computed(() => Math.max(1, pagination.value.total_pages || 0));
  const statuses = {
    open: "Open",
    partially_ordered: "Partially ordered",
    ordered: "Ordered",
    cancelled: "Cancelled",
  };

  async function loadRequests() {
    loading.value = true;
    error.value = "";
    try {
      const { data } = await purchaseRequestService.list({
        page: page.value,
        limit: 20,
        search: search.value,
        status: statusFilter.value,
        branch_id: branchFilter.value,
      });
      requests.value = data.data || [];
      pagination.value = data.pagination || { total: 0, total_pages: 0 };
    } catch (err) {
      error.value = err.response?.data?.message || "Failed to load purchase requests.";
    } finally {
      loading.value = false;
    }
  }

  watch([statusFilter, branchFilter], () => {
    page.value = 1;
    loadRequests();
  });
  watch(page, loadRequests);

  function openCreate() {
    selected.value = null;
    formError.value = "";
    form.value = emptyForm(branches.value, suppliers.value, today);
    dialog.value = true;
  }

  async function openEdit(row) {
    formError.value = "";
    try {
      const { data } = await purchaseRequestService.get(row.id);
      selected.value = data.data;
      form.value = {
        branch_id: String(data.data.branch_id),
        supplier_id: String(data.data.supplier_id),
        pr_date: toDateInputValue(data.data.pr_date),
        req_date: toDateInputValue(data.data.req_date),
        lines: data.data.lines.map((line) => ({
          item_id: String(line.item_id),
          qty: line.qty,
          price: line.price,
        })),
      };
      dialog.value = true;
    } catch (err) {
      error.value = err.response?.data?.message || "Failed to open purchase request.";
    }
  }

  function addLine() {
    form.value.lines.push({ item_id: "", qty: "1", price: "0" });
  }

  function removeLine(index) {
    if (form.value.lines.length > 1) form.value.lines.splice(index, 1);
  }

  function setItemPrice(line) {
    const item = items.value.find((candidate) => String(candidate.id) === line.item_id);
    if (item) line.price = item.cost_price || "0";
  }

  async function save() {
    saving.value = true;
    formError.value = "";
    const payload = {
      ...form.value,
      branch_id: Number(form.value.branch_id),
      supplier_id: Number(form.value.supplier_id),
      lines: form.value.lines.map((line) => ({
        item_id: Number(line.item_id),
        qty: Number(line.qty),
        price: String(line.price),
      })),
    };
    try {
      if (selected.value) await purchaseRequestService.update(selected.value.id, payload);
      else await purchaseRequestService.create(payload);
      dialog.value = false;
      message.value = selected.value ? "Purchase request updated." : "Purchase request created.";
      await loadRequests();
    } catch (err) {
      formError.value = err.response?.data?.message || "Failed to save purchase request.";
    } finally {
      saving.value = false;
    }
  }

  async function cancelRequest(row) {
    if (!window.confirm(`Cancel ${row.pr_number}?`)) return;
    try {
      await purchaseRequestService.cancel(row.id);
      message.value = `${row.pr_number} was cancelled.`;
      await loadRequests();
    } catch (err) {
      error.value = err.response?.data?.message || "Failed to cancel purchase request.";
    }
  }

  onMounted(async () => {
    try {
      const [branchResult, supplierResult, itemResult] = await Promise.all([
        purchaseRequestService.branches(),
        purchaseRequestService.suppliers(),
        itemService.list({ page: 1, limit: 100, status: "active" }),
      ]);
      branches.value = branchResult.data.data || [];
      suppliers.value = supplierResult.data.data || [];
      items.value = itemResult.data.data || [];
    } catch (err) {
      error.value = err.response?.data?.message || "Failed to load purchase request references.";
    }
    await loadRequests();
  });

  return {
    requests, branches, suppliers, items, loading, saving, page, pagination, search,
    statusFilter, branchFilter, error, message, dialog, selected, formError, form,
    totalPages, statuses, loadRequests, openCreate, openEdit, addLine, removeLine,
    setItemPrice, save, cancelRequest,
  };
}
