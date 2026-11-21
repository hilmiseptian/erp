<script setup>
import { X } from 'lucide-vue-next'
import { formatCurrency, formatDateLabel } from '@/utils/formatters'

const formatSalesCurrency = (value) => formatCurrency(value, 'id-ID', 'IDR', 2)

defineProps({ open: Boolean, order: Object, statuses: Object })
defineEmits(['close'])
</script>

<template>
  <dialog v-if="open && order" open class="fixed inset-0 z-50 m-0 flex h-full w-full items-center justify-center overflow-y-auto border-0 bg-slate-950/50 p-4 backdrop:bg-transparent" aria-modal="true" aria-labelledby="sales-order-detail-title" @click.self="$emit('close')">
    <section class="my-auto max-h-[90vh] w-full max-w-3xl overflow-y-auto rounded-xl bg-white shadow-xl">
      <header class="flex items-start justify-between border-b border-slate-200 px-6 py-5"><div><h2 id="sales-order-detail-title" class="text-lg font-semibold text-slate-800">Sales Order {{ order.so_number }}</h2><p class="text-sm text-slate-500">{{ order.customer_name }}</p></div><button aria-label="Close" class="rounded-lg p-2 text-slate-500 hover:bg-slate-100" @click="$emit('close')"><X :size="19" /></button></header>
      <div class="space-y-5 p-6"><div class="grid gap-3 text-sm sm:grid-cols-2"><p><strong>Branch:</strong> {{ order.branch_name }}</p><p><strong>Status:</strong> {{ statuses[order.status] || order.status }}</p><p><strong>Order date:</strong> {{ formatDateLabel(order.so_date) }}</p><p><strong>Delivery date:</strong> {{ formatDateLabel(order.delivery_date) }}</p><p class="sm:col-span-2"><strong>Delivery address:</strong><br />{{ order.delivery_address || '—' }}</p></div>
        <div class="overflow-x-auto"><table class="w-full min-w-[600px] text-left text-sm"><thead class="border-y border-slate-100 text-xs uppercase text-slate-500"><tr><th class="py-3">Item</th><th class="py-3">Qty</th><th class="py-3">Unit price</th><th class="py-3">Discount</th><th class="py-3 text-right">Subtotal</th></tr></thead><tbody class="divide-y divide-slate-100"><tr v-for="line in order.lines" :key="line.id"><td class="py-3"><p class="font-medium">{{ line.sku }} — {{ line.item_name }}</p><p class="text-xs text-slate-500">{{ line.unit }}</p></td><td class="py-3">{{ line.qty }}</td><td class="py-3">{{ formatSalesCurrency(line.unit_price) }}</td><td class="py-3">{{ formatSalesCurrency(line.discount_amount) }}</td><td class="py-3 text-right font-medium">{{ formatSalesCurrency(line.subtotal) }}</td></tr></tbody></table></div>
        <div class="ml-auto grid max-w-sm grid-cols-2 gap-x-4 gap-y-2 border-t border-slate-200 pt-4 text-sm"><span class="text-slate-500">Total quantity</span><strong class="text-right">{{ order.total_qty }}</strong><span class="text-slate-500">Amount after line discounts</span><strong class="text-right">{{ formatSalesCurrency(order.total_amount) }}</strong><span class="text-slate-500">Order discount</span><strong class="text-right">{{ formatSalesCurrency(order.discount_amount) }}</strong><span class="font-semibold">Grand total</span><strong class="text-right text-base">{{ formatSalesCurrency(order.grand_total) }}</strong></div><p v-if="order.notes" class="rounded-lg bg-slate-50 p-3 text-sm"><strong>Notes:</strong> {{ order.notes }}</p>
      </div>
    </section>
  </dialog>
</template>
