<script setup>
import { Plus, Pencil, Trash2, RefreshCw, Truck } from "lucide-vue-next";
import SupplierFormModal from "@/components/suppliers/SupplierFormModal.vue";
import { useSuppliers } from "@/composables/useSuppliers";

const {
  suppliers, loading, submitting, deletingId, errorMessage, modalError,
  successMessage, showModal, selectedSupplier, fetchSuppliers,
  openCreateModal, openEditModal, closeModal, saveSupplier, deleteSupplier,
} = useSuppliers();
</script>

<template>
  <div class="space-y-6">
    <!-- Page Header -->
    <div class="flex flex-col justify-between gap-4 sm:flex-row sm:items-center">
      <div>
        <div class="flex items-center gap-2">
          <Truck :size="23" class="text-blue-600" />

          <h2 class="text-2xl font-bold text-slate-800">Suppliers</h2>
        </div>

        <p class="mt-1 text-sm text-slate-500">Manage your supplier information.</p>
      </div>

      <button
        class="inline-flex items-center justify-center gap-2 rounded-lg bg-blue-600 px-4 py-2.5 text-sm font-medium text-white transition hover:bg-blue-700"
        @click="openCreateModal"
      >
        <Plus :size="18" />
        Add Supplier
      </button>
    </div>

    <!-- Notifications -->
    <div
      v-if="successMessage"
      class="flex items-center justify-between rounded-lg border border-green-200 bg-green-50 px-4 py-3 text-sm text-green-700"
    >
      <span>{{ successMessage }}</span>

      <button class="font-medium hover:text-green-900" @click="successMessage = ''">Dismiss</button>
    </div>

    <div
      v-if="errorMessage"
      class="flex flex-col gap-2 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700 sm:flex-row sm:items-center sm:justify-between"
    >
      <span>{{ errorMessage }}</span>

      <button
        class="inline-flex items-center gap-1 font-medium hover:text-red-900"
        @click="fetchSuppliers"
      >
        <RefreshCw :size="14" />
        Retry
      </button>
    </div>

    <!-- Table -->
    <div class="overflow-hidden rounded-xl border border-slate-200 bg-white shadow-sm">
      <div class="flex items-center justify-between border-b border-slate-200 px-5 py-4">
        <div>
          <h3 class="font-semibold text-slate-800">Supplier List</h3>

          <p class="mt-1 text-xs text-slate-500">{{ suppliers.length }} supplier(s)</p>
        </div>

        <button
          class="rounded-lg p-2 text-slate-500 transition hover:bg-slate-100 hover:text-slate-800 disabled:opacity-50"
          title="Refresh suppliers"
          :disabled="loading"
          @click="fetchSuppliers"
        >
          <RefreshCw :size="18" :class="loading ? 'animate-spin' : ''" />
        </button>
      </div>

      <div class="overflow-x-auto">
        <table class="w-full text-left text-sm">
          <thead class="bg-slate-50 text-xs uppercase tracking-wide text-slate-500">
            <tr>
              <th class="px-5 py-4 font-semibold">Code</th>
              <th class="px-5 py-4 font-semibold">Supplier Name</th>
              <th class="px-5 py-4 font-semibold">Contact Person</th>
              <th class="px-5 py-4 font-semibold">Phone</th>
              <th class="px-5 py-4 font-semibold">Email</th>
              <th class="px-5 py-4 font-semibold">Status</th>
              <th class="px-5 py-4 text-right font-semibold">Action</th>
            </tr>
          </thead>

          <tbody class="divide-y divide-slate-100">
            <!-- Loading -->
            <tr v-if="loading">
              <td colspan="7" class="px-5 py-12 text-center text-slate-500">
                <RefreshCw :size="20" class="mx-auto mb-2 animate-spin" />
                Loading suppliers...
              </td>
            </tr>

            <!-- Empty -->
            <tr v-else-if="suppliers.length === 0">
              <td colspan="7" class="px-5 py-12 text-center">
                <Truck :size="30" class="mx-auto mb-2 text-slate-300" />

                <p class="font-medium text-slate-700">No suppliers found</p>

                <p class="mt-1 text-xs text-slate-500">Add your first supplier to get started.</p>
              </td>
            </tr>

            <!-- Data -->
            <tr
              v-for="supplier in suppliers"
              v-else
              :key="supplier.id"
              class="transition hover:bg-slate-50"
            >
              <td class="px-5 py-4 font-medium text-slate-700">
                {{ supplier.code }}
              </td>

              <td class="px-5 py-4 font-medium text-slate-800">
                {{ supplier.name }}
              </td>

              <td class="px-5 py-4 text-slate-600">
                {{ supplier.contact_person || "-" }}
              </td>

              <td class="px-5 py-4 text-slate-600">
                {{ supplier.phone || "-" }}
              </td>

              <td class="px-5 py-4 text-slate-600">
                {{ supplier.email || "-" }}
              </td>

              <td class="px-5 py-4">
                <span
                  class="inline-flex rounded-full px-2.5 py-1 text-xs font-medium"
                  :class="
                    supplier.is_active
                      ? 'bg-green-100 text-green-700'
                      : 'bg-slate-100 text-slate-600'
                  "
                >
                  {{ supplier.is_active ? "Active" : "Inactive" }}
                </span>
              </td>

              <td class="px-5 py-4">
                <div class="flex justify-end gap-2">
                  <button
                    class="rounded-lg p-2 text-slate-500 transition hover:bg-blue-50 hover:text-blue-600"
                    title="Edit supplier"
                    @click="openEditModal(supplier.id)"
                  >
                    <Pencil :size="17" />
                  </button>

                  <button
                    class="rounded-lg p-2 text-slate-500 transition hover:bg-red-50 hover:text-red-600 disabled:opacity-50"
                    title="Delete supplier"
                    :disabled="deletingId === supplier.id"
                    @click="deleteSupplier(supplier)"
                  >
                    <Trash2 :size="17" :class="deletingId === supplier.id ? 'animate-pulse' : ''" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Reusable Form Modal -->
    <SupplierFormModal
      :show="showModal"
      :supplier="selectedSupplier"
      :loading="submitting"
      :error-message="modalError"
      @close="closeModal"
      @submit="saveSupplier"
    />
  </div>
</template>
