import { computed, onMounted, ref, watch } from 'vue'
import goodDeliveryService from '@/services/goodDeliveryService'
import { formatDateLabel } from '@/utils/formatters'

function localDateInput() {
  const date = new Date()
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
}

export function useGoodDeliveries() {
  const deliveries = ref([])
  const sources = ref([])
  const sourceLines = ref([])
  const selectedSource = ref(null)
  const selectedDelivery = ref(null)
  const loading = ref(false)
  const saving = ref(false)
  const page = ref(1)
  const pagination = ref({ total: 0, total_pages: 0 })
  const searchInput = ref('')
  const search = ref('')
  const error = ref('')
  const notice = ref('')
  const modalError = ref('')
  const dialog = ref(false)
  const detailDialog = ref(false)
  const gdDate = ref(localDateInput())
  const totalPages = computed(() => Math.max(1, pagination.value.total_pages || 0))
  const totalQty = computed(() => sourceLines.value.reduce((sum, line) => sum + (Number(line.qty) || 0), 0))

  watch(page, loadDeliveries)

  async function loadDeliveries() {
    loading.value = true
    error.value = ''
    try {
      const { data } = await goodDeliveryService.list({ page: page.value, limit: 20, search: search.value })
      deliveries.value = data.data || []
      pagination.value = data.pagination || { total: 0, total_pages: 0 }
    } catch (err) { error.value = err.response?.data?.message || 'Failed to load goods deliveries.' }
    finally { loading.value = false }
  }

  async function loadSources() {
    try {
      const { data } = await goodDeliveryService.sources('sales_order')
      sources.value = data.data || []
    } catch (err) { modalError.value = err.response?.data?.message || 'Failed to load open sales orders.' }
  }

  async function selectSource(refNum) {
    selectedSource.value = sources.value.find((source) => source.ref_num === refNum) || null
    sourceLines.value = []
    modalError.value = ''
    if (!refNum) return
    try {
      const { data } = await goodDeliveryService.source('sales_order', refNum)
      selectedSource.value = data.data
      sourceLines.value = (data.data.lines || []).map((line) => ({ ...line, qty: '' }))
    } catch (err) { modalError.value = err.response?.data?.message || 'Failed to load remaining sales order quantities.' }
  }

  async function openCreate() {
    modalError.value = ''
    selectedSource.value = null
    sourceLines.value = []
    sources.value = []
    gdDate.value = localDateInput()
    dialog.value = true
    await loadSources()
  }

  async function save() {
    const lines = sourceLines.value.filter((line) => Number(line.qty) > 0)
      .map((line) => ({ item_id: Number(line.item_id), qty: Number(line.qty) }))
    if (!selectedSource.value) { modalError.value = 'Select a source sales order.'; return }
    if (!lines.length) { modalError.value = 'Enter a delivery quantity for at least one item.'; return }
    saving.value = true
    modalError.value = ''
    try {
      await goodDeliveryService.create({ ref_type: 'sales_order', ref_num: selectedSource.value.ref_num, gd_date: gdDate.value, lines })
      dialog.value = false
      notice.value = 'Goods delivery posted and branch stock updated.'
      await loadDeliveries()
    } catch (err) { modalError.value = err.response?.data?.message || 'Failed to post goods delivery.' }
    finally { saving.value = false }
  }

  async function openDetails(delivery) {
    try {
      const { data } = await goodDeliveryService.get(delivery.id)
      selectedDelivery.value = data.data
      detailDialog.value = true
    } catch (err) { error.value = err.response?.data?.message || 'Failed to load goods delivery details.' }
  }

  function applySearch() { search.value = searchInput.value.trim(); page.value = 1; loadDeliveries() }

  onMounted(loadDeliveries)

  return {
    deliveries, sources, sourceLines, selectedSource, selectedDelivery, loading, saving,
    page, pagination, searchInput, error, notice, modalError, dialog, detailDialog,
    gdDate, totalPages, totalQty, formatDateLabel,
    loadDeliveries, openCreate, selectSource, save, openDetails, applySearch,
  }
}
