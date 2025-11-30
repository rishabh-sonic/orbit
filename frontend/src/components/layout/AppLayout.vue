<script setup lang="ts">
import { ref } from 'vue'
import AppHeader from './AppHeader.vue'
import AppSidebar from './AppSidebar.vue'
import { RouterView } from 'vue-router'

const sidebarOpen = ref(false)
</script>

<template>
  <div class="min-h-screen bg-background flex flex-col">
    <AppHeader @toggle-sidebar="sidebarOpen = !sidebarOpen" />
    <div class="flex flex-1 pt-14">
      <!-- Sidebar overlay on mobile -->
      <div
        v-if="sidebarOpen"
        class="fixed inset-0 z-20 bg-black/40 lg:hidden"
        @click="sidebarOpen = false"
      />

      <AppSidebar :open="sidebarOpen" @close="sidebarOpen = false" />

      <main class="flex-1 lg:ml-56 min-w-0">
        <div class="max-w-3xl mx-auto px-4 py-6">
          <RouterView />
        </div>
      </main>
    </div>
  </div>
</template>
