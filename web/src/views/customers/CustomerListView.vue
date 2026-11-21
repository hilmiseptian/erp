<script setup>
import { Plus, Pencil, RefreshCw, Users, Search, Power } from 'lucide-vue-next'
import { useCustomers } from '@/composables/useCustomers'
import CustomerFormDialog from '@/components/customers/CustomerFormDialog.vue'

const { customers, loading, saving, error, notice, searchInput, status, page, pagination, pageCount, dialog, editing, formError, form, load, openCreate, openEdit, save, toggleActive, applySearch } = useCustomers()
</script>

<template>
  <section class="space-y-6">
    <header class="flex flex-col justify-between gap-4 sm:flex-row sm:items-center">
      <div class="flex items-center gap-3"><Users :size="24" class="text-blue-600" aria-hidden="true" /><div><h1 class="text-2xl font-bold text-slate-800">Customers</h1><p class="text-sm text-slate-500">Manage B2B customer records for sales orders.</p></div></div>
      <button class="inline-flex items-center gap-2 rounded-lg bg-blue-600 px-4 py-2.5 text-sm font-semibold text-white hover:bg-blue-700" @click="openCreate"><Plus :size="18" />Add Customer</button>
    </header>
    <div v-if="notice" role="status" class="rounded-lg border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-800">{{ notice }}</div>
    <div v-if="error" role="alert" class="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">{{ error }}</div>
    <form class="grid gap-3 rounded-xl border border-slate-200 bg-white p-4 shadow-sm sm:grid-cols-[minmax(14rem,1fr)_12rem_auto]" @submit.prevent="applySearch">
      <label class="relative"><span class="sr-only">Search customers</span><Search :size="17" class="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" /><input v-model="searchInput" class="w-full rounded-lg border border-slate-300 py-2.5 pl-9 pr-3 text-sm" placeholder="Search code, name, contact" /></label>
      <label><span class="sr-only">Filter customer status</span><select v-model="status" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-sm"><option value="all">All statuses</option><option value="active">Active</option><option value="inactive">Inactive</option></select></label>
      <button class="rounded-lg border border-slate-300 px-4 py-2.5 text-sm font-medium hover:bg-slate-50">Search</button>
    </form>
    <div class="overflow-hidden rounded-xl border border-slate-200 bg-white shadow-sm">
      <header class="flex items-center justify-between border-b border-slate-100 px-5 py-4"><div><h2 class="font-semibold text-slate-800">Customer list</h2><p class="text-xs text-slate-500">{{ pagination.total }} customer(s)</p></div><button aria-label="Refresh customers" class="rounded-lg p-2 text-slate-500 hover:bg-slate-100" :disabled="loading" @click="load"><RefreshCw :size="17" :class="loading ? 'animate-spin' : ''" /></button></header>
      <div class="overflow-x-auto"><table class="w-full min-w-[900px] text-left text-sm"><thead class="bg-slate-50 text-xs uppercase tracking-wide text-slate-500"><tr><th class="px-5 py-3">Code</th><th class="px-5 py-3">Customer</th><th class="px-5 py-3">Contact</th><th class="px-5 py-3">Phone</th><th class="px-5 py-3">Email</th><th class="px-5 py-3">Status</th><th class="px-5 py-3 text-right">Actions</th></tr></thead>
        <tbody class="divide-y divide-slate-100"><tr v-if="loading"><td colspan="7" class="px-5 py-12 text-center text-slate-500">Loading customers…</td></tr><tr v-else-if="!customers.length"><td colspan="7" class="px-5 py-12 text-center text-slate-500">No customers found.</td></tr><tr v-for="customer in customers" v-else :key="customer.id" class="hover:bg-slate-50"><td class="px-5 py-4 font-medium text-slate-700">{{ customer.customer_code }}</td><td class="px-5 py-4"><p class="font-medium text-slate-800">{{ customer.name }}</p><p v-if="customer.tax_number" class="text-xs text-slate-500">Tax no. {{ customer.tax_number }}</p></td><td class="px-5 py-4 text-slate-600">{{ customer.contact_person || '—' }}</td><td class="px-5 py-4 text-slate-600">{{ customer.phone || '—' }}</td><td class="px-5 py-4 text-slate-600">{{ customer.email || '—' }}</td><td class="px-5 py-4"><span class="rounded-full px-2.5 py-1 text-xs font-medium" :class="customer.is_active ? 'bg-emerald-50 text-emerald-700' : 'bg-slate-100 text-slate-600'">{{ customer.is_active ? 'Active' : 'Inactive' }}</span></td><td class="px-5 py-4"><div class="flex justify-end gap-2"><button class="rounded-lg p-2 text-slate-500 hover:bg-blue-50 hover:text-blue-700" :aria-label="`Edit ${customer.name}`" @click="openEdit(customer)"><Pencil :size="17" /></button><button class="rounded-lg p-2 text-slate-500 hover:bg-slate-100 hover:text-slate-800" :aria-label="`${customer.is_active ? 'Deactivate' : 'Activate'} ${customer.name}`" @click="toggleActive(customer)"><Power :size="17" /></button></div></td></tr></tbody></table></div>
      <footer class="flex items-center justify-between border-t border-slate-100 px-5 py-3 text-sm"><span class="text-slate-500">Page {{ page }} of {{ pageCount }}</span><div class="flex gap-2"><button class="rounded-md border border-slate-300 px-3 py-1.5 disabled:opacity-40" :disabled="page <= 1 || loading" @click="page--">Previous</button><button class="rounded-md border border-slate-300 px-3 py-1.5 disabled:opacity-40" :disabled="page >= pageCount || loading" @click="page++">Next</button></div></footer>
    </div>
    <CustomerFormDialog :open="dialog" :editing="editing" :form="form" :saving="saving" :error="formError" @close="dialog = false" @save="save" />
  </section>
</template>
