<script setup>
import { computed, nextTick, ref, watch } from 'vue'
import { X, Plus, Trash2 } from 'lucide-vue-next'
import { formatCurrency } from '@/utils/formatters'

const formatSalesCurrency = (value) => formatCurrency(value, 'id-ID', 'IDR', 2)

const props = defineProps({ open: Boolean, editing: Object, form: Object, branches: Array, customers: Array, items: Array, saving: Boolean, error: String, totalQty: Number, totalAmount: Number, grandTotal: Number, newItemId: [String, Number], lineSubtotal: Function, customerSearchInput: String, itemSearchInput: String })
defineEmits(['close', 'save', 'add-item', 'remove-line', 'update:new-item-id'])
const availableItems = computed(() => props.items.filter((candidate) => !props.form.lines.some((line) => Number(line.item_id) === Number(candidate.id))))
const dialogElement = ref(null)
watch(() => props.open, async (isOpen) => {
  await nextTick()
  if (isOpen && dialogElement.value && !dialogElement.value.open) dialogElement.value.showModal()
  if (!isOpen && dialogElement.value?.open) dialogElement.value.close()
}, { immediate: true })
</script>

<template>
  <dialog ref="dialogElement" v-show="open" class="fixed inset-0 z-50 m-0 flex h-full w-full items-center justify-center overflow-y-auto border-0 bg-slate-950/50 p-4 backdrop:bg-transparent" aria-modal="true" aria-labelledby="sales-order-dialog-title" @click.self="!saving && $emit('close')" @cancel.prevent="!saving && $emit('close')">
    <section class="my-auto max-h-[94vh] w-full max-w-5xl overflow-y-auto rounded-xl bg-white shadow-xl">
      <header class="sticky top-0 z-10 flex items-start justify-between border-b border-slate-200 bg-white px-6 py-5"><div><h2 id="sales-order-dialog-title" class="text-lg font-semibold text-slate-800">{{ editing ? `Edit Sales Order ${editing.so_number}` : 'Create Sales Order' }}</h2><p class="mt-1 text-sm text-slate-500">Prices come from the item master. Sales orders do not change stock.</p></div><button aria-label="Close" class="rounded-lg p-2 text-slate-500 hover:bg-slate-100" :disabled="saving" @click="$emit('close')"><X :size="19" /></button></header>
      <form @submit.prevent="$emit('save')">
        <div class="space-y-5 p-6">
          <div v-if="error" role="alert" class="rounded-lg bg-red-50 p-3 text-sm text-red-700">{{ error }}</div>
          <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
            <label class="space-y-1.5 text-sm font-medium text-slate-700">Branch <span class="text-red-600">*</span><select v-model="form.branch_id" required class="w-full rounded-lg border border-slate-300 px-3 py-2.5 font-normal"><option value="" disabled>Select branch</option><option v-for="branch in branches" :key="branch.id" :value="String(branch.id)">{{ branch.name }}</option></select></label>
            <div class="space-y-1.5 text-sm font-medium text-slate-700"><label for="sales-order-customer-search">Customer <span class="text-red-600">*</span></label><input id="sales-order-customer-search" :value="customerSearchInput" type="search" placeholder="Search customers" class="mb-1 w-full rounded-lg border border-slate-300 px-3 py-2 text-sm font-normal" @input="$emit('update:customer-search-input', $event.target.value)" /><select v-model="form.customer_id" required aria-label="Select customer" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 font-normal"><option value="" disabled>Select customer</option><option v-for="customer in customers" :key="customer.id" :value="String(customer.id)">{{ customer.customer_code }} — {{ customer.name }}</option></select></div>
            <label class="space-y-1.5 text-sm font-medium text-slate-700">Order date <span class="text-red-600">*</span><input v-model="form.so_date" required type="date" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 font-normal" /></label>
            <label class="space-y-1.5 text-sm font-medium text-slate-700">Delivery date<input v-model="form.delivery_date" type="date" :min="form.so_date" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 font-normal" /></label>
            <label class="space-y-1.5 text-sm font-medium text-slate-700 sm:col-span-2">Delivery address<textarea v-model="form.delivery_address" rows="2" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 font-normal"></textarea></label>
          </div>
          <div class="space-y-3">
            <div class="flex flex-col justify-between gap-3 sm:flex-row sm:items-end"><div><h3 class="font-semibold text-slate-800">Items</h3><p class="text-xs text-slate-500">Enter whole-number quantities. Unit prices are read-only.</p></div><div class="grid gap-2 sm:grid-cols-[14rem_14rem_auto]"><label class="sr-only" for="sales-order-item-search">Search items</label><input id="sales-order-item-search" :value="itemSearchInput" type="search" placeholder="Search items" class="rounded-lg border border-slate-300 px-3 py-2 text-sm" @input="$emit('update:item-search-input', $event.target.value)" /><label class="sr-only" for="sales-order-item">Add item</label><select id="sales-order-item" :value="newItemId" class="rounded-lg border border-slate-300 px-3 py-2 text-sm" @change="$emit('update:new-item-id', $event.target.value)"><option value="">Select item to add</option><option v-for="item in availableItems" :key="item.id" :value="String(item.id)">{{ item.sku }} — {{ item.name }}</option></select><button type="button" class="inline-flex items-center justify-center gap-1 rounded-lg border border-slate-300 px-3 py-2 text-sm disabled:opacity-50" :disabled="!newItemId" @click="$emit('add-item')"><Plus :size="16" />Add</button></div></div>
            <div v-for="(line, index) in form.lines" :key="line.id || `new-${line.item_id}`" class="grid gap-3 rounded-lg border border-slate-200 p-3 sm:grid-cols-[minmax(10rem,1fr)_6rem_8rem_8rem_9rem_auto] sm:items-end">
              <div class="min-w-0"><p class="truncate font-medium text-slate-800">{{ line.sku }} — {{ line.item_name }}</p><p class="text-xs text-slate-500">{{ line.unit }}</p></div>
              <label class="text-xs font-medium text-slate-600">Quantity<input v-model="line.qty" required type="number" min="1" step="1" class="mt-1 w-full rounded-md border border-slate-300 px-2 py-2 text-sm font-normal" /></label>
              <label class="text-xs font-medium text-slate-600">Unit price<input :value="line.unit_price" readonly aria-label="Unit price from item master" class="mt-1 w-full rounded-md border border-slate-200 bg-slate-50 px-2 py-2 text-sm font-normal text-slate-600" /></label>
              <label class="text-xs font-medium text-slate-600">Line discount<input v-model="line.discount_amount" type="number" min="0" step="0.01" class="mt-1 w-full rounded-md border border-slate-300 px-2 py-2 text-sm font-normal" /></label>
              <p class="pb-2 text-sm font-semibold text-slate-700">{{ formatSalesCurrency(lineSubtotal(line)) }}</p>
              <button type="button" :aria-label="`Remove ${line.item_name}`" class="mb-1 rounded-md p-2 text-slate-500 hover:bg-red-50 hover:text-red-700" @click="$emit('remove-line', index)"><Trash2 :size="17" /></button>
            </div>
            <p v-if="!form.lines.length" class="rounded-lg border border-dashed border-slate-300 p-6 text-center text-sm text-slate-500">Select an item to add it to this order.</p>
          </div>
          <div class="grid gap-4 sm:grid-cols-2">
            <label class="space-y-1.5 text-sm font-medium text-slate-700">Order discount<input v-model="form.discount_amount" type="number" min="0" step="0.01" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 font-normal" /></label>
            <label class="space-y-1.5 text-sm font-medium text-slate-700">Notes<textarea v-model="form.notes" rows="2" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 font-normal"></textarea></label>
          </div>
          <div class="ml-auto grid max-w-md grid-cols-2 gap-x-4 gap-y-2 rounded-lg bg-slate-50 p-4 text-sm"><span class="text-slate-500">Total quantity</span><strong class="text-right">{{ totalQty }}</strong><span class="text-slate-500">Amount after line discounts</span><strong class="text-right">{{ formatSalesCurrency(totalAmount) }}</strong><span class="text-slate-500">Order discount</span><strong class="text-right">{{ formatSalesCurrency(form.discount_amount) }}</strong><span class="font-semibold">Grand total</span><strong class="text-right text-base text-slate-900">{{ formatSalesCurrency(grandTotal) }}</strong></div>
        </div>
        <footer class="sticky bottom-0 flex justify-end gap-3 border-t border-slate-200 bg-white px-6 py-4"><button type="button" class="rounded-lg border border-slate-300 px-4 py-2 text-sm" :disabled="saving" @click="$emit('close')">Close</button><button type="submit" class="rounded-lg bg-blue-600 px-4 py-2 text-sm font-semibold text-white disabled:opacity-60" :disabled="saving || !form.lines.length">{{ saving ? 'Saving…' : editing ? 'Save Changes' : 'Create Sales Order' }}</button></footer>
      </form>
    </section>
  </dialog>
</template>
