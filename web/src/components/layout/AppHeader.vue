<script setup>
import { computed } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useAuthStore } from "@/stores/auth";

import { Menu, PanelLeftClose, PanelLeftOpen, LogOut, UserRound } from "lucide-vue-next";

defineProps({
  collapsed: {
    type: Boolean,
    default: false,
  },
});

const emit = defineEmits(["toggle-sidebar", "toggle-mobile-sidebar"]);

const auth = useAuthStore();
const route = useRoute();
const router = useRouter();

const pageTitle = computed(() => {
  return route.meta.title || "Dashboard";
});

const userName = computed(() => {
  return auth.user?.name || "User";
});

async function handleLogout() {
  const confirmed = window.confirm("Are you sure you want to logout?");

  if (!confirmed) return;

  try {
    await auth.logout();
  } catch (error) {
    // Local logout tetap dilakukan oleh auth store.
    console.error("Logout request failed:", error);
  } finally {
    await router.replace("/login");
  }
}
</script>

<template>
  <header
    class="sticky top-0 z-30 flex h-16 items-center justify-between border-b border-slate-200 bg-white px-4 md:px-6"
  >
    <div class="flex items-center gap-3">
      <!-- Desktop sidebar toggle -->
      <button
        class="hidden rounded-lg p-2 text-slate-500 hover:bg-slate-100 hover:text-slate-800 md:block"
        title="Toggle sidebar"
        @click="emit('toggle-sidebar')"
      >
        <PanelLeftClose v-if="!collapsed" :size="21" />
        <PanelLeftOpen v-else :size="21" />
      </button>

      <!-- Mobile sidebar toggle -->
      <button
        class="rounded-lg p-2 text-slate-500 hover:bg-slate-100 md:hidden"
        @click="emit('toggle-mobile-sidebar')"
      >
        <Menu :size="22" />
      </button>

      <h1 class="text-lg font-semibold text-slate-800">
        {{ pageTitle }}
      </h1>
    </div>

    <div class="flex items-center gap-3">
      <div class="hidden items-center gap-2 sm:flex">
        <div
          class="flex h-9 w-9 items-center justify-center rounded-full bg-blue-100 text-blue-700"
        >
          <UserRound :size="18" />
        </div>

        <div class="text-sm">
          <p class="font-medium text-slate-800">
            {{ userName }}
          </p>

          <p class="text-xs text-slate-500">Administrator</p>
        </div>
      </div>

      <button
        class="flex items-center gap-2 rounded-lg border border-slate-200 px-3 py-2 text-sm text-slate-600 transition hover:bg-red-50 hover:text-red-600"
        @click="handleLogout"
      >
        <LogOut :size="17" />

        <span class="hidden sm:inline"> Logout </span>
      </button>
    </div>
  </header>
</template>
