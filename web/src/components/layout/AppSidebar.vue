<script setup>
import { ref, watch } from "vue";
import { useRoute } from "vue-router";
import {
  BookOpen,
  Boxes,
  Building2,
  ChevronDown,
  ClipboardCheck,
  ClipboardList,
  Code2,
  LayoutDashboard,
  Package,
  PackageCheck,
  ShoppingCart,
  SlidersHorizontal,
  Tags,
  Truck,
  Users,
  X,
} from "lucide-vue-next";
import { useAuthStore } from "@/stores/auth";

const auth = useAuthStore();

defineProps({
  collapsed: {
    type: Boolean,
    default: false,
  },
  mobileOpen: {
    type: Boolean,
    default: false,
  },
});

const emit = defineEmits(["close-mobile"]);

const route = useRoute();

const dashboard = {
  label: "Dashboard",
  path: "/dashboard",
  icon: LayoutDashboard,
};

const menuGroups = [
  {
    label: "Master",
    icon: Tags,
    children: [
      { label: "Items", path: "/items", icon: Boxes },
      { label: "Suppliers", path: "/suppliers", icon: Truck },
      { label: "Customers", path: "/sales/customers", icon: Users },
    ],
  },
  {
    label: "Purchasing",
    icon: ClipboardList,
    children: [
      { label: "Purchase Requests", path: "/inventory/purchase-requests", icon: ClipboardList },
      { label: "Purchase Orders", path: "/inventory/purchase-orders", icon: ClipboardCheck },
      { label: "Good Receipts", path: "/inventory/good-receipts", icon: PackageCheck },
      { label: "Goods Deliveries", path: "/inventory/goods-deliveries", icon: Truck },
    ],
  },
  {
    label: "Inventory",
    icon: Package,
    children: [
      { label: "Stock", path: "/inventory/stock", icon: Boxes },
      { label: "Stock Adjustments", path: "/inventory/stock-adjustments", icon: SlidersHorizontal },
    ],
  },
  {
    label: "Sales",
    icon: ShoppingCart,
    children: [
      { label: "Sales Orders", path: "/sales/orders", icon: ClipboardCheck },
    ],
  },
  {
    label: "Dev Tools",
    icon: Code2,
    children: [
      { label: "API Docs", path: "/api-docs", icon: BookOpen },
    ],
  },
];

const expandedGroups = ref({});

function isActive(path) {
  return route.path === path;
}

function syncGroupState(label, event) {
  expandedGroups.value[label] = event.currentTarget.open;
}

watch(
  () => route.path,
  (path) => {
    const activeGroup = menuGroups.find((group) => group.children.some((child) => child.path === path));
    if (activeGroup) expandedGroups.value[activeGroup.label] = true;
  },
  { immediate: true },
);
</script>

