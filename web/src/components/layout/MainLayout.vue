<script setup>
import { ref } from "vue";
import AppSidebar from "@/components/layout/AppSidebar.vue";
import AppHeader from "@/components/layout/AppHeader.vue";

const sidebarCollapsed = ref(false);
const mobileSidebarOpen = ref(false);

function toggleSidebar() {
  sidebarCollapsed.value = !sidebarCollapsed.value;
}

function toggleMobileSidebar() {
  mobileSidebarOpen.value = !mobileSidebarOpen.value;
}
</script>

<template>
  <div class="min-h-screen bg-slate-50">
    <!-- Sidebar -->
    <AppSidebar
      :collapsed="sidebarCollapsed"
      :mobile-open="mobileSidebarOpen"
      @close-mobile="mobileSidebarOpen = false"
    />

    <!-- Main content -->
    <div
      class="min-h-screen transition-all duration-300"
      :class="[sidebarCollapsed ? 'md:ml-20' : 'md:ml-64']"
    >
      <AppHeader
        :collapsed="sidebarCollapsed"
        @toggle-sidebar="toggleSidebar"
        @toggle-mobile-sidebar="toggleMobileSidebar"
      />

      <main class="p-4 md:p-6">
        <RouterView />
      </main>
    </div>
  </div>
</template>
