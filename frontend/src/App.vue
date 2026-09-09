<script setup lang="ts">
import { RouterView } from 'vue-router'
import { onMounted } from 'vue';

import AppErrorBoundary from '@/components/ui/AppErrorBoundary.vue'
import ToastViewport from '@/components/ui/ToastViewport.vue'
import { useAppStore } from './stores/app';

const appStore = useAppStore()

onMounted(() => {
  appStore.initializeTheme()

  // 健康检查不需要阻塞页面显示。
  void appStore.checkApiHealth()
})
</script>

<template>
  <AppErrorBoundary>
    <RouterView />
  </AppErrorBoundary>

  <ToastViewport />
</template>