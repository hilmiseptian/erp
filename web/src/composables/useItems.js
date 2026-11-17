import { computed, onMounted, reactive, ref, watch } from "vue";
import itemService from "@/services/itemService";

const emptyForm = () => ({
  category_id: "",
  name: "",
  sku: "",
  description: "",
  cost_price: "0",
  selling_price: "0",
  min_stock: "0",
  unit: "pcs",
  is_active: true,
});

export function useItems() {
  const items = ref([]);
  const categories = ref([]);
  const loading = ref(false);
  const saving = ref(false);
  const error = ref("");
  const notice = ref("");
  const searchInput = ref("");
  const search = ref("");
  const status = ref("all");
  const categoryId = ref("");
  const page = ref(1);
  const pagination = ref({ page: 1, limit: 20, total: 0, total_pages: 0 });
  const formOpen = ref(false);
  const editing = ref(null);
  const formError = ref("");
  const form = reactive(emptyForm());
  const pageCount = computed(() => Math.max(pagination.value.total_pages, 1));

  async function loadItems() {
    loading.value = true;
    error.value = "";
    try {
      const { data } = await itemService.list({
        page: page.value,
        limit: 20,
        search: search.value,
        status: status.value === "all" ? "" : status.value,
        category_id: categoryId.value || "",
      });
      items.value = data.data || [];
      pagination.value = data.pagination || pagination.value;
    } catch (err) {
      error.value = err.response?.data?.message || "Failed to load items.";
    } finally {
      loading.value = false;
    }
  }

  watch([search, status, categoryId], () => {
    page.value = 1;
    loadItems();
  });
  watch(page, loadItems);

  function openCreate() {
    editing.value = null;
    Object.assign(form, emptyForm(), { category_id: categories.value[0]?.id || "" });
    formError.value = "";
    formOpen.value = true;
  }

  async function openEdit(item) {
    formError.value = "";
    try {
      const { data } = await itemService.get(item.id);
      editing.value = data.data;
      Object.assign(form, {
        ...emptyForm(),
        ...data.data,
        category_id: String(data.data.category_id),
        description: data.data.description || "",
        min_stock: String(data.data.min_stock),
        cost_price: data.data.cost_price,
        selling_price: data.data.selling_price,
      });
      formOpen.value = true;
    } catch (err) {
      error.value = err.response?.data?.message || "Failed to load item details.";
    }
  }

  async function saveItem() {
    saving.value = true;
    formError.value = "";
    const payload = {
      ...form,
      category_id: Number(form.category_id),
      min_stock: Number(form.min_stock),
      description: form.description.trim() || null,
    };
    try {
      if (editing.value) await itemService.update(editing.value.id, payload);
      else await itemService.create(payload);
      notice.value = editing.value ? "Item updated." : "Item created.";
      formOpen.value = false;
      await loadItems();
    } catch (err) {
      formError.value = err.response?.data?.message || "Failed to save item.";
    } finally {
      saving.value = false;
    }
  }

  async function toggleStatus(item) {
    try {
      await itemService.setActive(item.id, !item.is_active);
      notice.value = item.is_active ? "Item deactivated." : "Item activated.";
      await loadItems();
    } catch (err) {
      error.value = err.response?.data?.message || "Failed to update item status.";
    }
  }

  onMounted(async () => {
    try {
      const { data } = await itemService.categories();
      categories.value = data.data || [];
    } catch (err) {
      error.value = err.response?.data?.message || "Failed to load categories.";
    }
    await loadItems();
  });

  return {
    items, categories, loading, saving, error, notice, searchInput, search, status,
    categoryId, page, pagination, formOpen, editing, formError, form, pageCount,
    loadItems, openCreate, openEdit, saveItem, toggleStatus,
  };
}
