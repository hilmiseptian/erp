<script setup>
import { Archive, Package, Pencil, Plus, Search, X } from 'lucide-vue-next';
import { formatCurrency } from '@/utils/formatters';
import { useItems } from '@/composables/useItems';

const {
  items, categories, loading, saving, error, notice, searchInput, search, status,
  categoryId, page, pagination, formOpen, editing, formError, form, pageCount,
  loadItems, openCreate, openEdit, saveItem, toggleStatus,
} = useItems();
</script>

<template>
  <section class="space-y-6">
    <header class="flex flex-col justify-between gap-4 sm:flex-row sm:items-center">
      <div class="flex items-center gap-3">
        <Package :size="25" class="text-blue-600" aria-hidden="true" />
        <div><h1 class="text-2xl font-bold text-slate-800">Master Items</h1><p class="text-sm text-slate-500">Manage your company’s ERP item catalog.</p></div>
      </div>
      <button class="inline-flex items-center justify-center gap-2 rounded-lg bg-blue-600 px-4 py-2.5 text-sm font-semibold text-white hover:bg-blue-700" @click="openCreate"><Plus :size="18" />Add item</button>
    </header>

    <div v-if="notice" role="status" class="rounded-lg border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-800">{{ notice }}</div>
    <div v-if="error" role="alert" class="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">{{ error }} <button class="ml-2 underline" @click="loadItems">Retry</button></div>

    <div class="grid gap-3 rounded-xl border border-slate-200 bg-white p-4 shadow-sm md:grid-cols-[minmax(14rem,1fr)_12rem_14rem]">
      <label class="relative block"><span class="sr-only">Search items</span><Search :size="17" class="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" /><input v-model="searchInput" class="w-full rounded-lg border border-slate-300 py-2.5 pl-9 pr-3 text-sm focus:border-blue-500 focus:outline-none focus:ring-2 focus:ring-blue-100" placeholder="Search name, SKU, or description" @input="search = searchInput" /></label>
      <label class="sr-only" for="item-status">Status</label><select id="item-status" v-model="status" class="rounded-lg border border-slate-300 px-3 py-2.5 text-sm"><option value="all">All statuses</option><option value="active">Active</option><option value="inactive">Inactive</option></select>
      <label class="sr-only" for="item-category">Category</label><select id="item-category" v-model="categoryId" class="rounded-lg border border-slate-300 px-3 py-2.5 text-sm"><option value="">All categories</option><option v-for="category in categories" :key="category.id" :value="category.id">{{ category.name }}</option></select>
    </div>

    <div class="overflow-hidden rounded-xl border border-slate-200 bg-white shadow-sm">
      <div class="flex items-center justify-between border-b border-slate-100 px-5 py-4"><div><h2 class="font-semibold text-slate-800">Item list</h2><p class="text-xs text-slate-500">{{ pagination.total }} items</p></div></div>
      <div class="overflow-x-auto">
        <table class="w-full min-w-[850px] text-left text-sm">
          <thead class="bg-slate-50 text-xs uppercase tracking-wide text-slate-500"><tr><th scope="col" class="px-5 py-3">Item</th><th scope="col" class="px-5 py-3">Category</th><th scope="col" class="px-5 py-3">Cost price</th><th scope="col" class="px-5 py-3">Selling price</th><th scope="col" class="px-5 py-3">Stock*</th><th scope="col" class="px-5 py-3">Status</th><th scope="col" class="px-5 py-3">Actions</th></tr></thead>
          <tbody class="divide-y divide-slate-100">
            <tr v-if="loading"><td colspan="7" class="px-5 py-12 text-center text-slate-500">Loading items…</td></tr>
            <tr v-else-if="!items.length"><td colspan="7" class="px-5 py-12 text-center text-slate-500">No matching items found.</td></tr>
            <tr v-for="item in items" v-else :key="item.id" class="hover:bg-slate-50">
              <td class="px-5 py-4"><p class="font-medium text-slate-800">{{ item.name }}</p><p class="mt-0.5 text-xs text-slate-500">{{ item.sku }}</p></td><td class="px-5 py-4 text-slate-600">{{ item.category_name || '—' }}</td><td class="px-5 py-4 text-slate-600">{{ formatCurrency(item.cost_price, 'id-ID', 'IDR', 2) }}</td><td class="px-5 py-4 text-slate-600">{{ formatCurrency(item.selling_price, 'id-ID', 'IDR', 2) }}</td><td class="px-5 py-4 text-slate-600">{{ item.stock }} {{ item.unit }}<span class="block text-xs text-slate-400">Min. {{ item.min_stock }}</span></td><td class="px-5 py-4"><span class="rounded-full px-2.5 py-1 text-xs font-medium" :class="item.is_active ? 'bg-emerald-50 text-emerald-700' : 'bg-slate-100 text-slate-600'">{{ item.is_active ? 'Active' : 'Inactive' }}</span></td><td class="px-5 py-4"><div class="flex gap-2"><button class="rounded-md p-2 text-blue-700 hover:bg-blue-50" :aria-label="`Edit ${item.name}`" @click="openEdit(item)"><Pencil :size="17" /></button><button class="rounded-md p-2 text-slate-600 hover:bg-slate-100" :aria-label="`${item.is_active ? 'Deactivate' : 'Activate'} ${item.name}`" @click="toggleStatus(item)"><Archive :size="17" /></button></div></td>
            </tr>
          </tbody>
        </table>
      </div>
      <footer class="flex items-center justify-between border-t border-slate-100 px-5 py-3 text-sm"><span class="text-slate-500">Page {{ page }} of {{ pageCount }}</span><div class="flex gap-2"><button class="rounded-md border border-slate-300 px-3 py-1.5 disabled:opacity-40" :disabled="page <= 1 || loading" @click="page--">Previous</button><button class="rounded-md border border-slate-300 px-3 py-1.5 disabled:opacity-40" :disabled="page >= pageCount || loading" @click="page++">Next</button></div></footer>
    </div>
    <p class="text-xs text-slate-500">* Stock is read-only here. Balance changes will be managed through inventory transactions.</p>

    <div v-if="formOpen" class="fixed inset-0 z-50 flex items-center justify-center overflow-y-auto bg-slate-950/50 p-4" @click.self="!saving && (formOpen = false)">
      <section role="dialog" aria-modal="true" aria-labelledby="item-form-title" class="my-auto w-full max-w-2xl rounded-xl bg-white shadow-xl">
        <header class="flex items-start justify-between border-b border-slate-200 px-6 py-5"><div><h2 id="item-form-title" class="text-lg font-semibold text-slate-800">{{ editing ? 'Edit item' : 'Add item' }}</h2><p class="mt-1 text-sm text-slate-500">Core item data for ERP and POS.</p></div><button class="rounded-lg p-2 text-slate-500 hover:bg-slate-100" :disabled="saving" aria-label="Close" @click="formOpen = false"><X :size="19" /></button></header>
        <form @submit.prevent="saveItem"><div class="grid gap-4 p-6 sm:grid-cols-2">
          <div v-if="formError" role="alert" class="sm:col-span-2 rounded-lg bg-red-50 p-3 text-sm text-red-700">{{ formError }}</div>
          <label class="space-y-1.5 text-sm font-medium text-slate-700">Item name <span class="text-red-600">*</span><input v-model.trim="form.name" required maxlength="255" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 font-normal" /></label>
          <label class="space-y-1.5 text-sm font-medium text-slate-700">SKU <span class="text-red-600">*</span><input v-model.trim="form.sku" required maxlength="100" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 font-normal" /></label>
          <label class="space-y-1.5 text-sm font-medium text-slate-700">Category <span class="text-red-600">*</span><select v-model="form.category_id" required class="w-full rounded-lg border border-slate-300 px-3 py-2.5 font-normal"><option value="" disabled>Select a category</option><option v-for="category in categories" :key="category.id" :value="String(category.id)">{{ category.name }}</option></select></label>
          <label class="space-y-1.5 text-sm font-medium text-slate-700">Satuan dasar <span class="text-red-600">*</span><select v-model="form.unit" required class="w-full rounded-lg border border-slate-300 px-3 py-2.5 font-normal"><option value="pcs">Pcs</option><option value="kg">Kg</option><option value="liter">Liter</option><option value="box">Box</option></select></label>
          <label class="space-y-1.5 text-sm font-medium text-slate-700">Cost price<input v-model="form.cost_price" required type="number" min="0" step="0.01" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 font-normal" /></label>
          <label class="space-y-1.5 text-sm font-medium text-slate-700">Selling price<input v-model="form.selling_price" required type="number" min="0" step="0.01" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 font-normal" /></label>
          <label class="space-y-1.5 text-sm font-medium text-slate-700">Minimum stock<input v-model="form.min_stock" required type="number" min="0" step="1" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 font-normal" /></label>
          <label class="flex items-center gap-2 self-end pb-3 text-sm text-slate-700"><input v-model="form.is_active" type="checkbox" class="h-4 w-4 rounded border-slate-300" />Active and available for selection</label>
          <label class="space-y-1.5 text-sm font-medium text-slate-700 sm:col-span-2">Description<textarea v-model="form.description" rows="3" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 font-normal"></textarea></label>
          <p class="sm:col-span-2 text-xs text-slate-500">Stock balances are not edited here. Use inventory transactions when that module is available.</p>
        </div><footer class="flex justify-end gap-3 border-t border-slate-200 px-6 py-4"><button type="button" class="rounded-lg border border-slate-300 px-4 py-2 text-sm" :disabled="saving" @click="formOpen = false">Cancel</button><button type="submit" class="rounded-lg bg-blue-600 px-4 py-2 text-sm font-semibold text-white disabled:opacity-60" :disabled="saving">{{ saving ? 'Saving…' : 'Save item' }}</button></footer></form>
      </section>
    </div>
  </section>
</template>
