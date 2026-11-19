<script setup>
import { ClipboardList, Plus, Search, X } from "lucide-vue-next";
import { formatCurrency, formatDateLabel } from "@/utils/formatters";
import { usePurchaseRequests } from "@/composables/usePurchaseRequests";

const {
  requests, branches, suppliers, items, loading, saving, page, pagination, search,
  statusFilter, branchFilter, error, message, dialog, selected, formError, form,
  totalPages, statuses, loadRequests, openCreate, openEdit, addLine, removeLine,
  setItemPrice, save, cancelRequest,
} = usePurchaseRequests();
</script>

<template>
  <section class="space-y-6">
    <header class="flex flex-col justify-between gap-4 sm:flex-row sm:items-center">
      <div class="flex items-center gap-3">
        <ClipboardList :size="25" class="text-blue-600" aria-hidden="true" />
        <div>
          <h1 class="text-2xl font-bold text-slate-800">Purchase Requests</h1>
          <p class="text-sm text-slate-500">Request to purchase goods for the branch.</p>
        </div>
      </div>
      <button
        class="inline-flex items-center justify-center gap-2 rounded-lg bg-blue-600 px-4 py-2.5 text-sm font-semibold text-white hover:bg-blue-700 disabled:opacity-50"
        :disabled="branches.length === 0 || suppliers.length === 0 || items.length === 0"
        @click="openCreate"
      >
        <Plus :size="18" />Create PR
      </button>
    </header>
    <div
      v-if="message"
      role="status"
      class="rounded-lg border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-800"
    >
      {{ message }}
    </div>
    <div
      v-if="error"
      role="alert"
      class="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700"
    >
      {{ error }}
    </div>
    <div
      class="grid gap-3 rounded-xl border border-slate-200 bg-white p-4 shadow-sm md:grid-cols-[minmax(14rem,1fr)_12rem_14rem]"
    >
      <label class="relative"
        ><span class="sr-only">Search PR</span
        ><Search :size="17" class="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" /><input
          v-model="search"
          class="w-full rounded-lg border border-slate-300 py-2.5 pl-9 pr-3 text-sm"
          placeholder="Search PR number, branch, or requester"
          @keydown.enter="
            page = 1;
            loadRequests();
          "
      /></label>
      <label class="sr-only" for="pr-status-filter">Status</label
      ><select
        id="pr-status-filter"
        v-model="statusFilter"
        class="rounded-lg border border-slate-300 px-3 py-2.5 text-sm"
      >
        <option value="">All statuses</option>
        <option v-for="(label, value) in statuses" :key="value" :value="value">{{ label }}</option>
      </select>
      <label class="sr-only" for="pr-branch-filter">Branch</label
      ><select
        id="pr-branch-filter"
        v-model="branchFilter"
        class="rounded-lg border border-slate-300 px-3 py-2.5 text-sm"
      >
        <option value="">All branches</option>
        <option v-for="branch in branches" :key="branch.id" :value="branch.id">
          {{ branch.name }}
        </option>
      </select>
    </div>
    <div class="overflow-hidden rounded-xl border border-slate-200 bg-white shadow-sm">
      <div class="flex items-center justify-between border-b border-slate-100 px-5 py-4">
        <div>
          <h2 class="font-semibold text-slate-800">List PR</h2>
                <p class="text-xs text-slate-500">{{ pagination.total || 0 }} requests</p>
        </div>
        <button class="text-sm text-blue-700 underline" @click="loadRequests">Reload</button>
      </div>
      <div class="overflow-x-auto">
        <table class="w-full min-w-[900px] text-left text-sm">
          <thead class="bg-slate-50 text-xs uppercase tracking-wide text-slate-500">
            <tr>
              <th class="px-5 py-3">PR Number</th>
              <th class="px-5 py-3">Branch</th>
              <th class="px-5 py-3">Supplier</th>
              <th class="px-5 py-3">Total quantity</th>
              <th class="px-5 py-3">Total price</th>
              <th class="px-5 py-3">Status</th>
              <th class="px-5 py-3">Action</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100">
            <tr v-if="loading">
              <td colspan="7" class="px-5 py-12 text-center text-slate-500">Loading purchase requests…</td>
            </tr>
            <tr v-else-if="!requests.length">
              <td colspan="7" class="px-5 py-12 text-center text-slate-500">
                No purchase requests yet.
              </td>
            </tr>
            <tr v-for="row in requests" v-else :key="row.id" class="hover:bg-slate-50">
              <td class="px-5 py-4">
                <p class="font-semibold text-slate-800">{{ row.pr_number }}</p>
                <p class="text-xs text-slate-500">PR date: {{ formatDateLabel(row.pr_date) }}</p>
                <p class="text-xs text-slate-500">Required date: {{ formatDateLabel(row.req_date) }}</p>
              </td>
              <td class="px-5 py-4">{{ row.branch_name }}</td>
              <td class="px-5 py-4">{{ row.supplier_name }}</td>
              <td class="px-5 py-4">{{ row.total_qty }}</td>
              <td class="px-5 py-4">
                {{
                  formatCurrency(row.total_price)
                }}
              </td>
              <td class="px-5 py-4">
                <span
                  class="rounded-full bg-blue-50 px-2.5 py-1 text-xs font-medium text-blue-700"
                  >{{ statuses[row.status] || row.status }}</span
                >
              </td>
              <td class="px-5 py-4">
                <div class="flex gap-3">
                  <button class="text-blue-700 hover:underline" @click="openEdit(row)">
                    Detail / Edit</button
                  ><button
                    v-if="['open', 'partially_ordered'].includes(row.status)"
                    class="text-red-700 hover:underline"
                    @click="cancelRequest(row)"
                  >
                    Cancel
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <footer class="flex items-center justify-between border-t border-slate-100 px-5 py-3 text-sm">
        <span class="text-slate-500">Page {{ page }} from {{ totalPages }}</span>
        <div class="flex gap-2">
          <button
            class="rounded-md border border-slate-300 px-3 py-1.5 disabled:opacity-40"
            :disabled="page <= 1 || loading"
            @click="page--"
          >
            Previous</button
          ><button
            class="rounded-md border border-slate-300 px-3 py-1.5 disabled:opacity-40"
            :disabled="page >= totalPages || loading"
            @click="page++"
          >
            Next
          </button>
        </div>
      </footer>
    </div>

    <div
      v-if="dialog"
      class="fixed inset-0 z-50 flex items-center justify-center overflow-y-auto bg-slate-950/50 p-4"
      @click.self="!saving && (dialog = false)"
    >
      <section
        role="dialog"
        aria-modal="true"
        aria-labelledby="pr-dialog-title"
        class="my-auto max-h-[92vh] w-full max-w-3xl overflow-y-auto rounded-xl bg-white shadow-xl"
      >
        <header
          class="sticky top-0 flex items-start justify-between border-b border-slate-200 bg-white px-6 py-5"
        >
          <div>
            <h2 id="pr-dialog-title" class="text-lg font-semibold text-slate-800">
              {{ selected ? `Purchase Request ${selected.pr_number}` : "Create Purchase Request" }}
            </h2>
            <p class="mt-1 text-sm text-slate-500">Approval is not available at this stage.</p>
          </div>
          <button
            aria-label="Close"
            class="rounded-lg p-2 text-slate-500 hover:bg-slate-100"
            :disabled="saving"
            @click="dialog = false"
          >
            <X :size="19" />
          </button>
        </header>
        <form @submit.prevent="save">
          <div class="space-y-5 p-6">
            <div
              v-if="formError"
              role="alert"
              class="rounded-lg bg-red-50 p-3 text-sm text-red-700"
            >
              {{ formError }}
            </div>
            <div class="grid gap-4 sm:grid-cols-3">
              <label class="space-y-1.5 text-sm font-medium text-slate-700"
                >Branch <span class="text-red-600">*</span
                ><select
                  v-model="form.branch_id"
                  required
                  class="w-full rounded-lg border border-slate-300 px-3 py-2.5 font-normal"
                >
                  <option value="" disabled>Select branch</option>
                  <option v-for="branch in branches" :key="branch.id" :value="String(branch.id)">
                    {{ branch.name }}
                  </option>
                </select></label
              ><label class="space-y-1.5 text-sm font-medium text-slate-700"
                >Supplier <span class="text-red-600">*</span
                ><select
                  v-model="form.supplier_id"
                  required
                  class="w-full rounded-lg border border-slate-300 px-3 py-2.5 font-normal"
                >
                  <option value="" disabled>Select supplier</option>
                  <option
                    v-for="supplier in suppliers"
                    :key="supplier.id"
                    :value="String(supplier.id)"
                  >
                    {{ supplier.code }} — {{ supplier.name }}
                  </option>
                </select></label
              ><label class="space-y-1.5 text-sm font-medium text-slate-700"
                >PR date <span class="text-red-600">*</span
                ><input
                  v-model="form.pr_date"
                  required
                  type="date"
                  class="w-full rounded-lg border border-slate-300 px-3 py-2.5 font-normal"
              /></label>
              <label class="space-y-1.5 text-sm font-medium text-slate-700"
                >Required date <span class="text-red-600">*</span
                ><input
                  v-model="form.req_date"
                  required
                  type="date"
                  class="w-full rounded-lg border border-slate-300 px-3 py-2.5 font-normal"
              /></label>
            </div>
            <div class="space-y-3">
              <div class="flex items-center justify-between">
                <h3 class="font-semibold text-slate-800">Item details</h3>
                <button type="button" class="text-sm font-medium text-blue-700" @click="addLine">
                  + Add line
                </button>
              </div>
              <div
                v-for="(line, index) in form.lines"
                :key="index"
                class="grid gap-3 rounded-lg border border-slate-200 p-3 sm:grid-cols-[minmax(12rem,1fr)_7rem_9rem_auto]"
              >
                <label class="space-y-1 text-xs font-medium text-slate-600"
                  >Item<select
                    v-model="line.item_id"
                    required
                    class="w-full rounded-md border border-slate-300 px-2.5 py-2 text-sm font-normal"
                    @change="setItemPrice(line)"
                  >
                    <option value="" disabled>Select an item</option>
                    <option v-for="item in items" :key="item.id" :value="String(item.id)">
                      {{ item.sku }} — {{ item.name }} ({{ item.unit }})
                    </option>
                  </select></label
                ><label class="space-y-1 text-xs font-medium text-slate-600"
                  >Qty<input
                    v-model="line.qty"
                    required
                    type="number"
                    min="1"
                    step="1"
                    class="w-full rounded-md border border-slate-300 px-2.5 py-2 text-sm font-normal" /></label
                ><label class="space-y-1 text-xs font-medium text-slate-600"
                  >Price<input
                    v-model="line.price"
                    required
                    type="number"
                    min="0"
                    step="0.01"
                    class="w-full rounded-md border border-slate-300 px-2.5 py-2 text-sm font-normal" /></label
                ><button
                  v-if="form.lines.length > 1"
                  type="button"
                  class="self-end rounded-md p-2 text-red-600 hover:bg-red-50"
                  :aria-label="`Remove line ${index + 1}`"
                  @click="removeLine(index)"
                >
                  <X :size="17" />
                </button>
              </div>
            </div>
          </div>
          <footer
            class="sticky bottom-0 flex justify-end gap-3 border-t border-slate-200 bg-white px-6 py-4"
          >
            <button
              type="button"
              class="rounded-lg border border-slate-300 px-4 py-2 text-sm"
              :disabled="saving"
              @click="dialog = false"
            >
              Close</button
            ><button
              type="submit"
              class="rounded-lg bg-blue-600 px-4 py-2 text-sm font-semibold text-white disabled:opacity-60"
              :disabled="saving"
            >
              {{ saving ? "Saving…" : "Save PR" }}
            </button>
          </footer>
        </form>
      </section>
    </div>
  </section>
</template>
