<script setup>
import { Boxes, Eye, Search, X } from "lucide-vue-next";
import { useStock } from "@/composables/useStock";

const { branches, branchId, rows, cards, loading, cardsLoading, error, search, page, pagination, selectedItem, cardDialog, cardPage, cardPagination, cardTotalPages, totalPages, formatDateLabel, loadStock, openCards } = useStock();
</script>

<template>
  <section class="space-y-6">
    <header class="flex flex-col justify-between gap-4 sm:flex-row sm:items-center">
      <div class="flex items-center gap-3"><Boxes :size="25" class="text-blue-600"/><div><h1 class="text-2xl font-bold text-slate-800">Stock</h1><p class="text-sm text-slate-500">Current item balances by branch and their movements.</p></div></div>
    </header>
    <div v-if="error" role="alert" class="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">{{ error }}</div>
    <div class="grid gap-3 rounded-xl border border-slate-200 bg-white p-4 shadow-sm md:grid-cols-[minmax(14rem,1fr)_14rem]">
      <label class="relative"><span class="sr-only">Search items</span><Search :size="17" class="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400"/><input v-model="search" class="w-full rounded-lg border border-slate-300 py-2.5 pl-9 pr-3 text-sm" placeholder="Search item name or SKU" @keydown.enter="page=1;loadStock()"/></label>
      <select v-model="branchId" aria-label="Branch" class="rounded-lg border border-slate-300 px-3 py-2.5 text-sm"><option value="" disabled>Select a branch</option><option v-for="branch in branches" :key="branch.id" :value="String(branch.id)">{{ branch.name }}</option></select>
    </div>
    <div class="overflow-hidden rounded-xl border border-slate-200 bg-white shadow-sm">
      <div class="flex items-center justify-between border-b border-slate-100 px-5 py-4"><div><h2 class="font-semibold text-slate-800">Current stock balances</h2><p class="text-xs text-slate-500">{{ pagination.total || 0 }} items</p></div><button class="text-sm text-blue-700 underline" :disabled="loading" @click="loadStock">Refresh</button></div>
      <div class="overflow-x-auto"><table class="w-full min-w-[750px] text-left text-sm">
        <thead class="bg-slate-50 text-xs uppercase tracking-wide text-slate-500"><tr><th class="px-5 py-3">Item</th><th class="px-5 py-3">Unit</th><th class="px-5 py-3">Quantity</th><th class="px-5 py-3">Minimum</th><th class="px-5 py-3">Last movement</th><th class="px-5 py-3">Action</th></tr></thead>
        <tbody class="divide-y divide-slate-100"><tr v-if="loading"><td colspan="6" class="px-5 py-12 text-center text-slate-500">Loading stock…</td></tr><tr v-else-if="!rows.length"><td colspan="6" class="px-5 py-12 text-center text-slate-500">No items found for this branch.</td></tr>
          <tr v-for="row in rows" v-else :key="row.item_id"><td class="px-5 py-4"><p class="font-medium">{{ row.item_name }}</p><p class="text-xs text-slate-500">{{ row.sku }}</p></td><td class="px-5 py-4">{{ row.unit }}</td><td class="px-5 py-4 font-semibold">{{ row.qty }}</td><td class="px-5 py-4">{{ row.min_stock }}</td><td class="px-5 py-4">{{ formatDateLabel(row.stock_date) }}</td><td class="px-5 py-4"><button class="inline-flex items-center gap-1 text-blue-700 hover:underline" @click="openCards(row)"><Eye :size="16"/>Details</button></td></tr>
        </tbody>
      </table></div>
      <footer class="flex items-center justify-between border-t border-slate-100 px-5 py-3 text-sm"><span class="text-slate-500">Page {{ page }} of {{ totalPages }}</span><div class="flex gap-2"><button class="rounded-md border border-slate-300 px-3 py-1.5 disabled:opacity-40" :disabled="page<=1||loading" @click="page--">Previous</button><button class="rounded-md border border-slate-300 px-3 py-1.5 disabled:opacity-40" :disabled="page>=totalPages||loading" @click="page++">Next</button></div></footer>
    </div>
    <div v-if="cardDialog && selectedItem" class="fixed inset-0 z-50 flex items-center justify-center overflow-y-auto bg-slate-950/50 p-4" @click.self="cardDialog=false">
      <section role="dialog" aria-modal="true" aria-labelledby="stock-card-title" class="my-auto max-h-[90vh] w-full max-w-4xl overflow-y-auto rounded-xl bg-white shadow-xl">
        <header class="flex items-start justify-between border-b border-slate-200 px-6 py-5"><div><h2 id="stock-card-title" class="text-lg font-semibold">Stock card</h2><p class="text-sm text-slate-500">{{ selectedItem.sku }} — {{ selectedItem.item_name }} · {{ branches.find((b)=>String(b.id)===branchId)?.name }}</p></div><button aria-label="Close" class="rounded-lg p-2 text-slate-500 hover:bg-slate-100" @click="cardDialog=false"><X :size="19"/></button></header>
        <div class="overflow-x-auto p-6"><table class="w-full min-w-[650px] text-left text-sm"><thead class="text-xs uppercase text-slate-500"><tr><th class="py-3">Date</th><th class="py-3">Reference</th><th class="py-3">Type</th><th class="py-3">In</th><th class="py-3">Out</th><th class="py-3">HPP</th></tr></thead><tbody class="divide-y divide-slate-100"><tr v-if="cardsLoading"><td colspan="6" class="py-10 text-center">Loading stock card…</td></tr><tr v-else-if="!cards.length"><td colspan="6" class="py-10 text-center text-slate-500">No stock movements recorded.</td></tr><tr v-for="card in cards" v-else :key="card.id"><td class="py-3">{{ new Intl.DateTimeFormat('en-US',{dateStyle:'medium',timeStyle:'short'}).format(new Date(card.stock_date)) }}</td><td class="py-3">{{ card.ref_num }}</td><td class="py-3">{{ card.trans_type }}</td><td class="py-3">{{ card.in_qty }}</td><td class="py-3">{{ card.out_qty }}</td><td class="py-3">{{ card.hpp }}</td></tr></tbody></table></div>
        <footer class="flex items-center justify-between border-t border-slate-100 px-6 py-3 text-sm"><span class="text-slate-500">{{ cardPagination.total || 0 }} movements · Page {{ cardPage }} of {{ cardTotalPages }}</span><div class="flex gap-2"><button class="rounded-md border border-slate-300 px-3 py-1.5 disabled:opacity-40" :disabled="cardPage <= 1 || cardsLoading" @click="cardPage--">Previous</button><button class="rounded-md border border-slate-300 px-3 py-1.5 disabled:opacity-40" :disabled="cardPage >= cardTotalPages || cardsLoading" @click="cardPage++">Next</button></div></footer>
      </section>
    </div>
  </section>
</template>
