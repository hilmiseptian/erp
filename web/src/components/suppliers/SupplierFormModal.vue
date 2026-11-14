<script setup>
import { reactive, watch } from "vue";
import { X } from "lucide-vue-next";

const props = defineProps({
  show: {
    type: Boolean,
    default: false,
  },
  supplier: {
    type: Object,
    default: null,
  },
  loading: {
    type: Boolean,
    default: false,
  },
  errorMessage: {
    type: String,
    default: "",
  },
});

const emit = defineEmits(["close", "submit"]);

const defaultForm = () => ({
  code: "",
  name: "",
  contact_person: "",
  phone: "",
  email: "",
  address: "",
  is_active: true,
});

const form = reactive(defaultForm());

const isEditing = () => Boolean(props.supplier?.id);

// Isi form setiap kali modal dibuka atau data supplier berubah.
watch(
  () => [props.show, props.supplier],
  ([show, supplier]) => {
    if (!show) return;

    Object.assign(form, defaultForm());

    if (supplier) {
      Object.assign(form, {
        code: supplier.code || "",
        name: supplier.name || "",
        contact_person: supplier.contact_person || "",
        phone: supplier.phone || "",
        email: supplier.email || "",
        address: supplier.address || "",
        is_active: supplier.is_active ?? true,
      });
    }
  },
  { immediate: true },
);

function handleSubmit() {
  emit("submit", {
    code: form.code.trim(),
    name: form.name.trim(),
    contact_person: form.contact_person.trim() || null,
    phone: form.phone.trim() || null,
    email: form.email.trim() || null,
    address: form.address.trim() || null,
    is_active: form.is_active,
  });
}
</script>

<template>
  <div
    v-if="show"
    class="fixed inset-0 z-50 flex items-center justify-center overflow-y-auto bg-slate-950/50 p-4"
    @click.self="emit('close')"
  >
    <div class="my-auto w-full max-w-2xl rounded-xl bg-white shadow-xl">
      <!-- Header -->
      <div class="flex items-start justify-between border-b border-slate-200 px-6 py-5">
        <div>
          <h2 class="text-lg font-semibold text-slate-800">
            {{ isEditing() ? "Edit Supplier" : "Add Supplier" }}
          </h2>

          <p class="mt-1 text-sm text-slate-500">
            {{ isEditing() ? "Update supplier information." : "Enter supplier information." }}
          </p>
        </div>

        <button
          type="button"
          :disabled="loading"
          class="rounded-lg p-2 text-slate-400 transition hover:bg-slate-100 hover:text-slate-700 disabled:opacity-50"
          @click="emit('close')"
        >
          <X :size="20" />
        </button>
      </div>

      <!-- Form -->
      <form @submit.prevent="handleSubmit">
        <div class="space-y-5 p-6">
          <div
            v-if="errorMessage"
            class="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700"
          >
            {{ errorMessage }}
          </div>

          <div class="grid grid-cols-1 gap-5 sm:grid-cols-2">
            <div>
              <label for="supplier-code" class="mb-1.5 block text-sm font-medium text-slate-700">
                Supplier Code <span class="text-red-500">*</span>
              </label>

              <input
                id="supplier-code"
                v-model="form.code"
                type="text"
                maxlength="30"
                required
                placeholder="SUP-001"
                class="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-sm outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-100"
              />
            </div>

            <div>
              <label for="supplier-name" class="mb-1.5 block text-sm font-medium text-slate-700">
                Supplier Name <span class="text-red-500">*</span>
              </label>

              <input
                id="supplier-name"
                v-model="form.name"
                type="text"
                maxlength="150"
                required
                placeholder="Supplier name"
                class="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-sm outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-100"
              />
            </div>

            <div>
              <label for="supplier-contact" class="mb-1.5 block text-sm font-medium text-slate-700">
                Contact Person
              </label>

              <input
                id="supplier-contact"
                v-model="form.contact_person"
                type="text"
                placeholder="Contact person"
                class="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-sm outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-100"
              />
            </div>

            <div>
              <label for="supplier-phone" class="mb-1.5 block text-sm font-medium text-slate-700">
                Phone
              </label>

              <input
                id="supplier-phone"
                v-model="form.phone"
                type="text"
                placeholder="Phone number"
                class="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-sm outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-100"
              />
            </div>

            <div class="sm:col-span-2">
              <label for="supplier-email" class="mb-1.5 block text-sm font-medium text-slate-700">
                Email
              </label>

              <input
                id="supplier-email"
                v-model="form.email"
                type="email"
                placeholder="supplier@example.com"
                class="w-full rounded-lg border border-slate-300 px-3 py-2.5 text-sm outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-100"
              />
            </div>

            <div class="sm:col-span-2">
              <label for="supplier-address" class="mb-1.5 block text-sm font-medium text-slate-700">
                Address
              </label>

              <textarea
                id="supplier-address"
                v-model="form.address"
                rows="3"
                placeholder="Supplier address"
                class="w-full resize-y rounded-lg border border-slate-300 px-3 py-2.5 text-sm outline-none transition focus:border-blue-500 focus:ring-2 focus:ring-blue-100"
              ></textarea>
            </div>

            <div class="sm:col-span-2">
              <label
                class="inline-flex cursor-pointer items-center gap-2 text-sm font-medium text-slate-700"
              >
                <input
                  v-model="form.is_active"
                  type="checkbox"
                  class="h-4 w-4 rounded border-slate-300 text-blue-600 focus:ring-blue-500"
                />
                Active supplier
              </label>
            </div>
          </div>
        </div>

        <!-- Footer -->
        <div class="flex justify-end gap-3 border-t border-slate-200 px-6 py-4">
          <button
            type="button"
            :disabled="loading"
            class="rounded-lg border border-slate-300 px-4 py-2.5 text-sm font-medium text-slate-700 transition hover:bg-slate-50 disabled:opacity-50"
            @click="emit('close')"
          >
            Cancel
          </button>

          <button
            type="submit"
            :disabled="loading"
            class="rounded-lg bg-blue-600 px-4 py-2.5 text-sm font-medium text-white transition hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-60"
          >
            {{ loading ? "Saving..." : isEditing() ? "Update Supplier" : "Save Supplier" }}
          </button>
        </div>
      </form>
    </div>
  </div>
</template>
