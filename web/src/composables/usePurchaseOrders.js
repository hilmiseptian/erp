import { computed, onMounted, ref, watch } from "vue";
import purchaseOrderService from "@/services/purchaseOrderService";
import { formatDateLabel, toDateInputValue } from "@/utils/formatters";

function localDateInput() {
  const date = new Date();
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, "0")}-${String(date.getDate()).padStart(2, "0")}`;
}

export function usePurchaseOrders() {
  const orders = ref([]);
  const sources = ref([]);
  const sourceLines = ref([]);
  const loading = ref(false);
  const saving = ref(false);
  const page = ref(1);
  const pagination = ref({ total: 0, total_pages: 0 });
  const search = ref("");
  const statusFilter = ref("");
  const error = ref("");
  const notice = ref("");
  const modalError = ref("");
  const dialog = ref(false);
  const detailDialog = ref(false);
  const selectedOrder = ref(null);
  const selectedSource = ref(null);
  const poDate = ref(localDateInput());
  const dueDate = ref("");
  const totalPages = computed(() => Math.max(1, pagination.value.total_pages || 0));
  const statuses = {
    draft: "Draft",
    ordered: "Ordered",
    partially_received: "Partially received",
    received: "Received",
    cancelled: "Cancelled",
  };

  async function loadOrders() {
    loading.value = true;
    error.value = "";
    try {
      const { data } = await purchaseOrderService.list({
        page: page.value,
        limit: 20,
        search: search.value,
        status: statusFilter.value,
      });
      orders.value = data.data || [];
      pagination.value = data.pagination || { total: 0, total_pages: 0 };
    } catch (err) {
      error.value = err.response?.data?.message || "Failed to load purchase orders.";
    } finally {
      loading.value = false;
    }
  }

  async function loadSources() {
    try {
      const { data } = await purchaseOrderService.sources();
      sources.value = data.data || [];
    } catch (err) {
      error.value = err.response?.data?.message || "Failed to load available purchase requests.";
    }
  }

  watch([statusFilter], () => { page.value = 1; loadOrders(); });
  watch(page, loadOrders);

  async function selectSource(prNumber) {
    selectedSource.value = sources.value.find((source) => source.pr_number === prNumber) || null;
    sourceLines.value = [];
    modalError.value = "";
    if (!prNumber) return;
    try {
      const { data } = await purchaseOrderService.source(prNumber);
      selectedSource.value = data.data;
      sourceLines.value = (data.data.lines || []).map((line) => ({
        ...line,
        qty: line.remaining_qty,
        price: line.price,
        pr_price: line.price,
      }));
      dueDate.value = toDateInputValue(data.data.req_date);
    } catch (err) {
      modalError.value = err.response?.data?.message || "Failed to load purchase request details.";
    }
  }

  async function openCreate() {
    modalError.value = "";
    selectedSource.value = null;
    sourceLines.value = [];
    poDate.value = localDateInput();
    dueDate.value = "";
    dialog.value = true;
    await loadSources();
  }

  async function save() {
    if (!selectedSource.value) {
      modalError.value = "Select a source purchase request.";
      return;
    }
    saving.value = true;
    modalError.value = "";
    try {
      await purchaseOrderService.create({
        ref_num: selectedSource.value.pr_number,
        po_date: poDate.value,
        due_date: dueDate.value,
        lines: sourceLines.value.map((line) => ({
          item_id: Number(line.item_id),
          qty: Number(line.qty),
          price: String(line.price),
        })),
      });
      dialog.value = false;
      notice.value = "Purchase order created.";
      await Promise.all([loadOrders(), loadSources()]);
    } catch (err) {
      modalError.value = err.response?.data?.message || "Failed to create purchase order.";
    } finally {
      saving.value = false;
    }
  }

  async function openDetails(order) {
    try {
      const { data } = await purchaseOrderService.get(order.id);
      selectedOrder.value = data.data;
      detailDialog.value = true;
    } catch (err) {
      error.value = err.response?.data?.message || "Failed to load purchase order details.";
    }
  }

  async function cancelOrder(order) {
    if (!window.confirm(`Cancel purchase order ${order.po_number}?`)) return;
    try {
      await purchaseOrderService.cancel(order.id);
      notice.value = "Purchase order cancelled.";
      await Promise.all([loadOrders(), loadSources()]);
    } catch (err) {
      error.value = err.response?.data?.message || "Failed to cancel purchase order.";
    }
  }

  onMounted(async () => {
    await Promise.all([loadOrders(), loadSources()]);
  });

  return {
    orders, sources, sourceLines, loading, saving, page, pagination, search,
    statusFilter, error, notice, modalError, dialog, detailDialog, selectedOrder,
    selectedSource, poDate, dueDate, totalPages, statuses, formatDateLabel,
    loadOrders, openCreate, selectSource, save, openDetails, cancelOrder,
  };
}
