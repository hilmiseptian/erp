import { computed, onMounted, ref, watch } from 'vue'
import stockTransferService from '@/services/stockTransferService'
import { formatDateLabel } from '@/utils/formatters'

function localDateInput() {
  const date = new Date()
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
}

export function useStockTransfers() {
  const transfers = ref([])
  const branches = ref([])
  const items = ref([])
  const selectedTransfer = ref(null)
  const selectedDispatch = ref(null)
  const loading = ref(false)
  const saving = ref(false)
  const page = ref(1)
  const pagination = ref({ total: 0, total_pages: 0 })
  const searchInput = ref('')
  const search = ref('')
  const statusFilter = ref('')
  const error = ref('')
  const notice = ref('')
  const modalError = ref('')
  const transferDialog = ref(false)
  const dispatchDialog = ref(false)
  const receiveDialog = ref(false)
  const detailDialog = ref(false)
  const transferDate = ref(localDateInput())
  const movementDate = ref(localDateInput())
  const transferForm = ref({ from_branch_id: '', to_branch_id: '', notes: '', lines: [] })
  const movementLines = ref([])
  const selectedDispatchID = ref('')
  const editingTransferID = ref(null)
  const totalMovementQty = computed(() => movementLines.value.reduce((total, line) => total + (Number(line.qty) || 0), 0))
  const totalPages = computed(() => Math.max(1, pagination.value.total_pages || 0))

  watch(page, loadTransfers)

  async function loadTransfers() {
    loading.value = true
    error.value = ''
    try {
      const { data } = await stockTransferService.list({ page: page.value, limit: 20, search: search.value, status: statusFilter.value })
      transfers.value = data.data || []
      pagination.value = data.pagination || { total: 0, total_pages: 0 }
    } catch (err) { error.value = err.response?.data?.message || 'Failed to load stock transfers.' }
    finally { loading.value = false }
  }

  async function loadBranches() {
    try {
      const { data } = await stockTransferService.branches()
      branches.value = data.data || []
    } catch (err) { modalError.value = err.response?.data?.message || 'Failed to load branches.' }
  }

  async function loadItems(branchId) {
    if (!branchId) { items.value = []; return }
    try {
      const { data } = await stockTransferService.items(branchId)
      items.value = data.data || []
    } catch (err) { modalError.value = err.response?.data?.message || 'Failed to load source branch stock.' }
  }

  async function openCreate() {
    editingTransferID.value = null
    modalError.value = ''
    transferDate.value = localDateInput()
    transferForm.value = { from_branch_id: '', to_branch_id: '', notes: '', lines: [] }
    items.value = []
    transferDialog.value = true
    await loadBranches()
  }

  async function openEdit(transfer) {
    if (transfer.status !== 'draft') return
    try {
      const { data } = await stockTransferService.get(transfer.id)
      transfer = data.data
    } catch (err) {
      error.value = err.response?.data?.message || 'Failed to load stock transfer draft.'
      return
    }
    editingTransferID.value = transfer.id
    modalError.value = ''
    transferDate.value = transfer.transfer_date
    transferForm.value = {
      from_branch_id: String(transfer.from_branch_id),
      to_branch_id: String(transfer.to_branch_id),
      notes: transfer.notes || '',
      lines: (transfer.lines || []).map((line) => ({ item_id: line.item_id, qty: line.qty })),
    }
    transferDialog.value = true
    await Promise.all([loadBranches(), loadItems(transfer.from_branch_id)])
  }

  async function changeSourceBranch(branchId) {
    transferForm.value.from_branch_id = branchId
    transferForm.value.lines = []
    await loadItems(branchId)
  }

  function addItem(item) {
    if (transferForm.value.lines.some((line) => Number(line.item_id) === Number(item.id))) return
    transferForm.value.lines.push({ item_id: item.id, qty: 1, sku: item.sku, name: item.name, unit: item.unit, stock_qty: item.stock_qty })
  }

  function removeItem(itemId) {
    transferForm.value.lines = transferForm.value.lines.filter((line) => Number(line.item_id) !== Number(itemId))
  }

  async function saveTransfer() {
    if (!transferForm.value.lines.length) { modalError.value = 'Add at least one item.'; return }
    saving.value = true
    modalError.value = ''
    const payload = {
      from_branch_id: Number(transferForm.value.from_branch_id),
      to_branch_id: Number(transferForm.value.to_branch_id),
      transfer_date: transferDate.value,
      notes: transferForm.value.notes || null,
      lines: transferForm.value.lines.map(({ item_id, qty }) => ({ item_id: Number(item_id), qty: Number(qty) })),
    }
    try {
      if (editingTransferID.value) await stockTransferService.update(editingTransferID.value, payload)
      else await stockTransferService.create(payload)
      transferDialog.value = false
      notice.value = editingTransferID.value ? 'Stock transfer draft updated.' : 'Stock transfer draft created.'
      await loadTransfers()
    } catch (err) { modalError.value = err.response?.data?.message || 'Failed to save stock transfer.' }
    finally { saving.value = false }
  }

  async function openDetails(row) {
    try {
      const { data } = await stockTransferService.get(row.id)
      selectedTransfer.value = data.data
      detailDialog.value = true
    } catch (err) { error.value = err.response?.data?.message || 'Failed to load stock transfer details.' }
  }

  async function openDispatch(row) {
    try {
      const { data } = await stockTransferService.get(row.id)
      selectedTransfer.value = data.data
      movementDate.value = localDateInput()
      movementLines.value = data.data.lines.filter((line) => line.remaining_dispatch_qty > 0).map((line) => ({ ...line, qty: '' }))
      modalError.value = ''
      dispatchDialog.value = true
    } catch (err) { error.value = err.response?.data?.message || 'Failed to load remaining transfer quantities.' }
  }

  async function openReceive(row) {
    try {
      const { data } = await stockTransferService.get(row.id)
      selectedTransfer.value = data.data
      selectedDispatch.value = null
      selectedDispatchID.value = ''
      movementLines.value = []
      movementDate.value = localDateInput()
      modalError.value = ''
      receiveDialog.value = true
    } catch (err) { error.value = err.response?.data?.message || 'Failed to load transfer shipments.' }
  }

  function selectDispatch(gdHeaderID) {
    selectedDispatchID.value = gdHeaderID
    selectedDispatch.value = selectedTransfer.value?.dispatches?.find((dispatch) => String(dispatch.gd_header_id) === String(gdHeaderID)) || null
    movementLines.value = (selectedDispatch.value?.lines || []).filter((line) => line.remaining_qty > 0).map((line) => ({ ...line, qty: '' }))
    modalError.value = ''
  }

  async function saveMovement(mode) {
    const lines = movementLines.value.filter((line) => Number(line.qty) > 0).map((line) => ({ item_id: Number(line.item_id), qty: Number(line.qty) }))
    if (!lines.length) { modalError.value = 'Enter a positive quantity for at least one item.'; return }
    saving.value = true
    modalError.value = ''
    try {
      if (mode === 'dispatch') {
        await stockTransferService.dispatch(selectedTransfer.value.id, { date: movementDate.value, lines })
        dispatchDialog.value = false
        notice.value = 'Goods delivery posted; source branch stock reduced.'
      } else {
        await stockTransferService.receive(selectedTransfer.value.id, { gd_header_id: Number(selectedDispatchID.value), date: movementDate.value, lines })
        receiveDialog.value = false
        notice.value = 'Transfer receipt posted; destination branch stock increased.'
      }
      await loadTransfers()
    } catch (err) { modalError.value = err.response?.data?.message || 'Failed to post stock movement.' }
    finally { saving.value = false }
  }

  async function cancelTransfer(transfer) {
    if (!window.confirm(`Cancel draft ${transfer.transfer_number}?`)) return
    try {
      await stockTransferService.cancel(transfer.id)
      notice.value = 'Stock transfer draft cancelled.'
      await loadTransfers()
    } catch (err) { error.value = err.response?.data?.message || 'Failed to cancel stock transfer.' }
  }

  function applySearch() { search.value = searchInput.value.trim(); page.value = 1; loadTransfers() }
  function applyStatus() { page.value = 1; loadTransfers() }
  function statusLabel(status) { return status.replaceAll('_', ' ') }
  function dispatchHasRemaining(dispatch) { return dispatch.lines.some((line) => line.remaining_qty > 0) }
  function statusClass(status) {
    if (status === 'completed') return 'bg-emerald-50 text-emerald-700'
    if (status === 'in_progress') return 'bg-amber-50 text-amber-700'
    if (status === 'cancelled') return 'bg-slate-100 text-slate-600'
    return 'bg-blue-50 text-blue-700'
  }

  onMounted(loadTransfers)

  return {
    transfers, branches, items, selectedTransfer, selectedDispatch, loading, saving, page, pagination,
    searchInput, statusFilter, error, notice, modalError, transferDialog, dispatchDialog, receiveDialog,
    detailDialog, transferDate, movementDate, transferForm, movementLines, selectedDispatchID,
    editingTransferID, totalMovementQty, totalPages, formatDateLabel, statusLabel, statusClass, dispatchHasRemaining,
    loadTransfers, openCreate, openEdit, changeSourceBranch, addItem, removeItem, saveTransfer,
    openDetails, openDispatch, openReceive, selectDispatch, saveMovement, cancelTransfer, applySearch, applyStatus,
  }
}
