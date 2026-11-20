<script setup>
import { nextTick, ref, watch } from "vue";
import { Ban, Boxes, Eye, Pencil, Plus, Search, Send, Trash2, X } from "lucide-vue-next";
import { useStockAdjustments } from "@/composables/useStockAdjustments";

const {
  adjustments, branches, availableItems, selectableItems, selectedItemId, itemSearch,
  loading, saving, detailLoading, actionId, page, pagination, branchFilter,
  statusFilter, typeFilter, search, error, notice, formError, formOpen, detailDialog,
  editing, selectedAdjustment, form, totalPages, formatDateLabel, loadAdjustments,
  loadAvailableItems, openCreate, openEdit, addLine, removeLine, save, submitSearch,
  openDetails, post, cancel,
} = useStockAdjustments();

const editorDialog = ref(null);
const detailsElement = ref(null);

watch(formOpen, async (open) => {
  await nextTick();
  if (open && editorDialog.value && !editorDialog.value.open) editorDialog.value.showModal();
  if (!open && editorDialog.value?.open) editorDialog.value.close();
});

watch(detailDialog, async (open) => {
  await nextTick();
  if (open && detailsElement.value && !detailsElement.value.open) detailsElement.value.showModal();
  if (!open && detailsElement.value?.open) detailsElement.value.close();
});

function closeEditor() {
  if (!saving.value) formOpen.value = false;
}

function confirmPost(adjustment) {
  if (window.confirm(`Post ${adjustment.adjustment_number}? This will update branch stock.`)) post(adjustment);
}

function confirmCancel(adjustment) {
  if (window.confirm(`Cancel draft ${adjustment.adjustment_number}?`)) cancel(adjustment);
}

function typeLabel(type) {
  return type === "opening_balance" ? "Opening balance" : "Stock adjustment";
}

function difference(line) {
  return Number(line.actual_qty || 0) - Number(line.system_qty || 0);
}
</script>

