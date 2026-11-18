<script setup>
import { ClipboardList, Eye, Plus, Search, X } from "lucide-vue-next";
import { formatCurrency } from "@/utils/formatters";
import { usePurchaseOrders } from "@/composables/usePurchaseOrders";

const {
  orders, sources, sourceLines, loading, saving, page, pagination, search,
  statusFilter, error, notice, modalError, dialog, detailDialog, selectedOrder,
  selectedSource, poDate, dueDate, totalPages, statuses, formatDateLabel,
  loadOrders, openCreate, selectSource, save, openDetails, cancelOrder,
} = usePurchaseOrders();
</script>

<template>
  <section class="space-y-6">
    <header class="flex flex-col justify-between gap-4 sm:flex-row sm:items-center">
      <div class="flex items-center gap-3">
        <ClipboardList :size="25" class="text-blue-600" aria-hidden="true" />
        <div>
          <h1 class="text-2xl font-bold text-slate-800">Purchase Orders</h1>
          <p class="text-sm text-slate-500">Create orders from available purchase request quantities.</p>
        </div>
      </div>
      <button class="inline-flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2.5 text-sm font-semibold text-white hover:bg-blue-700" @click="openCreate">
        <Plus :size="18" />Create PO
      </button>
    </header>

    <div v-if="notice" role="status" class="rounded-lg border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-800">{{ notice }}</div>
    <div v-if="error" role="alert" class="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">{{ error }}</div>

    <div class="grid gap-3 rounded-xl border border-slate-200 bg-white p-4 shadow-sm md:grid-cols-[minmax(14rem,1fr)_14rem]">
      <label class="relative">
        <span class="sr-only">Search purchase orders</span>
        <Search :size="17" class="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
        <input v-model="search" class="w-full rounded-lg border border-slate-300 py-2.5 pl-9 pr-3 text-sm" placeholder="Search PO number, PR number, supplier" @keydown.enter="page = 1; loadOrders()" />
      </label>
      <select v-model="statusFilter" aria-label="Filter by status" class="rounded-lg border border-slate-300 px-3 py-2.5 text-sm">
        <option value="">All statuses</option>
        <option v-for="(label, value) in statuses" :key="value" :value="value">{{ label }}</option>
      </select>
    </div>

    <div class="overflow-hidden rounded-xl border border-slate-200 bg-white shadow-sm">
      <div class="flex items-center justify-between border-b border-slate-100 px-5 py-4">
        <div><h2 class="font-semibold text-slate-800">Purchase order list</h2><p class="text-xs text-slate-500">{{ pagination.total || 0 }} orders</p></div>
        <button class="text-sm text-blue-700 underline" :disabled="loading" @click="loadOrders">Refresh</button>
      </div>
      <div class="overflow-x-auto">
        <table class="w-full min-w-[1000px] text-left text-sm">
          <thead class="bg-slate-50 text-xs uppercase tracking-wide text-slate-500">
            <tr><th class="px-5 py-3">PO / PR number</th><th class="px-5 py-3">Branch</th><th class="px-5 py-3">Supplier</th><th class="px-5 py-3">PO date</th><th class="px-5 py-3">Due date</th><th class="px-5 py-3">Quantity</th><th class="px-5 py-3">Total</th><th class="px-5 py-3">Status</th><th class="px-5 py-3">Actions</th></tr>
          </thead>
          <tbody class="divide-y divide-slate-100">
            <tr v-if="loading"><td colspan="9" class="px-5 py-12 text-center text-slate-500">Loading purchase orders…</td></tr>
            <tr v-else-if="!orders.length"><td colspan="9" class="px-5 py-12 text-center text-slate-500">No purchase orders found.</td></tr>
            <tr v-for="order in orders" v-else :key="order.id" class="hover:bg-slate-50">
              <td class="px-5 py-4"><p class="font-semibold">{{ order.po_number }}</p><p class="text-xs text-slate-500">PR: {{ order.ref_num }}</p></td>
              <td class="px-5 py-4">{{ order.branch_name }}</td>
              <td class="px-5 py-4">{{ order.supplier_name }}</td>
              <td class="px-5 py-4">{{ formatDateLabel(order.po_date) }}</td>
              <td class="px-5 py-4">{{ formatDateLabel(order.due_date) }}</td>
              <td class="px-5 py-4">{{ order.total_qty }}</td>
              <td class="px-5 py-4">{{ formatCurrency(order.total_price) }}</td>
              <td class="px-5 py-4"><span class="rounded-full bg-blue-50 px-2.5 py-1 text-xs font-medium text-blue-700">{{ statuses[order.status] || order.status }}</span></td>
              <td class="px-5 py-4"><div class="flex gap-3"><button class="text-blue-700 hover:underline" @click="openDetails(order)"><Eye :size="16" class="inline" /> Details</button><button v-if="order.status === 'ordered'" class="text-red-700 hover:underline" @click="cancelOrder(order)">Cancel</button></div></td>
            </tr>
          </tbody>
        </table>
      </div>
      <footer class="flex items-center justify-between border-t border-slate-100 px-5 py-3 text-sm">
        <span class="text-slate-500">Page {{ page }} of {{ totalPages }}</span>
        <div class="flex gap-2"><button class="rounded-md border border-slate-300 px-3 py-1.5 disabled:opacity-40" :disabled="page <= 1 || loading" @click="page--">Previous</button><button class="rounded-md border border-slate-300 px-3 py-1.5 disabled:opacity-40" :disabled="page >= totalPages || loading" @click="page++">Next</button></div>
      </footer>
    </div>

    <div v-if="dialog" class="fixed inset-0 z-50 flex items-center justify-center overflow-y-auto bg-slate-950/50 p-4" @click.self="!saving && (dialog = false)">
      <section role="dialog" aria-modal="true" aria-labelledby="po-create-title" class="my-auto max-h-[92vh] w-full max-w-4xl overflow-y-auto rounded-xl bg-white shadow-xl">
        <header class="sticky top-0 flex items-start justify-between border-b border-slate-200 bg-white px-6 py-5">
          <div><h2 id="po-create-title" class="text-lg font-semibold text-slate-800">Create Purchase Order</h2><p class="mt-1 text-sm text-slate-500">A purchase order can be created from one PR at a time.</p></div>
          <button aria-label="Close" class="rounded-lg p-2 text-slate-500 hover:bg-slate-100" :disabled="saving" @click="dialog = false"><X :size="19" /></button>
        </header>
        <form @submit.prevent="save">
          <div class="space-y-5 p-6">
            <div v-if="modalError" role="alert" class="rounded-lg bg-red-50 p-3 text-sm text-red-700">{{ modalError }}</div>
            <div class="grid gap-4 sm:grid-cols-2">
              <label class="space-y-1.5 text-sm font-medium text-slate-700">Source purchase request <span class="text-red-600">*</span>
                <select required class="w-full rounded-lg border border-slate-300 px-3 py-2.5 font-normal" :value="selectedSource?.pr_number || ''" @change="selectSource($event.target.value)">
                  <option value="" disabled>Select a PR</option>
                  <option v-for="source in sources" :key="source.pr_number" :value="source.pr_number">{{ source.pr_number }} — {{ source.branch_name }} / {{ source.supplier_name }}</option>
                </select>
              </label>
              <div v-if="selectedSource" class="rounded-lg bg-slate-50 p-3 text-sm text-slate-600"><p><strong>Branch:</strong> {{ selectedSource.branch_name }}</p><p><strong>Supplier:</strong> {{ selectedSource.supplier_name }}</p></div>
              <label class="space-y-1.5 text-sm font-medium text-slate-700">PO date <span class="text-red-600">*</span><input v-model="poDate" required type="date" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 font-normal" /></label>
              <label class="space-y-1.5 text-sm font-medium text-slate-700">Due date <span class="text-red-600">*</span><input v-model="dueDate" required type="date" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 font-normal" /></label>
            </div>
            <div v-if="selectedSource" class="space-y-3">
              <h3 class="font-semibold text-slate-800">Remaining PR quantities</h3>
              <div v-for="line in sourceLines" :key="line.item_id" class="grid gap-3 rounded-lg border border-slate-200 p-3 sm:grid-cols-[minmax(12rem,1fr)_7rem_8rem_8rem]">
                <div><p class="font-medium">{{ line.sku }} — {{ line.item_name }} ({{ line.unit }})</p><p class="text-xs text-slate-500">Requested {{ line.requested_qty }} · Already ordered {{ line.ordered_qty }} · Remaining {{ line.remaining_qty }}</p></div>
                <label class="text-xs font-medium text-slate-600">Order qty<input v-model="line.qty" required type="number" min="0.001" :max="line.remaining_qty" step="0.001" class="mt-1 w-full rounded-md border border-slate-300 px-2.5 py-2 text-sm font-normal" /></label>
                <label class="text-xs font-medium text-slate-600">PO unit price<input v-model="line.price" required type="number" min="0" step="0.01" class="mt-1 w-full rounded-md border border-slate-300 px-2.5 py-2 text-sm font-normal" /></label>
                <p class="self-center text-xs text-slate-500">PR price: {{ formatCurrency(line.pr_price) }}</p>
              </div>
              <p v-if="!sourceLines.length" class="rounded-lg bg-amber-50 p-3 text-sm text-amber-800">No remaining quantities are available on this PR.</p>
            </div>
          </div>
          <footer class="sticky bottom-0 flex justify-end gap-3 border-t border-slate-200 bg-white px-6 py-4">
            <button type="button" class="rounded-lg border border-slate-300 px-4 py-2 text-sm" :disabled="saving" @click="dialog = false">Close</button>
            <button type="submit" class="rounded-lg bg-blue-600 px-4 py-2 text-sm font-semibold text-white disabled:opacity-60" :disabled="saving || !sourceLines.length">{{ saving ? 'Saving…' : 'Create PO' }}</button>
          </footer>
        </form>
      </section>
    </div>

    <div v-if="detailDialog && selectedOrder" class="fixed inset-0 z-50 flex items-center justify-center overflow-y-auto bg-slate-950/50 p-4" @click.self="detailDialog = false">
      <section role="dialog" aria-modal="true" aria-labelledby="po-detail-title" class="my-auto max-h-[90vh] w-full max-w-3xl overflow-y-auto rounded-xl bg-white shadow-xl">
        <header class="flex items-start justify-between border-b border-slate-200 px-6 py-5"><div><h2 id="po-detail-title" class="text-lg font-semibold">Purchase Order {{ selectedOrder.po_number }}</h2><p class="text-sm text-slate-500">Source PR {{ selectedOrder.ref_num }}</p></div><button aria-label="Close" class="rounded-lg p-2 text-slate-500 hover:bg-slate-100" @click="detailDialog = false"><X :size="19" /></button></header>
        <div class="space-y-4 p-6"><div class="grid gap-3 text-sm sm:grid-cols-2"><p><strong>Branch:</strong> {{ selectedOrder.branch_name }}</p><p><strong>Supplier:</strong> {{ selectedOrder.supplier_name }}</p><p><strong>PO date:</strong> {{ formatDateLabel(selectedOrder.po_date) }}</p><p><strong>Due date:</strong> {{ formatDateLabel(selectedOrder.due_date) }}</p><p><strong>Status:</strong> {{ statuses[selectedOrder.status] || selectedOrder.status }}</p></div>
          <div v-for="line in selectedOrder.lines" :key="line.id" class="flex justify-between border-t border-slate-100 py-3 text-sm"><span>{{ line.sku }} — {{ line.item_name }} · {{ line.qty }} × {{ formatCurrency(line.price) }}</span><strong>{{ formatCurrency(Number(line.qty) * Number(line.price)) }}</strong></div>
        </div>
      </section>
    </div>
  </section>
</template>