<template>
  <!-- Mobile overlay -->
  <div
    v-if="mobileOpen"
    class="fixed inset-0 z-40 bg-black/50 md:hidden"
    @click="emit('close-mobile')"
  ></div>

  <aside
    class="fixed inset-y-0 left-0 z-50 flex flex-col bg-slate-900 text-slate-300 transition-all duration-300 md:translate-x-0"
    :class="[
      collapsed ? 'md:w-20' : 'md:w-64',
      mobileOpen ? 'translate-x-0 w-64' : '-translate-x-full md:translate-x-0',
    ]"
  >
    <!-- Sidebar Header -->
    <div class="border-b border-slate-700 px-5 py-4">
      <!-- Logo -->
      <div class="flex h-9 items-center justify-between">
        <div v-if="!collapsed || mobileOpen" class="flex items-center gap-3">
          <div class="flex h-9 w-9 items-center justify-center rounded-lg bg-blue-600 text-white">
            <Boxes :size="21" />
          </div>

          <span class="text-lg font-bold tracking-wide text-white">ERP</span>
        </div>

        <div
          v-else
          class="mx-auto flex h-9 w-9 items-center justify-center rounded-lg bg-blue-600 text-white"
        >
          <Boxes :size="21" />
        </div>

        <button class="text-slate-400 hover:text-white md:hidden" @click="emit('close-mobile')">
          <X :size="21" />
        </button>
      </div>

      <!-- Company Information -->
      <div
        class="mt-4 rounded-lg bg-slate-800/70 p-3"
        :class="collapsed && !mobileOpen ? 'flex justify-center px-2' : ''"
        :title="collapsed && !mobileOpen ? auth.user?.company_name || 'Unknown Company' : ''"
      >
        <div class="flex items-center gap-3">
          <Building2 class="h-5 w-5 shrink-0 text-blue-400" />

          <div v-if="!collapsed || mobileOpen" class="min-w-0">
            <p class="text-[10px] font-semibold tracking-wider text-slate-400">CURRENT COMPANY</p>

            <p class="mt-1 truncate text-sm font-medium text-white">
              {{ auth.user?.company_name || "Unknown Company" }}
            </p>
          </div>
        </div>
      </div>
    </div>

    <!-- Navigation -->
    <nav aria-label="Main navigation" class="flex-1 space-y-1 overflow-y-auto px-3 py-5">
      <p
        v-if="!collapsed || mobileOpen"
        class="mb-3 px-3 text-xs font-semibold uppercase tracking-wider text-slate-500"
      >
        Main Menu
      </p>

      <RouterLink
        :to="dashboard.path"
        @click="emit('close-mobile')"
        class="flex items-center gap-3 rounded-lg px-3 py-3 text-sm font-medium transition-colors"
        :class="[
          isActive(dashboard.path)
            ? 'bg-blue-600 text-white'
            : 'text-slate-400 hover:bg-slate-800 hover:text-white',
          collapsed && !mobileOpen ? 'justify-center' : '',
        ]"
        :title="collapsed && !mobileOpen ? dashboard.label : undefined"
      >
        <component :is="dashboard.icon" :size="20" aria-hidden="true" />
        <span v-if="!collapsed || mobileOpen">{{ dashboard.label }}</span>
      </RouterLink>

      <details
        v-for="group in menuGroups"
        :key="group.label"
        :open="expandedGroups[group.label] || false"
        class="group/menu"
        @toggle="syncGroupState(group.label, $event)"
      >
        <summary
          :aria-label="group.label"
          :title="collapsed && !mobileOpen ? group.label : undefined"
          class="flex min-h-11 cursor-pointer list-none items-center gap-3 rounded-lg px-3 py-3 text-sm font-medium text-slate-400 transition-colors hover:bg-slate-800 hover:text-white focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-blue-400 [&::-webkit-details-marker]:hidden"
          :class="collapsed && !mobileOpen ? 'justify-center' : ''"
        >
          <component :is="group.icon" :size="20" aria-hidden="true" />
          <span v-if="!collapsed || mobileOpen" class="flex-1">{{ group.label }}</span>
          <ChevronDown
            v-if="!collapsed || mobileOpen"
            :size="16"
            aria-hidden="true"
            class="transition-transform"
            :class="expandedGroups[group.label] ? 'rotate-180' : ''"
          />
        </summary>

        <div class="mt-1 space-y-1">
          <RouterLink
            v-for="child in group.children"
            :key="child.path"
            :to="child.path"
            @click="emit('close-mobile')"
            class="flex min-h-10 items-center gap-3 rounded-lg px-3 py-2 text-sm transition-colors"
            :class="[
              isActive(child.path)
                ? 'bg-blue-600/20 text-blue-300'
                : 'text-slate-500 hover:bg-slate-800 hover:text-white',
              collapsed && !mobileOpen ? 'justify-center' : 'pl-11',
            ]"
            :title="collapsed && !mobileOpen ? child.label : undefined"
          >
            <component :is="child.icon" :size="18" aria-hidden="true" />
            <span v-if="!collapsed || mobileOpen">{{ child.label }}</span>
          </RouterLink>
        </div>
      </details>
    </nav>

    <!-- Sidebar footer -->
    <div class="border-t border-slate-700 p-4">
      <p v-if="!collapsed || mobileOpen" class="text-xs text-slate-500">ERP Application</p>

      <p v-else class="text-center text-xs text-slate-500">v1</p>
    </div>
  </aside>
</template>
