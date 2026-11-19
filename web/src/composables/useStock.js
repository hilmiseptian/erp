import { computed, onMounted, ref, watch } from "vue";
import stockService from "@/services/stockService";
import { formatDateLabel } from "@/utils/formatters";

export function useStock() {
  const branches = ref([]);
  const branchId = ref("");
  const rows = ref([]);
  const cards = ref([]);
  const cardPage = ref(1);
  const cardPagination = ref({ total: 0, total_pages: 0 });
  const loading = ref(false);
  const cardsLoading = ref(false);
  const error = ref("");
  const search = ref("");
  const page = ref(1);
  const pagination = ref({ total: 0, total_pages: 0 });
  const selectedItem = ref(null);
  const cardDialog = ref(false);
  const totalPages = computed(() => Math.max(1, pagination.value.total_pages || 0));
  const cardTotalPages = computed(() => Math.max(1, cardPagination.value.total_pages || 0));

  async function loadStock() {
    if (!branchId.value) { rows.value = []; return; }
    loading.value = true;
    error.value = "";
    try {
      const { data } = await stockService.list({ branch_id: branchId.value, search: search.value, page: page.value, limit: 20 });
      rows.value = data.data || [];
      pagination.value = data.pagination || { total: 0, total_pages: 0 };
    } catch (err) {
      error.value = err.response?.data?.message || "Failed to load stock balances.";
    } finally { loading.value = false; }
  }

  watch(branchId, () => { page.value = 1; loadStock(); });
  watch(page, loadStock);

  async function loadCards() {
    if (!selectedItem.value || !branchId.value) return;
    cardsLoading.value = true;
    try {
      const { data } = await stockService.cards(branchId.value, selectedItem.value.item_id, { page: cardPage.value, limit: 20 });
      cards.value = data.data || [];
      cardPagination.value = data.pagination || { total: 0, total_pages: 0 };
    } catch (err) {
      error.value = err.response?.data?.message || "Failed to load stock card.";
      cardDialog.value = false;
    } finally { cardsLoading.value = false; }
  }

  async function openCards(item) {
    selectedItem.value = item;
    cardDialog.value = true;
    if (cardPage.value === 1) await loadCards();
    else cardPage.value = 1;
  }

  watch(cardPage, loadCards);

  onMounted(async () => {
    try {
      const { data } = await stockService.branches();
      branches.value = data.data || [];
      if (branches.value.length) branchId.value = String(branches.value[0].id);
    } catch (err) {
      error.value = err.response?.data?.message || "Failed to load branches.";
    }
  });

  return { branches, branchId, rows, cards, loading, cardsLoading, error, search, page, pagination, selectedItem, cardDialog, cardPage, cardPagination, cardTotalPages, totalPages, formatDateLabel, loadStock, loadCards, openCards };
}
