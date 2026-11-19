import { computed, onMounted, ref, watch } from "vue";
import goodReceiptService from "@/services/goodReceiptService";
import { formatDateLabel } from "@/utils/formatters";

function localDateInput() {
  const date = new Date();
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, "0")}-${String(date.getDate()).padStart(2, "0")}`;
}

export function useGoodReceipts() {
  const receipts = ref([]);
  const sources = ref([]);
  const sourceLines = ref([]);
  const selectedSource = ref(null);
  const selectedReceipt = ref(null);
  const loading = ref(false);
  const saving = ref(false);
  const detailLoading = ref(false);
  const page = ref(1);
  const pagination = ref({ total: 0, total_pages: 0 });
  const search = ref("");
  const statusFilter = ref("");
  const error = ref("");
  const notice = ref("");
  const modalError = ref("");
  const dialog = ref(false);
  const detailDialog = ref(false);
  const grDate = ref(localDateInput());
  const totalPages = computed(() => Math.max(1, pagination.value.total_pages || 0));

  async function loadReceipts() {
    loading.value = true;
    error.value = "";
    try {
      const { data } = await goodReceiptService.list({
        page: page.value,
        limit: 20,
        search: search.value,
        status: statusFilter.value,
      });
      receipts.value = data.data || [];
      pagination.value = data.pagination || { total: 0, total_pages: 0 };
    } catch (err) {
      error.value = err.response?.data?.message || "Failed to load good receipts.";
    } finally {
      loading.value = false;
    }
  }

  async function loadSources() {
    try {
      const { data } = await goodReceiptService.sources();
      sources.value = data.data || [];
    } catch (err) {
      modalError.value = err.response?.data?.message || "Failed to load purchase orders available for receipt.";
    }
  }

  watch([statusFilter], () => {
    page.value = 1;
    loadReceipts();
  });
  watch(page, loadReceipts);

  async function selectSource(poNumber) {
    selectedSource.value = sources.value.find((source) => source.po_number === poNumber) || null;
    sourceLines.value = [];
    modalError.value = "";
    if (!poNumber) return;

    try {
      const { data } = await goodReceiptService.source(poNumber);
      selectedSource.value = data.data;
      sourceLines.value = (data.data.lines || []).map((line) => ({ ...line, qty: "" }));
    } catch (err) {
      modalError.value = err.response?.data?.message || "Failed to load remaining purchase order quantities.";
    }
  }

  async function openCreate() {
    modalError.value = "";
    selectedSource.value = null;
    sourceLines.value = [];
    grDate.value = localDateInput();
    dialog.value = true;
    await loadSources();
  }

  async function save() {
    const lines = sourceLines.value
      .filter((line) => Number(line.qty) > 0)
      .map((line) => ({ item_id: Number(line.item_id), qty: Number(line.qty) }));
    if (!selectedSource.value) {
      modalError.value = "Select a purchase order.";
      return;
    }
    if (!lines.length) {
      modalError.value = "Enter a receipt quantity for at least one item.";
      return;
    }

    saving.value = true;
    modalError.value = "";
    try {
      await goodReceiptService.create({
        po_number: selectedSource.value.po_number,
        gr_date: grDate.value,
        lines,
      });
      dialog.value = false;
      notice.value = "Good receipt posted and branch stock updated.";
      await Promise.all([loadReceipts(), loadSources()]);
    } catch (err) {
      modalError.value = err.response?.data?.message || "Failed to post good receipt.";
    } finally {
      saving.value = false;
    }
  }

  async function openDetails(receipt) {
    detailLoading.value = true;
    try {
      const { data } = await goodReceiptService.get(receipt.id);
      selectedReceipt.value = data.data;
      detailDialog.value = true;
    } catch (err) {
      error.value = err.response?.data?.message || "Failed to load good receipt details.";
    } finally {
      detailLoading.value = false;
    }
  }

  onMounted(loadReceipts);

  return {
    receipts, sources, sourceLines, selectedSource, selectedReceipt,
    loading, saving, detailLoading, page, pagination, search, statusFilter,
    error, notice, modalError, dialog, detailDialog, grDate, totalPages,
    formatDateLabel, loadReceipts, openCreate, selectSource, save, openDetails,
  };
}
