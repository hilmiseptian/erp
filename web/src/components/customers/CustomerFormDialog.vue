<script setup>
import { nextTick, ref, watch } from 'vue'
import { X } from 'lucide-vue-next'

const props = defineProps({ open: Boolean, editing: Object, form: Object, saving: Boolean, error: String })
defineEmits(['close', 'save'])
const dialogElement = ref(null)
watch(() => props.open, async (isOpen) => {
  await nextTick()
  if (isOpen && dialogElement.value && !dialogElement.value.open) dialogElement.value.showModal()
  if (!isOpen && dialogElement.value?.open) dialogElement.value.close()
}, { immediate: true })
</script>

<template>
  <dialog ref="dialogElement" v-show="open" class="fixed inset-0 z-50 m-0 flex h-full w-full items-center justify-center overflow-y-auto border-0 bg-slate-950/50 p-4 backdrop:bg-transparent" aria-modal="true" aria-labelledby="customer-dialog-title" @click.self="!saving && $emit('close')" @cancel.prevent="!saving && $emit('close')">
    <section class="my-auto max-h-[92vh] w-full max-w-2xl overflow-y-auto rounded-xl bg-white shadow-xl">
      <header class="flex items-start justify-between border-b border-slate-200 px-6 py-5">
        <div><h2 id="customer-dialog-title" class="text-lg font-semibold text-slate-800">{{ editing ? 'Edit Customer' : 'Add Customer' }}</h2><p class="mt-1 text-sm text-slate-500">Customer details for B2B sales orders.</p></div>
        <button type="button" aria-label="Close" class="rounded-lg p-2 text-slate-500 hover:bg-slate-100" :disabled="saving" @click="$emit('close')"><X :size="19" /></button>
      </header>
      <form @submit.prevent="$emit('save')">
        <div class="grid gap-4 p-6 sm:grid-cols-2">
          <div v-if="error" role="alert" class="rounded-lg bg-red-50 p-3 text-sm text-red-700 sm:col-span-2">{{ error }}</div>
          <label class="space-y-1.5 text-sm font-medium text-slate-700">Customer code <span class="text-red-600">*</span><input v-model="form.customer_code" required maxlength="40" autocomplete="off" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 font-normal" /></label>
          <label class="space-y-1.5 text-sm font-medium text-slate-700">Customer name <span class="text-red-600">*</span><input v-model="form.name" required maxlength="150" autocomplete="organization" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 font-normal" /></label>
          <label class="space-y-1.5 text-sm font-medium text-slate-700">Tax number<input v-model="form.tax_number" maxlength="50" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 font-normal" /></label>
          <label class="space-y-1.5 text-sm font-medium text-slate-700">Contact person<input v-model="form.contact_person" maxlength="100" autocomplete="name" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 font-normal" /></label>
          <label class="space-y-1.5 text-sm font-medium text-slate-700">Phone<input v-model="form.phone" maxlength="30" autocomplete="tel" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 font-normal" /></label>
          <label class="space-y-1.5 text-sm font-medium text-slate-700">Email<input v-model="form.email" type="email" maxlength="150" autocomplete="email" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 font-normal" /></label>
          <label class="space-y-1.5 text-sm font-medium text-slate-700 sm:col-span-2">Billing address<textarea v-model="form.billing_address" rows="2" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 font-normal"></textarea></label>
          <label class="space-y-1.5 text-sm font-medium text-slate-700 sm:col-span-2">Shipping address<textarea v-model="form.shipping_address" rows="2" class="w-full rounded-lg border border-slate-300 px-3 py-2.5 font-normal"></textarea></label>
        </div>
        <footer class="flex justify-end gap-3 border-t border-slate-200 px-6 py-4">
          <button type="button" class="rounded-lg border border-slate-300 px-4 py-2 text-sm" :disabled="saving" @click="$emit('close')">Close</button>
          <button type="submit" class="rounded-lg bg-blue-600 px-4 py-2 text-sm font-semibold text-white disabled:opacity-60" :disabled="saving">{{ saving ? 'Saving…' : 'Save Customer' }}</button>
        </footer>
      </form>
    </section>
  </dialog>
</template>
