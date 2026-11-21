import { computed, onMounted, reactive, ref, watch } from 'vue'
import customerService from '@/services/customerService'

const emptyForm = () => ({
  customer_code: '', name: '', tax_number: '', contact_person: '', phone: '', email: '',
  billing_address: '', shipping_address: '',
})

export function useCustomers() {
  const customers = ref([])
  const loading = ref(false)
  const saving = ref(false)
  const error = ref('')
  const notice = ref('')
  const searchInput = ref('')
  const search = ref('')
  const status = ref('all')
  const page = ref(1)
  const pagination = ref({ page: 1, limit: 20, total: 0, total_pages: 0 })
  const dialog = ref(false)
  const editing = ref(null)
  const formError = ref('')
  const form = reactive(emptyForm())
  const pageCount = computed(() => Math.max(pagination.value.total_pages, 1))

  async function load() {
    loading.value = true
    error.value = ''
    try {
      const { data } = await customerService.list({ page: page.value, limit: 20, search: search.value, status: status.value === 'all' ? '' : status.value })
      customers.value = data.data || []
      pagination.value = data.pagination || pagination.value
    } catch (err) {
      error.value = err.response?.data?.message || 'Failed to load customers.'
    } finally { loading.value = false }
  }

  watch([search, status], () => { page.value = 1; load() })
  watch(page, load)

  function openCreate() {
    editing.value = null
    Object.assign(form, emptyForm())
    formError.value = ''
    dialog.value = true
  }

  async function openEdit(customer) {
    formError.value = ''
    try {
      const { data } = await customerService.get(customer.id)
      editing.value = data.data
      Object.assign(form, Object.fromEntries(Object.keys(emptyForm()).map((key) => [key, data.data[key] || ''])))
      dialog.value = true
    } catch (err) { error.value = err.response?.data?.message || 'Failed to load customer details.' }
  }

  async function save() {
    saving.value = true
    formError.value = ''
    const payload = Object.fromEntries(Object.entries(form).map(([key, value]) => [key, typeof value === 'string' ? value.trim() || null : value]))
    payload.customer_code = form.customer_code.trim()
    payload.name = form.name.trim()
    try {
      if (editing.value) await customerService.update(editing.value.id, payload)
      else await customerService.create(payload)
      notice.value = editing.value ? 'Customer updated.' : 'Customer created.'
      dialog.value = false
      await load()
    } catch (err) { formError.value = err.response?.data?.message || 'Failed to save customer.' }
    finally { saving.value = false }
  }

  async function toggleActive(customer) {
    try {
      await customerService.setActive(customer.id, !customer.is_active)
      notice.value = customer.is_active ? 'Customer deactivated.' : 'Customer activated.'
      await load()
    } catch (err) { error.value = err.response?.data?.message || 'Failed to update customer status.' }
  }

  function applySearch() { search.value = searchInput.value.trim() }
  onMounted(load)

  return { customers, loading, saving, error, notice, searchInput, search, status, page, pagination, pageCount, dialog, editing, formError, form, load, openCreate, openEdit, save, toggleActive, applySearch }
}
