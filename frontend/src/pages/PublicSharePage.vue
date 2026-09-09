<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { LoaderCircle, MessageSquareText } from '@lucide/vue'
import { useRoute } from 'vue-router'

import type { PublicShareData } from '@/features/conversation/types'
import ShareMessage from '@/features/share/components/ShareMessage.vue'
import ShareLoginPrompt from '@/features/share/components/ShareLoginPrompt.vue'
import SharePageFooter from '@/features/share/components/SharePageFooter.vue'
import SharePageHeader from '@/features/share/components/SharePageHeader.vue'
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

type MetaAttribute = 'name' | 'property'

const SHARE_META_MARKER = 'data-sy-chat-share-meta'

function appendShareMeta(
    attribute: MetaAttribute,
    key: string,
    content: string,
) {
    const element = document.createElement('meta')

    element.setAttribute(attribute, key)
    element.setAttribute(SHARE_META_MARKER, '')
    element.content = content

    document.head.append(element)
}

function clearShareMetadata() {
    document.head
        .querySelectorAll(`meta[${SHARE_META_MARKER}]`)
        .forEach((element) => element.remove())
}

function applyShareMetadata(data: PublicShareData) {
    const title = `${data.title} - SY Chat 分享`
    const description = `查看 ${data.ownerName || '用户'} 分享的对话：${data.title}`

    clearShareMetadata()
    document.title = title

    appendShareMeta('name', 'description', description)
    appendShareMeta('name', 'author', data.ownerName || 'SY Chat')
    appendShareMeta('name', 'twitter:card', 'summary')
    appendShareMeta('name', 'twitter:title', title)
    appendShareMeta('name', 'twitter:description', description)
    appendShareMeta('property', 'og:type', 'article')
    appendShareMeta('property', 'og:title', title)
    appendShareMeta('property', 'og:description', description)
    appendShareMeta('property', 'og:url', window.location.href)
    if (data.sharedAt) {
        appendShareMeta('property', 'article:published_time', data.sharedAt)
    }
}

function applyUnavailableMetadata(message: string) {
    clearShareMetadata()
    document.title = '分享不可用 - SY Chat'

    appendShareMeta('name', 'description', message)
    appendShareMeta('name', 'robots', 'noindex')
}

async function loadSharedConversation() {
    if (!token.value) {
        errorMessage.value = '分享链接无效'
        applyUnavailableMetadata(errorMessage.value)
        isLoading.value = false
        return
    }

    isLoading.value = true
    errorMessage.value = ''

    try {
        conversation.value = await getSharedConversation(token.value)
        applyShareMetadata(conversation.value)
    } catch (error) {
        errorMessage.value = error instanceof ApiRequestError
            ? error.message
            : '读取分享内容失败'

        applyUnavailableMetadata(errorMessage.value)
    } finally {
        isLoading.value = false
    }
}

onMounted(loadSharedConversation)

onBeforeUnmount(() => {
    // 防止分享页的标题和元信息残留到其他 SPA 页面。
    clearShareMetadata()
    document.title = 'SY Chat'
})
</script>

<template>
    <div class="flex min-h-screen flex-col bg-white text-neutral-950 dark:bg-neutral-950 dark:text-neutral-100">
        <header v-if="!conversation"
            class="flex h-14 items-center border-b border-neutral-200 px-4 dark:border-neutral-800">
            <RouterLink to="/" class="flex items-center gap-2 text-sm font-semibold">
                <MessageSquareText :size="19" class="text-emerald-600 dark:text-emerald-400" />
                <span>SY Chat</span>
            </RouterLink>

            <span class="ml-auto text-xs text-neutral-500">
                只读分享
            </span>
        </header>

        <SharePageHeader v-else :title="conversation.title" :owner-name="conversation.ownerName"
            :shared-at="conversation.sharedAt" :view-count="conversation.viewCount" />

        <main class="mx-auto w-full max-w-4xl flex-1 px-4 py-8 sm:px-6">
            <div v-if="isLoading" class="flex min-h-64 items-center justify-center text-neutral-500">
                <LoaderCircle class="animate-spin" :size="22" />
                <span class="ml-2 text-sm">正在读取分享内容</span>
            </div>

            <div v-else-if="errorMessage" class="flex min-h-64 flex-col items-center justify-center text-center">
                <h1 class="text-lg font-semibold">分享不可用</h1>
                <p class="mt-2 text-sm text-neutral-500">{{ errorMessage }}</p>
            </div>

            <template v-else-if="conversation">
                <div v-if="conversation.messages.length === 0"
                    class="flex min-h-72 flex-col items-center justify-center text-center">
                    <MessageSquareText :size="30" class="text-neutral-400" aria-hidden="true" />

                    <p class="mt-3 text-sm text-neutral-500">
                        此会话暂无消息
                    </p>
                </div>

                <section v-else class="space-y-6">
                    <ShareMessage v-for="message in conversation.messages" :key="message.id" :message="message" />
                </section>
            </template>
        </main>

        <ShareLoginPrompt v-if="conversation" />

        <SharePageFooter />
    </div>
</template>