<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { Bot, LoaderCircle, MessageSquareText } from '@lucide/vue'
import { useRoute } from 'vue-router'

import MarkdownContent from '@/features/chat/components/MarkdownContent.vue'
import type { PublicShareData } from '@/features/conversation/types'
import { ApiRequestError } from '@/lib/api/client'
import { getSharedConversation } from '@/lib/api/conversation'

const route = useRoute()
const conversation = ref<PublicShareData | null>(null)
const isLoading = ref(true)
const errorMessage = ref('')

const token = computed(() => {
    const value = route.params.token
    return typeof value === 'string' ? value : ''
})

const sharedDate = computed(() => {
    if (!conversation.value?.sharedAt) return ''

    return new Date(conversation.value.sharedAt).toLocaleString('zh-CN')
})

async function loadSharedConversation() {
    if (!token.value) {
        errorMessage.value = '分享链接无效'
        isLoading.value = false
        return
    }

    isLoading.value = true
    errorMessage.value = ''

    try {
        conversation.value = await getSharedConversation(token.value)
        document.title = `${conversation.value.title} - SY Chat 分享`
    } catch (error) {
        errorMessage.value = error instanceof ApiRequestError
            ? error.message
            : '读取分享内容失败'
    } finally {
        isLoading.value = false
    }
}

onMounted(loadSharedConversation)
</script>

<template>
    <div class="min-h-screen bg-white text-neutral-950 dark:bg-neutral-950 dark:text-neutral-100">
        <header class="flex h-14 items-center border-b border-neutral-200 px-4 dark:border-neutral-800">
            <RouterLink to="/chat" class="flex items-center gap-2 text-sm font-semibold">
                <MessageSquareText :size="19" />
                <span>SY Chat</span>
            </RouterLink>

            <span class="ml-auto text-xs text-neutral-500">只读分享</span>
        </header>

        <main class="mx-auto w-full max-w-3xl px-4 py-8">
            <div v-if="isLoading" class="flex min-h-64 items-center justify-center text-neutral-500">
                <LoaderCircle class="animate-spin" :size="22" />
                <span class="ml-2 text-sm">正在读取分享内容</span>
            </div>

            <div v-else-if="errorMessage" class="flex min-h-64 flex-col items-center justify-center text-center">
                <h1 class="text-lg font-semibold">分享不可用</h1>
                <p class="mt-2 text-sm text-neutral-500">{{ errorMessage }}</p>
            </div>

            <template v-else-if="conversation">
                <section class="border-b border-neutral-200 pb-6 dark:border-neutral-800">
                    <h1 class="break-words text-2xl font-semibold">
                        {{ conversation.title }}
                    </h1>

                    <div class="mt-2 flex flex-wrap gap-x-3 gap-y-1 text-xs text-neutral-500">
                        <span>由 {{ conversation.ownerName }} 分享</span>
                        <span v-if="sharedDate">{{ sharedDate }}</span>
                        <span>浏览 {{ conversation.viewCount }} 次</span>
                    </div>
                </section>

                <section class="space-y-8 py-8">
                    <article v-for="message in conversation.messages" :key="message.id" class="flex w-full gap-3"
                        :class="message.role === 'user'
                            ? 'justify-end'
                            : 'justify-start'">
                        <div v-if="message.role === 'assistant'"
                            class="flex h-8 w-8 shrink-0 items-center justify-center rounded-md bg-neutral-950 text-white dark:bg-white dark:text-neutral-950">
                            <Bot :size="17" />
                        </div>

                        <div class="min-w-0 max-w-[80%]" :class="message.role === 'user' ? 'text-right' : ''">
                            <div v-if="message.role === 'user'"
                                class="inline-block max-w-full whitespace-pre-wrap break-words rounded-lg bg-neutral-100 px-4 py-3 text-left text-sm leading-6 dark:bg-neutral-800">
                                {{ message.content }}
                            </div>

                            <MarkdownContent v-else :content="message.content" class="py-3 text-sm" />
                        </div>
                    </article>
                </section>
            </template>
        </main>
    </div>
</template>