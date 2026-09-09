<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import {
    Bot,
    Brain,
    History,
    Send,
    Share2,
    Zap,
} from '@lucide/vue'

import LoginDialog from '@/features/auth/components/LoginDialog.vue'

const router = useRouter()
const loginDialogOpen = ref(false)
const landingMessage = ref('')
const pendingMessage = ref('')

function openLoginDialog() {
    pendingMessage.value = ''
    loginDialogOpen.value = true
}

function handleLandingSubmit() {
    const message = landingMessage.value.trim()
    if (!message) {
        return
    }

    pendingMessage.value = message
    loginDialogOpen.value = true
}

async function handleLoginSuccess() {
    loginDialogOpen.value = false
    await router.push({
        name: 'home',
        query: pendingMessage.value
            ? { message: pendingMessage.value }
            : undefined,
    })
}

const features = [
    {
        title: '实时响应',
        description: '流式输出回答，让对话自然连贯。',
        icon: Zap,
        color: 'text-amber-600 bg-amber-50 dark:bg-amber-950',
    },
    {
        title: '思考可见',
        description: '查看模型推理过程，更清楚地理解答案。',
        icon: Brain,
        color: 'text-violet-600 bg-violet-50 dark:bg-violet-950',
    },
    {
        title: '历史保存',
        description: '自动保存会话，随时继续之前的讨论。',
        icon: History,
        color: 'text-sky-600 bg-sky-50 dark:bg-sky-950',
    },
    {
        title: '一键分享',
        description: '生成公开链接，轻松分享完整对话。',
        icon: Share2,
        color: 'text-emerald-600 bg-emerald-50 dark:bg-emerald-950',
    },
]
</script>

<template>
    <div class="min-h-screen bg-white text-neutral-950 dark:bg-neutral-950 dark:text-white">
        <header class="border-b border-neutral-200 dark:border-neutral-800">
            <div class="mx-auto flex h-16 max-w-6xl items-center justify-between px-5">
                <RouterLink to="/" class="flex items-center gap-2 font-semibold">
                    <span
                        class="flex h-8 w-8 items-center justify-center rounded-md bg-neutral-950 text-white dark:bg-white dark:text-neutral-950">
                        <Bot :size="18" />
                    </span>
                    <span>SY Chat</span>
                </RouterLink>

                <nav class="flex items-center gap-2">
                    <button type="button"
                        class="px-3 py-2 text-sm font-medium text-neutral-600 hover:text-neutral-950 dark:text-neutral-300 dark:hover:text-white"
                        @click="openLoginDialog">
                        登录
                    </button>

                    <RouterLink :to="{ name: 'register', query: { redirect: '/chat' } }"
                        class="rounded-md bg-neutral-950 px-4 py-2 text-sm font-medium text-white hover:opacity-80 dark:bg-white dark:text-neutral-950">
                        注册
                    </RouterLink>
                </nav>
            </div>
        </header>

        <main>
            <section class="mx-auto max-w-5xl px-5 pb-16 pt-20 text-center">
                <h1 class="text-5xl font-bold md:text-6xl">SY Chat</h1>
                <p class="mx-auto mt-5 max-w-2xl text-lg leading-8 text-neutral-600 dark:text-neutral-300">
                    AI 对话助手，让智能对话变得简单而自然。
                </p>

                <form class="mx-auto mt-8 max-w-2xl" @submit.prevent="handleLandingSubmit">
                    <div class="relative">
                        <input v-model="landingMessage" type="text" autocomplete="off" autofocus
                            placeholder="输入消息开始对话..."
                            class="h-14 w-full rounded-lg border-2 border-neutral-300 bg-white px-4 pr-14 text-base outline-none transition focus:border-emerald-600 focus:ring-2 focus:ring-emerald-600/20 dark:border-neutral-700 dark:bg-neutral-900" />

                        <button type="submit"
                            class="absolute right-2 top-1/2 flex h-10 w-10 -translate-y-1/2 items-center justify-center rounded-md bg-neutral-950 text-white hover:opacity-80 disabled:cursor-not-allowed disabled:opacity-40 dark:bg-white dark:text-neutral-950"
                            :disabled="landingMessage.trim() === ''" aria-label="开始对话" title="开始对话">
                            <Send :size="18" />
                        </button>
                    </div>

                    <p class="mt-3 text-sm text-neutral-500">
                        发送后登录，即可继续这次对话
                    </p>
                </form>

                <div
                    class="mx-auto mt-14 max-w-3xl overflow-hidden rounded-lg border border-neutral-200 bg-neutral-50 text-left shadow-sm dark:border-neutral-800 dark:bg-neutral-900">
                    <div class="border-b border-neutral-200 px-4 py-3 text-sm font-medium dark:border-neutral-800">
                        新对话
                    </div>

                    <div class="space-y-5 p-5">
                        <div class="flex justify-end">
                            <p class="max-w-md rounded-lg bg-neutral-200 px-4 py-2.5 text-sm dark:bg-neutral-800">
                                帮我整理今天的学习计划。
                            </p>
                        </div>

                        <div class="flex items-start gap-3">
                            <span
                                class="flex h-8 w-8 shrink-0 items-center justify-center rounded-md bg-neutral-950 text-white dark:bg-white dark:text-neutral-950">
                                <Bot :size="17" />
                            </span>
                            <p class="pt-1 text-sm leading-7 text-neutral-700 dark:text-neutral-200">
                                当然。我们可以按照优先级、预计时间和完成状态来安排。
                            </p>
                        </div>
                    </div>
                </div>
            </section>

            <section
                class="border-t border-neutral-200 bg-neutral-50 py-14 dark:border-neutral-800 dark:bg-neutral-900">
                <div class="mx-auto grid max-w-6xl gap-4 px-5 sm:grid-cols-2 lg:grid-cols-4">
                    <article v-for="feature in features" :key="feature.title"
                        class="rounded-lg border border-neutral-200 bg-white p-5 dark:border-neutral-800 dark:bg-neutral-950">
                        <span class="flex h-10 w-10 items-center justify-center rounded-md" :class="feature.color">
                            <component :is="feature.icon" :size="20" />
                        </span>
                        <h2 class="mt-4 font-semibold">{{ feature.title }}</h2>
                        <p class="mt-2 text-sm leading-6 text-neutral-600 dark:text-neutral-400">
                            {{ feature.description }}
                        </p>
                    </article>
                </div>
            </section>
        </main>

        <LoginDialog :open="loginDialogOpen" @close="loginDialogOpen = false" @success="handleLoginSuccess" />
    </div>
</template>