<template>
  <section class="space-y-6">
    <header class="flex flex-col justify-between gap-4 sm:flex-row sm:items-center">
      <div class="flex items-center gap-3">
        <Boxes :size="25" class="text-blue-600" aria-hidden="true" />
        <div>
          <h1 class="text-2xl font-bold text-slate-800">Stock Adjustments</h1>
          <p class="text-sm text-slate-500">Record opening balances and physical stock count differences.</p>
        </div>
      </div>
      <button type="button" class="inline-flex min-h-11 items-center justify-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-semibold text-white hover:bg-blue-700" @click="openCreate">
        <Plus :size="17" aria-hidden="true" />New adjustment
      </button>
    </header>

    <div v-if="error" role="alert" class="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">{{ error }}</div>
    <div v-if="notice" role="status" class="rounded-lg border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-800">{{ notice }}</div>

    <div class="grid gap-3 rounded-xl border border-slate-200 bg-white p-4 shadow-sm md:grid-cols-5">
      <search class="md:col-span-2">
        <form action="/inventory/stock-adjustments" method="get" class="flex gap-2" @submit.prevent="submitSearch">
          <label for="stock-adjustment-search" class="sr-only">Search stock adjustments</label>
          <div class="relative min-w-0 flex-1">
            <Search :size="17" class="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" aria-hidden="true" />
            <input id="stock-adjustment-search" v-model="search" name="search" type="search" class="min-h-11 w-full rounded-lg border border-slate-300 py-2.5 pl-9 pr-3 text-sm" placeholder="Search number, branch, or reason" />
          </div>
          <button type="submit" class="min-h-11 rounded-lg border border-slate-300 px-4 text-sm font-medium text-slate-700 hover:bg-slate-50">Search</button>
        </form>
      </search>
      <label for="stock-adjustment-branch-filter" class="sr-only">Filter by branch</label>
      <select id="stock-adjustment-branch-filter" v-model="branchFilter" name="branch_id" class="min-h-11 rounded-lg border border-slate-300 px-3 py-2.5 text-sm">
        <option value="">All branches</option>
        <option v-for="branch in branches" :key="branch.id" :value="String(branch.id)">{{ branch.name }}</option>
      </select>
      <label for="stock-adjustment-status-filter" class="sr-only">Filter by status</label>
      <select id="stock-adjustment-status-filter" v-model="statusFilter" name="status" class="min-h-11 rounded-lg border border-slate-300 px-3 py-2.5 text-sm">
        <option value="">All statuses</option><option value="draft">Draft</option><option value="posted">Posted</option><option value="cancelled">Cancelled</option>
      </select>
      <label for="stock-adjustment-type-filter" class="sr-only">Filter by type</label>
      <select id="stock-adjustment-type-filter" v-model="typeFilter" name="adjustment_type" class="min-h-11 rounded-lg border border-slate-300 px-3 py-2.5 text-sm">
        <option value="">All types</option><option value="adjustment">Stock adjustment</option><option value="opening_balance">Opening balance</option>
      </select>
    </div>

    <div class="overflow-hidden rounded-xl border border-slate-200 bg-white shadow-sm">
      <div class="flex items-center justify-between border-b border-slate-100 px-5 py-4">
        <div><h2 class="font-semibold text-slate-800">Adjustment documents</h2><p class="text-xs text-slate-500">{{ pagination.total || 0 }} documents</p></div>
        <button type="button" class="text-sm text-blue-700 underline disabled:opacity-50" :disabled="loading" @click="loadAdjustments">Refresh</button>
      </div>
      <div class="overflow-x-auto">
        <table class="w-full min-w-[950px] text-left text-sm">
          <thead class="bg-slate-50 text-xs uppercase tracking-wide text-slate-500"><tr><th class="px-5 py-3">Document</th><th class="px-5 py-3">Branch</th><th class="px-5 py-3">Date</th><th class="px-5 py-3">Type</th><th class="px-5 py-3">Difference qty</th><th class="px-5 py-3">Status</th><th class="px-5 py-3">Actions</th></tr></thead>
          <tbody class="divide-y divide-slate-100">
            <tr v-if="loading"><td colspan="7" class="px-5 py-12 text-center text-slate-500">Loading stock adjustments…</td></tr>
            <tr v-else-if="!adjustments.length"><td colspan="7" class="px-5 py-12 text-center text-slate-500">No stock adjustment documents found.</td></tr>
            <tr v-for="adjustment in adjustments" v-else :key="adjustment.id">
              <td class="px-5 py-4"><p class="font-medium text-slate-800">{{ adjustment.adjustment_number }}</p><p class="text-xs text-slate-500">{{ adjustment.reason || "—" }}</p></td>
              <td class="px-5 py-4">{{ adjustment.branch_name }}</td>
              <td class="px-5 py-4">{{ formatDateLabel(adjustment.adjustment_date) }}</td>
              <td class="px-5 py-4">{{ typeLabel(adjustment.adjustment_type) }}</td>
              <td class="px-5 py-4">{{ adjustment.total_qty }}</td>
              <td class="px-5 py-4"><span class="rounded-full px-2.5 py-1 text-xs font-medium" :class="adjustment.status === 'posted' ? 'bg-emerald-100 text-emerald-800' : adjustment.status === 'cancelled' ? 'bg-slate-100 text-slate-600' : 'bg-amber-100 text-amber-800'">{{ adjustment.status }}</span></td>
              <td class="px-5 py-4"><div class="flex items-center gap-3">
                <button type="button" class="inline-flex min-h-9 items-center gap-1 text-blue-700 hover:underline disabled:opacity-50" :disabled="detailLoading" @click="openDetails(adjustment)"><Eye :size="16" aria-hidden="true" />Details</button>
                <template v-if="adjustment.status === 'draft'">
                  <button type="button" class="inline-flex min-h-9 items-center gap-1 text-slate-700 hover:underline" @click="openEdit(adjustment)"><Pencil :size="15" aria-hidden="true" />Edit</button>
                  <button type="button" class="inline-flex min-h-9 items-center gap-1 text-emerald-700 hover:underline disabled:opacity-50" :disabled="actionId === adjustment.id" @click="confirmPost(adjustment)"><Send :size="15" aria-hidden="true" />Post</button>
                  <button type="button" class="inline-flex min-h-9 items-center gap-1 text-red-700 hover:underline disabled:opacity-50" :disabled="actionId === adjustment.id" @click="confirmCancel(adjustment)"><Ban :size="15" aria-hidden="true" />Cancel</button>
                </template>
              </div></td>
            </tr>
          </tbody>
        </table>
      </div>
      <footer class="flex items-center justify-between border-t border-slate-100 px-5 py-3 text-sm"><span class="text-slate-500">Page {{ page }} of {{ totalPages }}</span><div class="flex gap-2"><button type="button" class="min-h-9 rounded-md border border-slate-300 px-3 py-1.5 disabled:opacity-40" :disabled="page <= 1 || loading" @click="page--">Previous</button><button type="button" class="min-h-9 rounded-md border border-slate-300 px-3 py-1.5 disabled:opacity-40" :disabled="page >= totalPages || loading" @click="page++">Next</button></div></footer>
    </div>

    <dialog ref="editorDialog" aria-labelledby="stock-adjustment-form-title" class="m-auto max-h-[92vh] w-full max-w-4xl overflow-y-auto rounded-xl border-0 bg-white p-0 text-slate-800 shadow-xl backdrop:bg-slate-950/50" @click.self="closeEditor" @close="formOpen = false" @cancel.prevent="closeEditor">
      <header class="sticky top-0 z-10 flex items-start justify-between border-b border-slate-200 bg-white px-6 py-5">
        <div><h2 id="stock-adjustment-form-title" class="text-lg font-semibold">{{ editing ? 'Edit stock adjustment draft' : 'New stock adjustment' }}</h2><p class="mt-1 text-sm text-slate-500">Stock changes only after the draft is posted.</p></div>
        <button type="button" aria-label="Close stock adjustment form" class="rounded-lg p-2 text-slate-500 hover:bg-slate-100 disabled:opacity-50" :disabled="saving" @click="closeEditor"><X :size="19" aria-hidden="true" /></button>
      </header>
      <form action="/inventory/stock-adjustments" method="post" aria-labelledby="stock-adjustment-form-title" @submit.prevent="save">
        <div class="space-y-5 p-6">
          <div v-if="formError" role="alert" class="rounded-lg border border-red-200 bg-red-50 p-3 text-sm text-red-700">{{ formError }}</div>
          <div class="grid gap-4 sm:grid-cols-2">
            <label for="stock-adjustment-form-branch" class="space-y-1.5 text-sm font-medium text-slate-700">Branch <span class="text-red-600" aria-hidden="true">*</span>
              <select id="stock-adjustment-form-branch" v-model="form.branch_id" name="branch_id" required class="min-h-11 w-full rounded-lg border border-slate-300 px-3 py-2.5 font-normal" :disabled="Boolean(editing)" @change="loadAvailableItems">
                <option value="" disabled>Select a branch</option><option v-for="branch in branches" :key="branch.id" :value="String(branch.id)">{{ branch.name }}</option>
              </select>
            </label>
            <label for="stock-adjustment-date" class="space-y-1.5 text-sm font-medium text-slate-700">Adjustment date <span class="text-red-600" aria-hidden="true">*</span>
              <input id="stock-adjustment-date" v-model="form.adjustment_date" name="adjustment_date" type="date" required class="min-h-11 w-full rounded-lg border border-slate-300 px-3 py-2.5 font-normal" />
            </label>
            <fieldset class="space-y-2 sm:col-span-2">
              <legend class="mb-2 text-sm font-medium text-slate-700">Adjustment type <span class="text-red-600" aria-hidden="true">*</span></legend>
              <div class="flex flex-col gap-3 sm:flex-row">
                <label class="flex min-h-11 items-center gap-2 rounded-lg border border-slate-300 px-3 py-2 text-sm"><input v-model="form.adjustment_type" name="adjustment_type" type="radio" value="adjustment" required /><span><strong>Stock adjustment</strong><span class="block text-xs text-slate-500">Correct a physical count difference.</span></span></label>
                <label class="flex min-h-11 items-center gap-2 rounded-lg border border-slate-300 px-3 py-2 text-sm"><input v-model="form.adjustment_type" name="adjustment_type" type="radio" value="opening_balance" required /><span><strong>Opening balance</strong><span class="block text-xs text-slate-500">Record the initial quantity for each item once per branch.</span></span></label>
              </div>
            </fieldset>
            <label for="stock-adjustment-reason" class="space-y-1.5 text-sm font-medium text-slate-700 sm:col-span-2">Reason
              <textarea id="stock-adjustment-reason" v-model="form.reason" name="reason" maxlength="255" rows="2" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 font-normal" placeholder="Optional note about this adjustment"></textarea>
            </label>
          </div>

          <fieldset class="space-y-3">
            <legend class="font-semibold text-slate-800">Items</legend>
            <div class="grid gap-2 sm:grid-cols-[minmax(0,1fr)_minmax(12rem,1fr)_auto]">
              <label for="stock-adjustment-item-search" class="sr-only">Search branch items</label>
              <input id="stock-adjustment-item-search" v-model="itemSearch" name="item_search" type="search" class="min-h-11 rounded-lg border border-slate-300 px-3 py-2.5 text-sm" placeholder="Search item name or SKU" @keydown.enter.prevent="loadAvailableItems" />
              <label for="stock-adjustment-item-select" class="sr-only">Item to add</label>
              <select id="stock-adjustment-item-select" v-model="selectedItemId" name="item_id" class="min-h-11 rounded-lg border border-slate-300 px-3 py-2.5 text-sm">
                <option value="">Select an item</option><option v-for="item in selectableItems" :key="item.item_id" :value="String(item.item_id)">{{ item.sku }} — {{ item.item_name }} ({{ item.qty }} {{ item.unit }})</option>
              </select>
              <button type="button" class="inline-flex min-h-11 items-center justify-center gap-2 rounded-lg border border-slate-300 px-4 text-sm font-medium text-slate-700 hover:bg-slate-50" @click="loadAvailableItems">Search items</button>
              <button type="button" class="inline-flex min-h-11 items-center justify-center gap-2 rounded-lg bg-slate-800 px-4 text-sm font-medium text-white hover:bg-slate-700 disabled:opacity-50 sm:col-start-3" :disabled="!selectedItemId" @click="addLine"><Plus :size="16" aria-hidden="true" />Add item</button>
            </div>
            <p class="text-xs text-slate-500">The system quantity is captured when the draft is saved. If stock changes before posting, update the draft to refresh it.</p>
            <div v-if="!form.lines.length" class="rounded-lg border border-dashed border-slate-300 px-4 py-8 text-center text-sm text-slate-500">Add at least one item to this document.</div>
            <div v-else class="overflow-x-auto rounded-lg border border-slate-200">
              <table class="w-full min-w-[700px] text-left text-sm">
                <thead class="bg-slate-50 text-xs uppercase text-slate-500"><tr><th class="px-3 py-3">Item</th><th class="px-3 py-3">System qty</th><th class="px-3 py-3">Actual qty</th><th class="px-3 py-3">Difference</th><th class="px-3 py-3"><span class="sr-only">Remove item</span></th></tr></thead>
                <tbody class="divide-y divide-slate-100">
                  <tr v-for="(line, index) in form.lines" :key="line.item_id">
                    <td class="px-3 py-3"><p class="font-medium">{{ line.sku }} — {{ line.item_name }}</p><p class="text-xs text-slate-500">{{ line.unit }}</p></td>
                    <td class="px-3 py-3">{{ line.system_qty }}</td>
                    <td class="px-3 py-3"><label :for="`actual-qty-${line.item_id}`" class="sr-only">Actual quantity for {{ line.item_name }}</label><input :id="`actual-qty-${line.item_id}`" v-model="line.actual_qty" name="actual_qty" type="number" inputmode="numeric" min="0" step="1" required class="min-h-10 w-28 rounded-md border border-slate-300 px-2.5 py-2" /></td>
                    <td class="px-3 py-3 font-semibold" :class="difference(line) < 0 ? 'text-red-700' : difference(line) > 0 ? 'text-emerald-700' : 'text-slate-500'">{{ difference(line) > 0 ? `+${difference(line)}` : difference(line) }}</td>
                    <td class="px-3 py-3"><button type="button" class="rounded-md p-2 text-red-700 hover:bg-red-50" :aria-label="`Remove ${line.item_name}`" @click="removeLine(index)"><Trash2 :size="16" aria-hidden="true" /></button></td>
                  </tr>
                </tbody>
              </table>
            </div>
          </fieldset>
        </div>
        <footer class="sticky bottom-0 flex justify-end gap-3 border-t border-slate-200 bg-white px-6 py-4">
          <button type="button" class="min-h-11 rounded-lg border border-slate-300 px-4 py-2 text-sm" :disabled="saving" @click="closeEditor">Close</button>
          <button type="submit" class="min-h-11 rounded-lg bg-blue-600 px-4 py-2 text-sm font-semibold text-white disabled:opacity-60" :disabled="saving">{{ saving ? 'Saving…' : 'Save draft' }}</button>
        </footer>
      </form>
    </dialog>

    <dialog ref="detailsElement" aria-labelledby="stock-adjustment-detail-title" class="m-auto max-h-[90vh] w-full max-w-3xl overflow-y-auto rounded-xl border-0 bg-white p-0 text-slate-800 shadow-xl backdrop:bg-slate-950/50" @click.self="detailDialog = false" @close="detailDialog = false">
      <template v-if="selectedAdjustment">
        <header class="flex items-start justify-between border-b border-slate-200 px-6 py-5">
          <div><h2 id="stock-adjustment-detail-title" class="text-lg font-semibold">{{ selectedAdjustment.adjustment_number }}</h2><p class="text-sm text-slate-500">{{ typeLabel(selectedAdjustment.adjustment_type) }}</p></div>
          <button type="button" aria-label="Close stock adjustment details" class="rounded-lg p-2 text-slate-500 hover:bg-slate-100" @click="detailDialog = false"><X :size="19" aria-hidden="true" /></button>
        </header>
        <div class="space-y-5 p-6">
          <div class="grid gap-3 text-sm sm:grid-cols-2"><p><strong>Branch:</strong> {{ selectedAdjustment.branch_name }}</p><p><strong>Date:</strong> {{ formatDateLabel(selectedAdjustment.adjustment_date) }}</p><p><strong>Status:</strong> {{ selectedAdjustment.status }}</p><p><strong>Total difference quantity:</strong> {{ selectedAdjustment.total_qty }}</p><p class="sm:col-span-2"><strong>Reason:</strong> {{ selectedAdjustment.reason || "—" }}</p></div>
          <div class="overflow-x-auto rounded-lg border border-slate-200">
            <table class="w-full min-w-[600px] text-left text-sm"><thead class="bg-slate-50 text-xs uppercase text-slate-500"><tr><th class="px-4 py-3">Item</th><th class="px-4 py-3">System qty</th><th class="px-4 py-3">Actual qty</th><th class="px-4 py-3">Difference</th></tr></thead>
              <tbody class="divide-y divide-slate-100"><tr v-for="line in selectedAdjustment.lines" :key="line.id"><td class="px-4 py-3"><p class="font-medium">{{ line.sku }} — {{ line.item_name }}</p><p class="text-xs text-slate-500">{{ line.unit }}</p></td><td class="px-4 py-3">{{ line.system_qty }}</td><td class="px-4 py-3">{{ line.actual_qty }}</td><td class="px-4 py-3 font-semibold">{{ line.difference_qty > 0 ? `+${line.difference_qty}` : line.difference_qty }}</td></tr></tbody>
            </table>
          </div>
          <p v-if="selectedAdjustment.status === 'draft'" class="rounded-lg bg-amber-50 p-3 text-sm text-amber-800">Posting updates branch stock and records each non-zero movement in the stock card. Posted documents cannot be edited.</p>
          <p v-if="selectedAdjustment.adjustment_type === 'opening_balance'" class="text-xs text-slate-500">Only one opening balance can be posted for each item in a branch.</p>
        </div>
        <footer class="flex justify-end border-t border-slate-200 px-6 py-4"><button type="button" class="min-h-11 rounded-lg border border-slate-300 px-4 py-2 text-sm" @click="detailDialog = false">Close</button></footer>
      </template>
    </dialog>
  </section>
</template>
