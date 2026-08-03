<script setup lang="ts">
import { computed } from 'vue';
import { storeToRefs } from 'pinia';

import { useAppStore } from '@/stores/app';

const appStore = useAppStore()
const { apiStatus } = storeToRefs(appStore)

const statusText = computed(() => {
    const labels = {
        checking: '连接中',
        online: '服务正常',
        offline: '服务离线',
    }
    return labels[apiStatus.value]
})

withDefaults(
    defineProps<{
        title?: string
    }>(),
    {
        title: '新对话',
    },
)
</script>

<template>
    <header class="flex h-14 shrink-0 items-center border-b border-neutral-200 px-4 dark:border-neutral-800">
        <h1 class="truncate text-sm font-medium">
            {{ title }}
        </h1>

        <div class="ml-auto flex items-center gap-2 text-xs text-neutral-500">
            <span class="h-2 w-2 rounded-full" :class="{
                'animate-pulse bg-amber-500': apiStatus === 'checking',
                'bg-emerald-500': apiStatus === 'online',
                'bg-red-500': apiStatus === 'offline',
            }" />

            <span>{{ statusText }}</span>
        </div>
    </header>
</template>