<script setup lang="ts">
import { RouterView } from 'vue-router'
import { onMounted } from 'vue';

import { useAppStore } from './stores/app';

const appStore = useAppStore()

onMounted(() => {
  appStore.initializeTheme
})
</script>

<template>
  <RouterView />
</template>