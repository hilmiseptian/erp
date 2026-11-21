<script setup>
import { ClipboardList, Eye, Pencil, Plus, Search, XCircle } from 'lucide-vue-next'
import { formatCurrency } from '@/utils/formatters'
import SalesOrderFormDialog from '@/components/sales/SalesOrderFormDialog.vue'
import SalesOrderDetailsDialog from '@/components/sales/SalesOrderDetailsDialog.vue'
import { useSalesOrders } from '@/composables/useSalesOrders'

const formatSalesCurrency = (value) => formatCurrency(value, 'id-ID', 'IDR', 2)

const {
  orders, branches, customers, items, loading, saving, page, pagination, searchInput, statusFilter, customerSearchInput, itemSearchInput,
  error, notice, formError, dialog, detailDialog, form, editing, selectedOrder, newItemId,
  totalPages, statuses, totalQty, totalAmount, grandTotal, formatDateLabel,
  loadOrders, openCreate, openEdit, addItem, removeLine, lineSubtotal, save, openDetails, cancelOrder, applySearch,
} = useSalesOrders()
</script>

<template>
  <section class="space-y-6">
    <header class="flex flex-col justify-between gap-4 sm:flex-row sm:items-center">
      <div class="flex items-center gap-3"><ClipboardList :size="25" class="text-blue-600" aria-hidden="true" /><div><h1 class="text-2xl font-bold text-slate-800">Sales Orders</h1><p class="text-sm text-slate-500">Manage B2B orders. Stock changes when Goods Delivery is posted.</p></div></div>
      <button class="inline-flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2.5 text-sm font-semibold text-white hover:bg-blue-700" @click="openCreate"><Plus :size="18" />Create Sales Order</button>
    </header>

    <div v-if="notice" role="status" class="rounded-lg border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-800">{{ notice }}</div>
    <div v-if="error" role="alert" class="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">{{ error }}</div>

    <form class="grid gap-3 rounded-xl border border-slate-200 bg-white p-4 shadow-sm md:grid-cols-[minmax(14rem,1fr)_14rem_auto]" @submit.prevent="applySearch">
      <label class="relative"><span class="sr-only">Search sales orders</span><Search :size="17" class="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" /><input v-model="searchInput" class="w-full rounded-lg border border-slate-300 py-2.5 pl-9 pr-3 text-sm" placeholder="Search SO number, customer, branch" /></label>
      <label><span class="sr-only">Filter sales order status</span><select v-model="statusFilter" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-sm"><option value="">All statuses</option><option v-for="(label, value) in statuses" :key="value" :value="value">{{ label }}</option></select></label>
      <button class="rounded-lg border border-slate-300 px-4 py-2.5 text-sm font-medium hover:bg-slate-50">Search</button>
    </form>

    <div class="overflow-hidden rounded-xl border border-slate-200 bg-white shadow-sm">
      <header class="flex items-center justify-between border-b border-slate-100 px-5 py-4"><div><h2 class="font-semibold text-slate-800">Sales order list</h2><p class="text-xs text-slate-500">{{ pagination.total || 0 }} order(s)</p></div><button class="text-sm text-blue-700 underline disabled:opacity-50" :disabled="loading" @click="loadOrders">Refresh</button></header>
      <div class="overflow-x-auto"><table class="w-full min-w-[1000px] text-left text-sm"><thead class="bg-slate-50 text-xs uppercase tracking-wide text-slate-500"><tr><th class="px-5 py-3">SO number</th><th class="px-5 py-3">Branch</th><th class="px-5 py-3">Customer</th><th class="px-5 py-3">Order date</th><th class="px-5 py-3">Delivery date</th><th class="px-5 py-3">Quantity</th><th class="px-5 py-3">Grand total</th><th class="px-5 py-3">Status</th><th class="px-5 py-3">Actions</th></tr></thead>
        <tbody class="divide-y divide-slate-100"><tr v-if="loading"><td colspan="9" class="px-5 py-12 text-center text-slate-500">Loading sales orders…</td></tr><tr v-else-if="!orders.length"><td colspan="9" class="px-5 py-12 text-center text-slate-500">No sales orders found.</td></tr><tr v-for="order in orders" v-else :key="order.id" class="hover:bg-slate-50"><td class="px-5 py-4 font-semibold text-slate-800">{{ order.so_number }}</td><td class="px-5 py-4">{{ order.branch_name }}</td><td class="px-5 py-4">{{ order.customer_name }}</td><td class="px-5 py-4">{{ formatDateLabel(order.so_date) }}</td><td class="px-5 py-4">{{ formatDateLabel(order.delivery_date) }}</td><td class="px-5 py-4">{{ order.total_qty }}</td><td class="px-5 py-4">{{ formatSalesCurrency(order.grand_total) }}</td><td class="px-5 py-4"><span class="rounded-full bg-blue-50 px-2.5 py-1 text-xs font-medium text-blue-700">{{ statuses[order.status] || order.status }}</span></td><td class="px-5 py-4"><div class="flex gap-2"><button class="rounded-md p-2 text-blue-700 hover:bg-blue-50" :aria-label="`View ${order.so_number}`" @click="openDetails(order)"><Eye :size="17" /></button><button v-if="order.status === 'open'" class="rounded-md p-2 text-slate-600 hover:bg-slate-100" :aria-label="`Edit ${order.so_number}`" @click="openEdit(order)"><Pencil :size="17" /></button><button v-if="['open', 'partially_delivered'].includes(order.status)" class="rounded-md p-2 text-red-700 hover:bg-red-50" :aria-label="`Cancel ${order.so_number}`" @click="cancelOrder(order)"><XCircle :size="17" /></button></div></td></tr></tbody>
      </table></div>
      <footer class="flex items-center justify-between border-t border-slate-100 px-5 py-3 text-sm"><span class="text-slate-500">Page {{ page }} of {{ totalPages }}</span><div class="flex gap-2"><button class="rounded-md border border-slate-300 px-3 py-1.5 disabled:opacity-40" :disabled="page <= 1 || loading" @click="page--">Previous</button><button class="rounded-md border border-slate-300 px-3 py-1.5 disabled:opacity-40" :disabled="page >= totalPages || loading" @click="page++">Next</button></div></footer>
    </div>
    <SalesOrderFormDialog :open="dialog" :editing="editing" :form="form" :branches="branches" :customers="customers" :items="items" :saving="saving" :error="formError" :total-qty="totalQty" :total-amount="totalAmount" :grand-total="grandTotal" :new-item-id="newItemId" :line-subtotal="lineSubtotal" :customer-search-input="customerSearchInput" :item-search-input="itemSearchInput" @update:new-item-id="newItemId = $event" @update:customer-search-input="customerSearchInput = $event" @update:item-search-input="itemSearchInput = $event" @close="dialog = false" @save="save" @add-item="addItem" @remove-line="removeLine" />
    <SalesOrderDetailsDialog :open="detailDialog" :order="selectedOrder" :statuses="statuses" @close="detailDialog = false" />
  </section>
</template>
