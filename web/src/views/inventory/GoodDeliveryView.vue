<script setup>
import { ClipboardList, Eye, Plus, Search, X } from 'lucide-vue-next'
import { useGoodDeliveries } from '@/composables/useGoodDeliveries'

const {
  deliveries, sources, sourceLines, selectedSource, selectedDelivery, loading, saving,
  page, pagination, searchInput, error, notice, modalError, dialog, detailDialog,
  gdDate, totalPages, totalQty, formatDateLabel,
  loadDeliveries, openCreate, selectSource, save, openDetails, applySearch,
} = useGoodDeliveries()
</script>

<template>
  <section class="space-y-6">
    <header class="flex flex-col justify-between gap-4 sm:flex-row sm:items-center">
      <div class="flex items-center gap-3"><ClipboardList :size="25" class="text-blue-600" aria-hidden="true" /><div><h1 class="text-2xl font-bold text-slate-800">Goods Deliveries</h1><p class="text-sm text-slate-500">Deliver open sales order quantities and update branch stock.</p></div></div>
      <button class="inline-flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2.5 text-sm font-semibold text-white hover:bg-blue-700" @click="openCreate"><Plus :size="18" />Post Goods Delivery</button>
    </header>
    <div v-if="notice" role="status" class="rounded-lg border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-800">{{ notice }}</div>
    <div v-if="error" role="alert" class="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">{{ error }}</div>

    <form class="grid gap-3 rounded-xl border border-slate-200 bg-white p-4 shadow-sm sm:grid-cols-[minmax(14rem,1fr)_auto]" @submit.prevent="applySearch">
      <label class="relative"><span class="sr-only">Search goods deliveries</span><Search :size="17" class="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" /><input v-model="searchInput" class="w-full rounded-lg border border-slate-300 py-2.5 pl-9 pr-3 text-sm" placeholder="Search GD number, SO number, branch" /></label>
      <button class="rounded-lg border border-slate-300 px-4 py-2.5 text-sm font-medium hover:bg-slate-50">Search</button>
    </form>

    <div class="overflow-hidden rounded-xl border border-slate-200 bg-white shadow-sm">
      <header class="flex items-center justify-between border-b border-slate-100 px-5 py-4"><div><h2 class="font-semibold text-slate-800">Goods delivery list</h2><p class="text-xs text-slate-500">{{ pagination.total || 0 }} delivery(ies)</p></div><button class="text-sm text-blue-700 underline disabled:opacity-50" :disabled="loading" @click="loadDeliveries">Refresh</button></header>
      <div class="overflow-x-auto"><table class="w-full min-w-[850px] text-left text-sm"><thead class="bg-slate-50 text-xs uppercase tracking-wide text-slate-500"><tr><th class="px-5 py-3">GD number</th><th class="px-5 py-3">Source</th><th class="px-5 py-3">Branch</th><th class="px-5 py-3">GD date</th><th class="px-5 py-3">Total quantity</th><th class="px-5 py-3">Status</th><th class="px-5 py-3">Actions</th></tr></thead><tbody class="divide-y divide-slate-100">
        <tr v-if="loading"><td colspan="7" class="px-5 py-12 text-center text-slate-500">Loading goods deliveries…</td></tr><tr v-else-if="!deliveries.length"><td colspan="7" class="px-5 py-12 text-center text-slate-500">No goods deliveries found.</td></tr><tr v-for="delivery in deliveries" v-else :key="delivery.id" class="hover:bg-slate-50"><td class="px-5 py-4 font-semibold text-slate-800">{{ delivery.gd_number }}</td><td class="px-5 py-4"><p>{{ delivery.ref_num }}</p><p class="text-xs text-slate-500">{{ delivery.ref_type.replaceAll('_', ' ') }}</p></td><td class="px-5 py-4">{{ delivery.branch_name }}</td><td class="px-5 py-4">{{ formatDateLabel(delivery.gd_date) }}</td><td class="px-5 py-4">{{ delivery.total_qty }}</td><td class="px-5 py-4"><span class="rounded-full bg-emerald-50 px-2.5 py-1 text-xs font-medium text-emerald-700">Posted</span></td><td class="px-5 py-4"><button class="inline-flex items-center gap-1 text-blue-700 hover:underline" @click="openDetails(delivery)"><Eye :size="16" />Details</button></td></tr>
      </tbody></table></div>
      <footer class="flex items-center justify-between border-t border-slate-100 px-5 py-3 text-sm"><span class="text-slate-500">Page {{ page }} of {{ totalPages }}</span><div class="flex gap-2"><button class="rounded-md border border-slate-300 px-3 py-1.5 disabled:opacity-40" :disabled="page <= 1 || loading" @click="page--">Previous</button><button class="rounded-md border border-slate-300 px-3 py-1.5 disabled:opacity-40" :disabled="page >= totalPages || loading" @click="page++">Next</button></div></footer>
    </div>

    <div v-if="dialog" class="fixed inset-0 z-50 flex items-center justify-center overflow-y-auto bg-slate-950/50 p-4" @click.self="!saving && (dialog = false)">
      <section role="dialog" aria-modal="true" aria-labelledby="gd-create-title" class="my-auto max-h-[92vh] w-full max-w-4xl overflow-y-auto rounded-xl bg-white shadow-xl">
        <header class="sticky top-0 flex items-start justify-between border-b border-slate-200 bg-white px-6 py-5"><div><h2 id="gd-create-title" class="text-lg font-semibold text-slate-800">Post Goods Delivery</h2><p class="mt-1 text-sm text-slate-500">Posting immediately reduces stock. The source sales order can be fulfilled in multiple deliveries.</p></div><button aria-label="Close" class="rounded-lg p-2 text-slate-500 hover:bg-slate-100" :disabled="saving" @click="dialog = false"><X :size="19" /></button></header>
        <form @submit.prevent="save"><div class="space-y-5 p-6">
          <div v-if="modalError" role="alert" class="rounded-lg bg-red-50 p-3 text-sm text-red-700">{{ modalError }}</div>
          <div class="grid gap-4 sm:grid-cols-2">
            <label class="space-y-1.5 text-sm font-medium text-slate-700">Source sales order <span class="text-red-600">*</span><select required class="w-full rounded-lg border border-slate-300 px-3 py-2.5 font-normal" :value="selectedSource?.ref_num || ''" @change="selectSource($event.target.value)"><option value="" disabled>Select a sales order</option><option v-for="source in sources" :key="source.ref_num" :value="source.ref_num">{{ source.ref_num }} — {{ source.customer_name }} / {{ source.branch_name }}</option></select></label>
            <label class="space-y-1.5 text-sm font-medium text-slate-700">Delivery date <span class="text-red-600">*</span><input v-model="gdDate" required type="date" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 font-normal" /></label>
          </div>
          <div v-if="selectedSource" class="rounded-lg bg-slate-50 p-3 text-sm text-slate-600"><p><strong>Branch:</strong> {{ selectedSource.branch_name }}</p><p><strong>Customer:</strong> {{ selectedSource.customer_name }}</p><p v-if="selectedSource.delivery_address"><strong>Delivery address:</strong> {{ selectedSource.delivery_address }}</p></div>
          <div v-if="selectedSource" class="space-y-3"><h3 class="font-semibold text-slate-800">Remaining sales order quantities</h3><div v-for="line in sourceLines" :key="line.item_id" class="grid gap-3 rounded-lg border border-slate-200 p-3 sm:grid-cols-[minmax(12rem,1fr)_7rem_8rem]"><div><p class="font-medium">{{ line.sku }} — {{ line.item_name }} ({{ line.unit }})</p><p class="text-xs text-slate-500">Ordered {{ line.ordered_qty }} · Delivered {{ line.delivered_qty }} · Remaining {{ line.remaining_qty }}</p></div><label class="text-xs font-medium text-slate-600">Delivery qty<input v-model="line.qty" type="number" min="1" :max="line.remaining_qty" step="1" class="mt-1 w-full rounded-md border border-slate-300 px-2.5 py-2 text-sm font-normal" /></label><p class="self-center text-xs text-slate-500">Maximum {{ line.remaining_qty }}</p></div><p v-if="!sourceLines.length" class="rounded-lg bg-amber-50 p-3 text-sm text-amber-800">No remaining quantities are available on this sales order.</p></div>
          <div v-if="sourceLines.length" class="flex justify-end border-t border-slate-100 pt-4 text-sm"><span class="text-slate-500">Total quantity to deliver:&nbsp;</span><strong>{{ totalQty }}</strong></div>
        </div><footer class="sticky bottom-0 flex justify-end gap-3 border-t border-slate-200 bg-white px-6 py-4"><button type="button" class="rounded-lg border border-slate-300 px-4 py-2 text-sm" :disabled="saving" @click="dialog = false">Close</button><button type="submit" class="rounded-lg bg-blue-600 px-4 py-2 text-sm font-semibold text-white disabled:opacity-60" :disabled="saving || !sourceLines.length">{{ saving ? 'Posting…' : 'Post Goods Delivery' }}</button></footer></form>
      </section>
    </div>

    <div v-if="detailDialog && selectedDelivery" class="fixed inset-0 z-50 flex items-center justify-center overflow-y-auto bg-slate-950/50 p-4" @click.self="detailDialog = false"><section role="dialog" aria-modal="true" aria-labelledby="gd-detail-title" class="my-auto max-h-[90vh] w-full max-w-3xl overflow-y-auto rounded-xl bg-white shadow-xl"><header class="flex items-start justify-between border-b border-slate-200 px-6 py-5"><div><h2 id="gd-detail-title" class="text-lg font-semibold text-slate-800">Goods Delivery {{ selectedDelivery.gd_number }}</h2><p class="text-sm text-slate-500">{{ selectedDelivery.ref_num }} · {{ selectedDelivery.branch_name }}</p></div><button aria-label="Close" class="rounded-lg p-2 text-slate-500 hover:bg-slate-100" @click="detailDialog = false"><X :size="19" /></button></header><div class="space-y-4 p-6"><div class="grid gap-3 text-sm sm:grid-cols-2"><p><strong>Source type:</strong> {{ selectedDelivery.ref_type.replaceAll('_', ' ') }}</p><p><strong>Delivery date:</strong> {{ formatDateLabel(selectedDelivery.gd_date) }}</p><p><strong>Status:</strong> Posted</p><p><strong>Total quantity:</strong> {{ selectedDelivery.total_qty }}</p></div><div v-for="line in selectedDelivery.lines" :key="line.id" class="flex justify-between border-t border-slate-100 py-3 text-sm"><span>{{ line.sku }} — {{ line.item_name }} · {{ line.unit }}</span><strong>{{ line.qty }}</strong></div></div></section></div>
  </section>
</template>
