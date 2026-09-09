<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { ArrowRight, LoaderCircle, MessageSquareText } from '@lucide/vue'
import { useRouter } from 'vue-router'

import LoginDialog from '@/features/auth/components/LoginDialog.vue'
import { useAuthStore } from '@/features/auth/stores/auth'

const AUTO_OPEN_DELAY = 5_000

const router = useRouter()
const authStore = useAuthStore()

const loginOpen = ref(false)

let autoOpenTimer: number | null = null
let disposed = false

function clearAutoOpenTimer() {
    if (autoOpenTimer === null) {
        return
    }

    window.clearTimeout(autoOpenTimer)
    autoOpenTimer = null
}

async function handleAction() {
    clearAutoOpenTimer()

    if (authStore.isAuthenticated) {
        await router.push({ name: 'home' })
        return
    }

    loginOpen.value = true
}

function handleClose() {
    loginOpen.value = false
}

async function handleLoginSuccess() {
    loginOpen.value = false
    await router.push({ name: 'home' })
}

onMounted(async () => {
    await authStore.initialize()

    if (disposed || authStore.isAuthenticated) {
        return
    }

    // 仅在本次页面访问中自动提示一次，关闭后不会反复打扰用户。
    autoOpenTimer = window.setTimeout(() => {
        loginOpen.value = true
        autoOpenTimer = null
    }, AUTO_OPEN_DELAY)
})

onBeforeUnmount(() => {
    disposed = true
    clearAutoOpenTimer()
})
</script>

<template>
    <section class="border-t border-neutral-200 bg-neutral-50 px-4 py-6 dark:border-neutral-800 dark:bg-neutral-900">
        <div class="mx-auto flex max-w-4xl flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
            <div class="flex items-center gap-3">
                <div
                    class="flex size-10 shrink-0 items-center justify-center rounded-lg bg-neutral-900 text-white dark:bg-neutral-100 dark:text-neutral-900">
                    <MessageSquareText :size="20" />
                </div>

                <div>
                    <p class="font-medium text-neutral-900 dark:text-neutral-100">
                        继续你的 AI 对话
                    </p>
                    <p class="mt-1 text-sm text-neutral-500 dark:text-neutral-400">
                        登录 SY Chat，创建和管理自己的会话。
                    </p>
                </div>
            </div>

            <button type="button"
                class="inline-flex min-h-10 items-center justify-center gap-2 rounded-lg bg-neutral-900 px-4 text-sm font-medium text-white transition hover:bg-neutral-700 disabled:cursor-not-allowed disabled:opacity-60 dark:bg-neutral-100 dark:text-neutral-900 dark:hover:bg-white"
                :disabled="authStore.isLoading" @click="handleAction">
                <LoaderCircle v-if="authStore.isLoading" :size="17" class="animate-spin" />
                <ArrowRight v-else :size="17" />

                {{ authStore.isAuthenticated ? '进入聊天' : '登录并开始聊天' }}
            </button>
        </div>
    </section>

    <LoginDialog :open="loginOpen" @close="handleClose" @success="handleLoginSuccess" />
</template>