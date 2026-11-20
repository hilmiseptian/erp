import { computed, onMounted, reactive, ref, watch } from "vue";
import stockAdjustmentService from "@/services/stockAdjustmentService";
import stockService from "@/services/stockService";
import { formatDateLabel, toDateInputValue } from "@/utils/formatters";

function localDateInput() {
  const date = new Date();
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, "0")}-${String(date.getDate()).padStart(2, "0")}`;
}

function emptyForm(branchId = "") {
  return {
    branch_id: branchId,
    adjustment_date: localDateInput(),
    adjustment_type: "adjustment",
    reason: "",
    lines: [],
  };
}

export function useStockAdjustments() {
  const adjustments = ref([]);
  const branches = ref([]);
  const availableItems = ref([]);
  const selectedItemId = ref("");
  const itemSearch = ref("");
  const loading = ref(false);
  const saving = ref(false);
  const detailLoading = ref(false);
  const actionId = ref(null);
  const page = ref(1);
  const pagination = ref({ total: 0, total_pages: 0 });
  const branchFilter = ref("");
  const statusFilter = ref("");
  const typeFilter = ref("");
  const search = ref("");
  const error = ref("");
  const notice = ref("");
  const formError = ref("");
  const formOpen = ref(false);
  const detailDialog = ref(false);
  const editing = ref(null);
  const selectedAdjustment = ref(null);
  const form = reactive(emptyForm());
  const totalPages = computed(() => Math.max(1, pagination.value.total_pages || 0));
  const selectableItems = computed(() => {
    const inForm = new Set(form.lines.map((line) => Number(line.item_id)));
    return availableItems.value.filter((item) => !inForm.has(Number(item.item_id)));
  });

  async function loadAdjustments() {
    loading.value = true;
    error.value = "";
    try {
      const { data } = await stockAdjustmentService.list({
        page: page.value,
        limit: 20,
        branch_id: branchFilter.value || "",
        status: statusFilter.value,
        adjustment_type: typeFilter.value,
        search: search.value.trim(),
      });
      adjustments.value = data.data || [];
      pagination.value = data.pagination || { total: 0, total_pages: 0 };
    } catch (err) {
      error.value = err.response?.data?.message || "Failed to load stock adjustments.";
    } finally {
      loading.value = false;
    }
  }

  async function loadAvailableItems() {
    if (!form.branch_id) {
      availableItems.value = [];
      return;
    }
    formError.value = "";
    try {
      const { data } = await stockService.list({
        branch_id: form.branch_id,
        search: itemSearch.value.trim(),
        page: 1,
        limit: 100,
      });
      availableItems.value = data.data || [];
    } catch (err) {
      formError.value = err.response?.data?.message || "Failed to load branch items.";
    }
  }

  watch([branchFilter, statusFilter, typeFilter], () => {
    if (page.value !== 1) page.value = 1;
    else loadAdjustments();
  });
  watch(page, loadAdjustments);

  function submitSearch() {
    if (page.value !== 1) page.value = 1;
    else loadAdjustments();
  }

  function openCreate() {
    notice.value = "";
    editing.value = null;
    Object.assign(form, emptyForm(branchFilter.value || branches.value[0]?.id?.toString() || ""));
    form.lines = [];
    selectedItemId.value = "";
    itemSearch.value = "";
    formError.value = "";
    formOpen.value = true;
    loadAvailableItems();
  }

  async function openEdit(adjustment) {
    notice.value = "";
    formError.value = "";
    try {
      const { data } = await stockAdjustmentService.get(adjustment.id);
      const value = data.data;
      editing.value = value;
      Object.assign(form, {
        branch_id: String(value.branch_id),
        adjustment_date: toDateInputValue(value.adjustment_date),
        adjustment_type: value.adjustment_type,
        reason: value.reason || "",
        lines: (value.lines || []).map((line) => ({
          ...line,
          item_id: String(line.item_id),
          actual_qty: String(line.actual_qty),
        })),
      });
      selectedItemId.value = "";
      itemSearch.value = "";
      formOpen.value = true;
      await loadAvailableItems();
    } catch (err) {
      error.value = err.response?.data?.message || "Failed to load stock adjustment details.";
    }
  }

  function addLine() {
    const item = availableItems.value.find((option) => String(option.item_id) === selectedItemId.value);
    if (!item) return;
    form.lines.push({
      item_id: String(item.item_id),
      sku: item.sku,
      item_name: item.item_name,
      unit: item.unit,
      system_qty: Number(item.qty || 0),
      actual_qty: form.adjustment_type === "opening_balance" ? "0" : String(item.qty || 0),
    });
    selectedItemId.value = "";
  }

  function removeLine(index) {
    form.lines.splice(index, 1);
  }

  async function save() {
    formError.value = "";
    if (!form.lines.length) {
      formError.value = "Add at least one item.";
      return;
    }
    if (form.lines.some((line) => !Number.isInteger(Number(line.actual_qty)) || Number(line.actual_qty) < 0)) {
      formError.value = "Actual quantities must be whole numbers of zero or more.";
      return;
    }

    saving.value = true;
    try {
      const payload = {
        branch_id: Number(form.branch_id),
        adjustment_date: form.adjustment_date,
        adjustment_type: form.adjustment_type,
        reason: form.reason.trim() || null,
        lines: form.lines.map((line) => ({ item_id: Number(line.item_id), actual_qty: Number(line.actual_qty) })),
      };
      if (editing.value) {
        await stockAdjustmentService.update(editing.value.id, payload);
        notice.value = "Stock adjustment draft updated.";
      } else {
        await stockAdjustmentService.create(payload);
        notice.value = "Stock adjustment draft created.";
      }
      formOpen.value = false;
      await loadAdjustments();
    } catch (err) {
      formError.value = err.response?.data?.message || "Failed to save stock adjustment draft.";
    } finally {
      saving.value = false;
    }
  }

  async function openDetails(adjustment) {
    detailLoading.value = true;
    try {
      const { data } = await stockAdjustmentService.get(adjustment.id);
      selectedAdjustment.value = data.data;
      detailDialog.value = true;
    } catch (err) {
      error.value = err.response?.data?.message || "Failed to load stock adjustment details.";
    } finally {
      detailLoading.value = false;
    }
  }

  async function post(adjustment) {
    actionId.value = adjustment.id;
    error.value = "";
    try {
      await stockAdjustmentService.post(adjustment.id);
      notice.value = "Stock adjustment posted and branch stock updated.";
      await loadAdjustments();
      if (selectedAdjustment.value?.id === adjustment.id) await openDetails(adjustment);
    } catch (err) {
      error.value = err.response?.data?.message || "Failed to post stock adjustment.";
    } finally {
      actionId.value = null;
    }
  }

  async function cancel(adjustment) {
    actionId.value = adjustment.id;
    error.value = "";
    try {
      await stockAdjustmentService.cancel(adjustment.id);
      notice.value = "Stock adjustment draft cancelled.";
      await loadAdjustments();
      if (selectedAdjustment.value?.id === adjustment.id) await openDetails({ ...adjustment, status: "cancelled" });
    } catch (err) {
      error.value = err.response?.data?.message || "Failed to cancel stock adjustment.";
    } finally {
      actionId.value = null;
    }
  }

  onMounted(async () => {
    try {
      const { data } = await stockService.branches();
      branches.value = data.data || [];
    } catch (err) {
      error.value = err.response?.data?.message || "Failed to load branches.";
    }
    await loadAdjustments();
  });

  return {
    adjustments, branches, availableItems, selectableItems, selectedItemId, itemSearch,
    loading, saving, detailLoading, actionId, page, pagination, branchFilter, statusFilter,
    typeFilter, search, error, notice, formError, formOpen, detailDialog, editing,
    selectedAdjustment, form, totalPages, formatDateLabel, loadAdjustments,
    loadAvailableItems, openCreate, openEdit, addLine, removeLine, save, submitSearch,
    openDetails, post, cancel,
  };
}
