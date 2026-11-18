<script setup>
import { useRoute } from "vue-router";
import { LayoutDashboard, Truck, Package, X, Boxes, Tags, BookOpen } from "lucide-vue-next";
import { Building2 } from "lucide-vue-next";
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

const menuItems = [
  {
    label: "Dashboard",
    path: "/dashboard",
    icon: LayoutDashboard,
    active: true,
  },
  {
    label: "Items",
    path: "/items",
    icon: Tags,
    active: true,
  },
  {
    label: "Suppliers",
    path: "/suppliers",
    icon: Truck,
    active: true,
  },
  {
    label: "API Docs",
    path: "/api-docs",
    icon: BookOpen,
    active: true,
  },
  {
    label: "Inventory",
    icon: Package,
    active: true,
    children: [
      { label: "Purchase Requests", path: "/inventory/purchase-requests" },
      { label: "Purchase Orders", path: "/inventory/purchase-orders" },
    ],
  },
];

function isActive(path) {
  return path && route.path === path;
}
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
    <nav class="flex-1 space-y-1 overflow-y-auto px-3 py-5">
      <p
        v-if="!collapsed || mobileOpen"
        class="mb-3 px-3 text-xs font-semibold uppercase tracking-wider text-slate-500"
      >
        Main Menu
      </p>

      <template v-for="item in menuItems" :key="item.label">
        <div v-if="item.children && item.active" class="space-y-1">
          <div class="flex items-center gap-3 rounded-lg px-3 py-3 text-sm font-medium text-slate-400" :class="collapsed && !mobileOpen ? 'justify-center' : ''">
            <component :is="item.icon" :size="20" />
            <span v-if="!collapsed || mobileOpen">{{ item.label }}</span>
          </div>
          <RouterLink v-for="child in item.children" :key="child.path" :to="child.path" @click="emit('close-mobile')" class="block rounded-lg py-2 pl-11 pr-3 text-sm transition-colors" :class="route.path === child.path ? 'bg-blue-600/20 text-blue-300' : 'text-slate-500 hover:bg-slate-800 hover:text-white'" :title="collapsed && !mobileOpen ? child.label : undefined">
            <span v-if="!collapsed || mobileOpen">{{ child.label }}</span>
          </RouterLink>
        </div>
        <RouterLink
          v-else-if="item.active"
          :to="item.path"
          @click="emit('close-mobile')"
          class="flex items-center gap-3 rounded-lg px-3 py-3 text-sm font-medium transition-colors"
          :class="[
            isActive(item.path)
              ? 'bg-blue-600 text-white'
              : 'text-slate-400 hover:bg-slate-800 hover:text-white',
            collapsed && !mobileOpen ? 'justify-center' : '',
          ]"
          :title="collapsed && !mobileOpen ? item.label : undefined"
        >
          <component :is="item.icon" :size="20" />

          <span v-if="!collapsed || mobileOpen">
            {{ item.label }}
          </span>
        </RouterLink>

        <div
          v-else
          class="flex cursor-not-allowed items-center gap-3 rounded-lg px-3 py-3 text-sm text-slate-600"
          :class="collapsed && !mobileOpen ? 'justify-center' : ''"
          :title="collapsed && !mobileOpen ? `${item.label} (Coming soon)` : undefined"
        >
          <component :is="item.icon" :size="20" />

          <span v-if="!collapsed || mobileOpen">
            {{ item.label }}
            <span class="ml-1 text-xs">(Soon)</span>
          </span>
        </div>
      </template>
    </nav>

    <!-- Sidebar footer -->
    <div class="border-t border-slate-700 p-4">
      <p v-if="!collapsed || mobileOpen" class="text-xs text-slate-500">ERP Application</p>

      <p v-else class="text-center text-xs text-slate-500">v1</p>
    </div>
  </aside>
</template>
