<script setup lang="ts">
import AppModal from '@/components/ui/AppModal.vue'
import AppToast from '@/components/ui/AppToast.vue'
</script>

<template>
  <div class="paper-grain min-h-screen text-ink">
    <router-view />
    <AppToast />
    <AppModal />
  </div>
</template>
