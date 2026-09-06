<script setup lang="ts">
import { computed, ref } from 'vue';
import { storeToRefs } from 'pinia';
import { useRouter } from 'vue-router'
import {
    LoaderCircle,
    LogOut,
    UserRound,
    Download,
    Share2,
} from '@lucide/vue'

import { useAppStore } from '@/stores/app';
import { useAuthStore } from '@/features/auth/stores/auth'
import { useConversationStore } from '@/features/conversation/stores/conversation'

withDefaults(
    defineProps<{
        title?: string
        canExport?: boolean
        canShare?: boolean
    }>(),
    {
        title: '新对话',
        canExport: false,
        canShare: false,
    },
)

const emit = defineEmits<{
    share: []
    export: []
}>()

const router = useRouter()
const appStore = useAppStore()
const authStore = useAuthStore()
const conversationStore = useConversationStore()

const { apiStatus } = storeToRefs(appStore)
const { user } = storeToRefs(authStore)

const isLoggingOut = ref(false)

const statusText = computed(() => {
    const labels = {
        checking: '连接中',
        online: '服务正常',
        offline: '服务离线',
    }
    return labels[apiStatus.value]
})

const displayName = computed(() => {
    return user.value?.name || user.value?.username || '用户'
})

async function handleLogout() {
    if (isLoggingOut.value) {
        return
    }

    isLoggingOut.value = true

    try {
        await authStore.logout()
        conversationStore.reset()
        await router.replace({ name: 'login' })
    } catch (error) {
        console.error('Logout failed:', error)
    } finally {
        isLoggingOut.value = false
    }
}
</script>

<template>
    <header class="flex h-14 shrink-0 items-center gap-3 border-b border-neutral-200 px-4 dark:border-neutral-800">
        <h1 class="min-w-0 flex-1 truncate text-sm font-medium">
            {{ title }}
        </h1>

        <div class="flex shrink-0 items-center gap-2 text-xs text-neutral-500" :title="statusText">
            <span class="h-2 w-2 rounded-full" :class="{
                'animate-pulse bg-amber-500': apiStatus === 'checking',
                'bg-emerald-500': apiStatus === 'online',
                'bg-red-500': apiStatus === 'offline',
            }" />

            <span class="hidden sm:inline">
                {{ statusText }}
            </span>
        </div>

        <button type="button"
            class="flex h-8 w-8 shrink-0 items-center justify-center rounded-md text-neutral-500 transition hover:bg-neutral-100 hover:text-neutral-950 disabled:opacity-40 dark:hover:bg-neutral-800 dark:hover:text-white"
            :disabled="!canShare" title="分享会话" aria-label="分享当前会话" @click="emit('share')">
            <Share2 class="h-4 w-4" aria-hidden="true" />
        </button>

        <button type="button"
            class="flex h-8 w-8 shrink-0 items-center justify-center rounded-md text-neutral-500 transition hover:bg-neutral-100 hover:text-neutral-950 disabled:cursor-not-allowed disabled:opacity-40 dark:hover:bg-neutral-800 dark:hover:text-white"
            :disabled="!canExport" title="导出会话" aria-label="导出当前会话" @click="emit('export')">

            <Download class="h-4 w-4" aria-hidden="true" />
        </button>

        <div class="hidden max-w-40 items-center gap-2 border-l border-neutral-200 pl-3 text-sm dark:border-neutral-800 sm:flex"
            :title="user?.email">
            <UserRound class="h-4 w-4 shrink-0 text-neutral-500" aria-hidden="true" />

            <span class="truncate">
                {{ displayName }}
            </span>
        </div>

        <button type="button"
            class="flex h-8 w-8 shrink-0 items-center justify-center rounded-md text-neutral-500 transition hover:bg-neutral-100 hover:text-neutral-950 disabled:cursor-not-allowed disabled:opacity-50 dark:hover:bg-neutral-800 dark:hover:text-white"
            :disabled="isLoggingOut" title="退出登录" aria-label="退出登录" @click="handleLogout">
            <LoaderCircle v-if="isLoggingOut" class="h-4 w-4 animate-spin" aria-hidden="true" />
            <LogOut v-else class="h-4 w-4" aria-hidden="true" />
        </button>
    </header>
</template>