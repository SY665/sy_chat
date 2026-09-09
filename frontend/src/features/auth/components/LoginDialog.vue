<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import {
    LoaderCircle,
    MessageSquareText,
    X,
} from '@lucide/vue'

import { useAuthStore } from '@/features/auth/stores/auth'

const props = defineProps<{
    open: boolean
}>()

const emit = defineEmits<{
    close: []
    success: []
}>()

const authStore = useAuthStore()

const identifierInput = ref<HTMLInputElement | null>(null)
const identifier = ref('')
const password = ref('')

const isDisabled = computed(() => {
    return (
        authStore.isLoading ||
        identifier.value.trim() === '' ||
        password.value === ''
    )
})

watch(
    () => props.open,
    async (open) => {
        if (!open) {
            return
        }

        authStore.clearError()
        await nextTick()
        identifierInput.value?.focus()
    },
)

function closeDialog() {
    if (!authStore.isLoading) {
        authStore.clearError()
        emit('close')
    }
}

async function handleSubmit() {
    if (isDisabled.value) {
        return
    }

    try {
        await authStore.login({
            identifier: identifier.value,
            password: password.value,
        })

        identifier.value = ''
        password.value = ''
        emit('success')
    } catch {
        // authStore 已保存后端返回的错误信息。
    }
}
</script>

<template>
    <Teleport to="body">
        <div v-if="open" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 px-4 py-8"
            role="presentation" @click.self="closeDialog" @keydown.esc="closeDialog">
            <section
                class="w-full max-w-sm rounded-lg border border-neutral-200 bg-white p-6 text-neutral-950 shadow-xl dark:border-neutral-800 dark:bg-neutral-900 dark:text-white"
                role="dialog" aria-modal="true" aria-labelledby="login-dialog-title">
                <header class="mb-6 flex items-start justify-between gap-4">
                    <div>
                        <div class="mb-3 flex items-center gap-2">
                            <MessageSquareText :size="22" class="text-emerald-600 dark:text-emerald-400"
                                aria-hidden="true" />
                            <span class="font-semibold">SY Chat</span>
                        </div>

                        <h2 id="login-dialog-title" class="text-xl font-semibold">
                            登录
                        </h2>
                    </div>

                    <button type="button"
                        class="flex h-8 w-8 shrink-0 items-center justify-center rounded-md text-neutral-500 hover:bg-neutral-100 disabled:opacity-40 dark:hover:bg-neutral-800"
                        :disabled="authStore.isLoading" aria-label="关闭登录弹窗" title="关闭" @click="closeDialog">
                        <X :size="18" />
                    </button>
                </header>

                <form class="space-y-4" @submit.prevent="handleSubmit">
                    <div>
                        <label for="dialog-identifier" class="mb-1.5 block text-sm font-medium">
                            用户名或邮箱
                        </label>

                        <input id="dialog-identifier" ref="identifierInput" v-model="identifier" type="text"
                            autocomplete="username"
                            class="h-10 w-full rounded-md border border-neutral-300 bg-white px-3 outline-none transition focus:border-emerald-600 focus:ring-2 focus:ring-emerald-600/20 dark:border-neutral-700 dark:bg-neutral-950" />
                    </div>

                    <div>
                        <label for="dialog-password" class="mb-1.5 block text-sm font-medium">
                            密码
                        </label>

                        <input id="dialog-password" v-model="password" type="password" autocomplete="current-password"
                            class="h-10 w-full rounded-md border border-neutral-300 bg-white px-3 outline-none transition focus:border-emerald-600 focus:ring-2 focus:ring-emerald-600/20 dark:border-neutral-700 dark:bg-neutral-950" />
                    </div>

                    <p v-if="authStore.error" class="text-sm text-red-600 dark:text-red-400" role="alert">
                        {{ authStore.error }}
                    </p>

                    <button type="submit"
                        class="flex h-10 w-full items-center justify-center gap-2 rounded-md bg-neutral-950 px-4 text-sm font-medium text-white hover:opacity-80 disabled:cursor-not-allowed disabled:opacity-50 dark:bg-white dark:text-neutral-950"
                        :disabled="isDisabled">
                        <LoaderCircle v-if="authStore.isLoading" :size="16" class="animate-spin" aria-hidden="true" />
                        登录
                    </button>
                </form>

                <p class="mt-5 text-center text-sm text-neutral-500">
                    还没有账号？
                    <RouterLink :to="{ name: 'register', query: { redirect: '/chat' } }"
                        class="font-medium text-emerald-700 hover:underline dark:text-emerald-400">
                        注册
                    </RouterLink>
                </p>
            </section>
        </div>
    </Teleport>
</template>