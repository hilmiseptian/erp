<script setup>
import { ClipboardCheck, Eye, Plus, Search, X } from "lucide-vue-next";
import { formatCurrency } from "@/utils/formatters";
import { useGoodReceipts } from "@/composables/useGoodReceipts";

const {
  receipts, sources, sourceLines, selectedSource, selectedReceipt, loading,
  saving, page, pagination, search, statusFilter, error, notice, modalError,
  dialog, detailDialog, grDate, totalPages, formatDateLabel, loadReceipts,
  openCreate, selectSource, save, openDetails,
} = useGoodReceipts();
</script>

<template>
  <section class="space-y-6">
    <header class="flex flex-col justify-between gap-4 sm:flex-row sm:items-center">
      <div class="flex items-center gap-3">
        <ClipboardCheck :size="25" class="text-blue-600" aria-hidden="true" />
        <div>
          <h1 class="text-2xl font-bold text-slate-800">Good Receipts</h1>
          <p class="text-sm text-slate-500">Receive all or part of an open purchase order into branch stock.</p>
        </div>
      </div>
      <button class="inline-flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2.5 text-sm font-semibold text-white hover:bg-blue-700" @click="openCreate">
        <Plus :size="18" />Receive items
      </button>
    </header>

    <div v-if="notice" role="status" class="rounded-lg border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-800">{{ notice }}</div>
    <div v-if="error" role="alert" class="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">{{ error }}</div>

    <div class="grid gap-3 rounded-xl border border-slate-200 bg-white p-4 shadow-sm md:grid-cols-[minmax(14rem,1fr)_14rem]">
      <label class="relative">
        <span class="sr-only">Search good receipts</span>
        <Search :size="17" class="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
        <input v-model="search" class="w-full rounded-lg border border-slate-300 py-2.5 pl-9 pr-3 text-sm" placeholder="Search GR, PO, supplier, branch" @keydown.enter="page = 1; loadReceipts()" />
      </label>
      <select v-model="statusFilter" aria-label="Filter by status" class="rounded-lg border border-slate-300 px-3 py-2.5 text-sm">
        <option value="">All statuses</option>
        <option value="posted">Posted</option>
      </select>
    </div>

    <div class="overflow-hidden rounded-xl border border-slate-200 bg-white shadow-sm">
      <div class="flex items-center justify-between border-b border-slate-100 px-5 py-4">
        <div><h2 class="font-semibold text-slate-800">Good receipt list</h2><p class="text-xs text-slate-500">{{ pagination.total || 0 }} receipts</p></div>
        <button class="text-sm text-blue-700 underline" :disabled="loading" @click="loadReceipts">Refresh</button>
      </div>
      <div class="overflow-x-auto">
        <table class="w-full min-w-[900px] text-left text-sm">
          <thead class="bg-slate-50 text-xs uppercase tracking-wide text-slate-500"><tr><th class="px-5 py-3">GR number</th><th class="px-5 py-3">PO number</th><th class="px-5 py-3">Branch</th><th class="px-5 py-3">Supplier</th><th class="px-5 py-3">GR date</th><th class="px-5 py-3">Quantity</th><th class="px-5 py-3">Total</th><th class="px-5 py-3">Status</th><th class="px-5 py-3">Action</th></tr></thead>
          <tbody class="divide-y divide-slate-100">
            <tr v-if="loading"><td colspan="9" class="px-5 py-12 text-center text-slate-500">Loading good receipts…</td></tr>
            <tr v-else-if="!receipts.length"><td colspan="9" class="px-5 py-12 text-center text-slate-500">No good receipts found.</td></tr>
            <tr v-for="receipt in receipts" v-else :key="receipt.id" class="hover:bg-slate-50">
              <td class="px-5 py-4 font-semibold">{{ receipt.gr_number }}</td><td class="px-5 py-4">{{ receipt.po_number }}</td><td class="px-5 py-4">{{ receipt.branch_name }}</td><td class="px-5 py-4">{{ receipt.supplier_name }}</td><td class="px-5 py-4">{{ formatDateLabel(receipt.gr_date) }}</td><td class="px-5 py-4">{{ receipt.total_qty }}</td><td class="px-5 py-4">{{ formatCurrency(receipt.total_price) }}</td><td class="px-5 py-4"><span class="rounded-full bg-emerald-50 px-2.5 py-1 text-xs font-medium text-emerald-700">Posted</span></td><td class="px-5 py-4"><button class="inline-flex items-center gap-1 text-blue-700 hover:underline" :disabled="detailLoading" @click="openDetails(receipt)"><Eye :size="16" />Details</button></td>
            </tr>
          </tbody>
        </table>
      </div>
      <footer class="flex items-center justify-between border-t border-slate-100 px-5 py-3 text-sm"><span class="text-slate-500">Page {{ page }} of {{ totalPages }}</span><div class="flex gap-2"><button class="rounded-md border border-slate-300 px-3 py-1.5 disabled:opacity-40" :disabled="page <= 1 || loading" @click="page--">Previous</button><button class="rounded-md border border-slate-300 px-3 py-1.5 disabled:opacity-40" :disabled="page >= totalPages || loading" @click="page++">Next</button></div></footer>
    </div>

    <div v-if="dialog" class="fixed inset-0 z-50 flex items-center justify-center overflow-y-auto bg-slate-950/50 p-4" @click.self="!saving && (dialog = false)">
      <section role="dialog" aria-modal="true" aria-labelledby="gr-create-title" class="my-auto max-h-[92vh] w-full max-w-4xl overflow-y-auto rounded-xl bg-white shadow-xl">
        <header class="sticky top-0 flex items-start justify-between border-b border-slate-200 bg-white px-6 py-5"><div><h2 id="gr-create-title" class="text-lg font-semibold text-slate-800">Receive purchase order</h2><p class="mt-1 text-sm text-slate-500">Posting this receipt immediately adds the received quantities to branch stock.</p></div><button aria-label="Close" class="rounded-lg p-2 text-slate-500 hover:bg-slate-100" :disabled="saving" @click="dialog = false"><X :size="19" /></button></header>
        <form @submit.prevent="save">
          <div class="space-y-5 p-6">
            <div v-if="modalError" role="alert" class="rounded-lg bg-red-50 p-3 text-sm text-red-700">{{ modalError }}</div>
            <div class="grid gap-4 sm:grid-cols-2">
              <label class="space-y-1.5 text-sm font-medium text-slate-700">Purchase order <span class="text-red-600">*</span><select required class="w-full rounded-lg border border-slate-300 px-3 py-2.5 font-normal" :value="selectedSource?.po_number || ''" @change="selectSource($event.target.value)"><option value="" disabled>Select a PO</option><option v-for="source in sources" :key="source.po_number" :value="source.po_number">{{ source.po_number }} — {{ source.branch_name }} / {{ source.supplier_name }}</option></select></label>
              <label class="space-y-1.5 text-sm font-medium text-slate-700">GR date <span class="text-red-600">*</span><input v-model="grDate" required type="date" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 font-normal" /></label>
              <div v-if="selectedSource" class="rounded-lg bg-slate-50 p-3 text-sm text-slate-600 sm:col-span-2"><p><strong>Branch:</strong> {{ selectedSource.branch_name }}</p><p><strong>Supplier:</strong> {{ selectedSource.supplier_name }}</p></div>
            </div>
            <div v-if="selectedSource" class="space-y-3">
              <h3 class="font-semibold text-slate-800">Remaining quantities</h3>
              <div v-for="line in sourceLines" :key="line.item_id" class="grid gap-3 rounded-lg border border-slate-200 p-3 sm:grid-cols-[minmax(12rem,1fr)_8rem_9rem]">
                <div><p class="font-medium">{{ line.sku }} — {{ line.item_name }} ({{ line.unit }})</p><p class="text-xs text-slate-500">Ordered {{ line.ordered_qty }} · Received {{ line.received_qty }} · Remaining {{ line.remaining_qty }}</p></div>
                <label class="text-xs font-medium text-slate-600">Receipt qty<input v-model="line.qty" type="number" min="0" :max="line.remaining_qty" step="1" placeholder="0" class="mt-1 w-full rounded-md border border-slate-300 px-2.5 py-2 text-sm font-normal" /></label>
                <div class="self-center text-sm"><p class="text-xs text-slate-500">PO unit price</p><strong>{{ formatCurrency(line.price) }}</strong></div>
              </div>
              <p v-if="!sourceLines.length" class="rounded-lg bg-amber-50 p-3 text-sm text-amber-800">No remaining quantities are available on this purchase order.</p>
            </div>
          </div>
          <footer class="sticky bottom-0 flex justify-end gap-3 border-t border-slate-200 bg-white px-6 py-4"><button type="button" class="rounded-lg border border-slate-300 px-4 py-2 text-sm" :disabled="saving" @click="dialog = false">Close</button><button type="submit" class="rounded-lg bg-blue-600 px-4 py-2 text-sm font-semibold text-white disabled:opacity-60" :disabled="saving || !sourceLines.length">{{ saving ? 'Posting…' : 'Post receipt' }}</button></footer>
        </form>
      </section>
    </div>

    <div v-if="detailDialog && selectedReceipt" class="fixed inset-0 z-50 flex items-center justify-center overflow-y-auto bg-slate-950/50 p-4" @click.self="detailDialog = false">
      <section role="dialog" aria-modal="true" aria-labelledby="gr-detail-title" class="my-auto max-h-[90vh] w-full max-w-3xl overflow-y-auto rounded-xl bg-white shadow-xl">
        <header class="flex items-start justify-between border-b border-slate-200 px-6 py-5"><div><h2 id="gr-detail-title" class="text-lg font-semibold">Good Receipt {{ selectedReceipt.gr_number }}</h2><p class="text-sm text-slate-500">Purchase order {{ selectedReceipt.po_number }}</p></div><button aria-label="Close" class="rounded-lg p-2 text-slate-500 hover:bg-slate-100" @click="detailDialog = false"><X :size="19" /></button></header>
        <div class="space-y-4 p-6"><div class="grid gap-3 text-sm sm:grid-cols-2"><p><strong>Branch:</strong> {{ selectedReceipt.branch_name }}</p><p><strong>Supplier:</strong> {{ selectedReceipt.supplier_name }}</p><p><strong>GR date:</strong> {{ formatDateLabel(selectedReceipt.gr_date) }}</p><p><strong>Status:</strong> Posted</p></div><div class="overflow-x-auto"><table class="w-full min-w-[550px] text-left text-sm"><thead class="border-y border-slate-100 text-xs uppercase text-slate-500"><tr><th class="py-3">Item</th><th class="py-3">Quantity</th><th class="py-3">PO price</th><th class="py-3">Total</th></tr></thead><tbody class="divide-y divide-slate-100"><tr v-for="line in selectedReceipt.lines" :key="line.id"><td class="py-3">{{ line.sku }} — {{ line.item_name }}</td><td class="py-3">{{ line.qty }}</td><td class="py-3">{{ formatCurrency(line.price) }}</td><td class="py-3">{{ formatCurrency(Number(line.qty) * Number(line.price)) }}</td></tr></tbody></table></div></div>
      </section>
    </div>
  </section>
</template>
