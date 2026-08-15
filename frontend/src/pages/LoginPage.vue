<script setup lang="ts">
import { computed, onBeforeUnmount, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { LoaderCircle, MessageSquareText } from '@lucide/vue'

import { useAuthStore } from '@/features/auth/stores/auth'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const identifier = ref('')
const password = ref('')

const isDisabled = computed(() => {
    return (
        authStore.isLoading ||
        identifier.value.trim() === '' ||
        password.value === ''
    )
})

async function handleSubmit() {
    if (isDisabled.value) {
        return
    }

    try {
        await authStore.login({
            identifier: identifier.value,
            password: password.value,
        })

        const redirect = route.query.redirect

        await router.replace(
            typeof redirect === 'string' && redirect.startsWith('/')
                ? redirect
                : "/",
        )
    } catch {
        // authStore 已经保存后端返回的错误信息。
    }
}

onBeforeUnmount(() => {
    authStore.clearError()
})

</script>

<template>
    <main
        class="flex min-h-screen items-center justify-center bg-neutral-50 px-4 py-10 text-neutral-900 dark:bg-neutral-950 dark:text-neutral-100">
        <section class="w-full max-w-sm">
            <div class="mb-8 flex items-center justify-center gap-2">
                <MessageSquareText class="h-7 w-7 text-emerald-600 dark:text-emerald-400" aria-hidden="true" />
                <span class="text-xl font-semibold">SY Chat</span>
            </div>

            <div
                class="rounded-lg border border-neutral-200 bg-white p-6 shadow-sm dark:border-neutral-800 dark:bg-neutral-900">
                <header class="mb-6">
                    <h1 class="text-xl font-semibold">登录</h1>
                </header>

                <form class="space-y-4" @submit.prevent="handleSubmit">
                    <div>
                        <label for="identifier" class="mb-1.5 block text-sm font-medium">
                            用户名或邮箱
                        </label>

                        <input id="identifier" v-model="identifier" type="text" name="identifier"
                            autocomplete="username" autofocus
                            class="h-10 w-full rounded-md border border-neutral-300 bg-white px-3 outline-none transition focus:border-emerald-600 focus:ring-2 focus:ring-emerald-600/20 dark:border-neutral-700 dark:bg-neutral-950" />
                    </div>

                    <div>
                        <label for="password" class="mb-1.5 block text-sm font-medium">
                            密码
                        </label>

                        <input id="password" v-model="password" type="password" name="password"
                            autocomplete="current-password"
                            class="h-10 w-full rounded-md border border-neutral-300 bg-white px-3 outline-none transition focus:border-emerald-600 focus:ring-2 focus:ring-emerald-600/20 dark:border-neutral-700 dark:bg-neutral-950" />
                    </div>

                    <p v-if="authStore.error" role="alert" class="text-sm text-red-600 dark:text-red-400">
                        {{ authStore.error }}
                    </p>

                    <button type="submit" :disabled="isDisabled"
                        class="flex h-10 w-full items-center justify-center gap-2 rounded-md bg-neutral-900 px-4 text-sm font-medium text-white transition hover:bg-neutral-700 disabled:cursor-not-allowed disabled:opacity-50 dark:bg-neutral-100 dark:text-neutral-900 dark:hover:bg-white">
                        <LoaderCircle v-if="authStore.isLoading" class="h-4 w-4 animate-spin" aria-hidden="true" />
                        登录
                    </button>
                </form>

                <p class="mt-5 text-center text-sm text-neutral-500">
                    还没有账号？
                    <RouterLink :to="{ name: 'register', query: { redirect: route.query.redirect } }"
                        class="font-medium text-emerald-700 hover:underline dark:text-emerald-400">
                        注册
                    </RouterLink>
                </p>
            </div>
        </section>
    </main>
</template>