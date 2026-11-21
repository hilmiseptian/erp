import { computed, onMounted, reactive, ref, watch } from 'vue'
import customerService from '@/services/customerService'
import itemService from '@/services/itemService'
import salesOrderService from '@/services/salesOrderService'
import { formatDateLabel, toDateInputValue } from '@/utils/formatters'

function localDateInput() {
  const date = new Date()
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
}

const emptyForm = () => ({ branch_id: '', customer_id: '', so_date: localDateInput(), delivery_date: '', delivery_address: '', discount_amount: '0.00', notes: '', lines: [] })

export function useSalesOrders() {
  const orders = ref([])
  const branches = ref([])
  const customers = ref([])
  const items = ref([])
  const loading = ref(false)
  const saving = ref(false)
  const page = ref(1)
  const pagination = ref({ total: 0, total_pages: 0 })
  const searchInput = ref('')
  const search = ref('')
  const customerSearchInput = ref('')
  const itemSearchInput = ref('')
  const statusFilter = ref('')
  const error = ref('')
  const notice = ref('')
  const formError = ref('')
  const dialog = ref(false)
  const detailDialog = ref(false)
  const form = reactive(emptyForm())
  const editing = ref(null)
  const selectedOrder = ref(null)
  const newItemId = ref('')
  const totalPages = computed(() => Math.max(1, pagination.value.total_pages || 0))
  const statuses = { open: 'Open', partially_delivered: 'Partially delivered', delivered: 'Delivered', cancelled: 'Cancelled' }
  const totalQty = computed(() => form.lines.reduce((sum, line) => sum + Number(line.qty || 0), 0))
  const totalAmount = computed(() => form.lines.reduce((sum, line) => {
    const gross = Number(line.unit_price || 0) * Number(line.qty || 0)
    return sum + Math.max(gross - Number(line.discount_amount || 0), 0)
  }, 0))
  const grandTotal = computed(() => Math.max(totalAmount.value - Number(form.discount_amount || 0), 0))

  watch(statusFilter, () => { page.value = 1; loadOrders() })
  watch(page, loadOrders)
  watch(() => form.customer_id, (value) => {
    const customer = customers.value.find((item) => String(item.id) === String(value))
    if (customer && !form.delivery_address) form.delivery_address = customer.shipping_address || ''
  })

  let customerSearchTimer
  let itemSearchTimer
  let customerSearchRequest = 0
  let itemSearchRequest = 0
  watch(customerSearchInput, () => {
    clearTimeout(customerSearchTimer)
    customerSearchTimer = setTimeout(() => loadCustomers(customerSearchInput.value), 250)
  })
  watch(itemSearchInput, () => {
    clearTimeout(itemSearchTimer)
    newItemId.value = ''
    itemSearchTimer = setTimeout(() => loadItems(itemSearchInput.value), 250)
  })

  async function loadOrders() {
    loading.value = true
    error.value = ''
    try {
      const { data } = await salesOrderService.list({ page: page.value, limit: 20, search: search.value, status: statusFilter.value })
      orders.value = data.data || []
      pagination.value = data.pagination || { total: 0, total_pages: 0 }
    } catch (err) { error.value = err.response?.data?.message || 'Failed to load sales orders.' }
    finally { loading.value = false }
  }

  async function loadOptions() {
    try {
      const [branchResult] = await Promise.all([salesOrderService.branches(), loadCustomers(''), loadItems('')])
      branches.value = branchResult.data.data || []
    } catch (err) { error.value = err.response?.data?.message || 'Failed to load sales order options.' }
  }

  async function loadCustomers(query) {
    const request = ++customerSearchRequest
    try {
      const { data } = await customerService.list({ page: 1, limit: 100, status: 'active', search: query })
      if (request !== customerSearchRequest) return
      const results = data.data || []
      const selected = customers.value.find((customer) => String(customer.id) === String(form.customer_id))
      if (selected && !results.some((customer) => customer.id === selected.id)) results.unshift(selected)
      customers.value = results
    } catch (err) { error.value = err.response?.data?.message || 'Failed to search customers.' }
  }

  async function loadItems(query) {
    const request = ++itemSearchRequest
    try {
      const { data } = await itemService.list({ page: 1, limit: 100, status: 'active', search: query })
      if (request !== itemSearchRequest) return
      items.value = data.data || []
    } catch (err) { error.value = err.response?.data?.message || 'Failed to search items.' }
  }

  function resetForm() {
    Object.assign(form, emptyForm(), { branch_id: branches.value[0] ? String(branches.value[0].id) : '' })
    customerSearchInput.value = ''
    itemSearchInput.value = ''
    newItemId.value = ''
  }

  function openCreate() { editing.value = null; formError.value = ''; resetForm(); dialog.value = true }

  async function openEdit(order) {
    formError.value = ''
    try {
      const { data } = await salesOrderService.get(order.id)
      const value = data.data
      const { data: customerData } = await customerService.get(value.customer_id)
      if (!customers.value.some((customer) => customer.id === customerData.data.id)) customers.value.unshift(customerData.data)
      editing.value = value
      customerSearchInput.value = ''
      itemSearchInput.value = ''
      Object.assign(form, {
        branch_id: String(value.branch_id), customer_id: String(value.customer_id), so_date: toDateInputValue(value.so_date),
        delivery_date: toDateInputValue(value.delivery_date), delivery_address: value.delivery_address || '',
        discount_amount: value.discount_amount || '0.00', notes: value.notes || '',
        lines: (value.lines || []).map((line) => ({ ...line, qty: String(line.qty), discount_amount: line.discount_amount || '0.00', notes: line.notes || '' })),
      })
      dialog.value = true
    } catch (err) { error.value = err.response?.data?.message || 'Failed to load sales order details.' }
  }

  function addItem() {
    const item = items.value.find((candidate) => String(candidate.id) === String(newItemId.value))
    if (!item || form.lines.some((line) => Number(line.item_id) === Number(item.id))) return
    form.lines.push({ id: 0, item_id: item.id, item_name: item.name, sku: item.sku, unit: item.unit, qty: '1', unit_price: item.selling_price, discount_amount: '0.00', notes: '' })
    newItemId.value = ''
  }

  function removeLine(index) { form.lines.splice(index, 1) }
  function lineSubtotal(line) { return Math.max(Number(line.unit_price || 0) * Number(line.qty || 0) - Number(line.discount_amount || 0), 0) }

  async function save() {
    saving.value = true
    formError.value = ''
    const payload = {
      branch_id: Number(form.branch_id), customer_id: Number(form.customer_id), so_date: form.so_date,
      delivery_date: form.delivery_date || null, delivery_address: form.delivery_address.trim() || null,
      discount_amount: String(form.discount_amount || '0'), notes: form.notes.trim() || null,
      lines: form.lines.map((line) => ({ id: Number(line.id || 0), item_id: Number(line.item_id), qty: Number(line.qty), discount_amount: String(line.discount_amount || '0'), notes: line.notes?.trim() || null })),
    }
    try {
      if (editing.value) await salesOrderService.update(editing.value.id, payload)
      else await salesOrderService.create(payload)
      dialog.value = false
      notice.value = editing.value ? 'Sales order updated.' : 'Sales order created.'
      await loadOrders()
    } catch (err) { formError.value = err.response?.data?.message || 'Failed to save sales order.' }
    finally { saving.value = false }
  }

  async function openDetails(order) {
    try {
      const { data } = await salesOrderService.get(order.id)
      selectedOrder.value = data.data
      detailDialog.value = true
    } catch (err) { error.value = err.response?.data?.message || 'Failed to load sales order details.' }
  }

  async function cancelOrder(order) {
    if (!window.confirm(`Cancel sales order ${order.so_number}?`)) return
    try { await salesOrderService.cancel(order.id); notice.value = 'Sales order cancelled.'; await loadOrders() }
    catch (err) { error.value = err.response?.data?.message || 'Failed to cancel sales order.' }
  }

  function applySearch() { search.value = searchInput.value.trim(); page.value = 1; loadOrders() }
  onMounted(async () => { await Promise.all([loadOrders(), loadOptions()]) })

  return {
    orders, branches, customers, items, loading, saving, page, pagination, searchInput, statusFilter, customerSearchInput, itemSearchInput,
    error, notice, formError, dialog, detailDialog, form, editing, selectedOrder, newItemId,
    totalPages, statuses, totalQty, totalAmount, grandTotal, formatDateLabel,
    loadOrders, openCreate, openEdit, addItem, removeLine, lineSubtotal, save, openDetails, cancelOrder, applySearch,
  }
}
