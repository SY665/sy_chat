<script setup lang="ts">
import { computed } from 'vue'
import {
    Bot,
    User,
} from '@lucide/vue'

import GeneratedImageResult from '@/features/chat/components/GeneratedImageResult.vue'
import MarkdownContent from '@/features/chat/components/MarkdownContent.vue'
import SearchSources from '@/features/chat/components/SearchSources.vue'
import ThinkingPanel from '@/features/chat/components/ThinkingPanel.vue'
import type { PublicShareMessage } from '@/features/conversation/types'

const props = defineProps<{
    message: PublicShareMessage
}>()

const isUser = computed(() => props.message.role === 'user')

const generatedImages = computed(() => {
    return props.message.toolEvents?.flatMap((event) => {
        return event.image ? [event.image] : []
    }) ?? []
})

const messageTime = computed(() => {
    const timestamp = Date.parse(props.message.createdAt)

    if (Number.isNaN(timestamp)) {
        return ''
    }

    return new Intl.DateTimeFormat('zh-CN', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
    }).format(timestamp)
})
</script>

<template>
    <article :id="`message-${message.id}`" class="rounded-lg border p-4 sm:p-6" :class="isUser
        ? 'border-emerald-200 bg-emerald-50/60 dark:border-emerald-900 dark:bg-emerald-950/20'
        : 'border-neutral-200 bg-neutral-50 dark:border-neutral-800 dark:bg-neutral-900/40'
        ">
        <header class="mb-4 flex flex-wrap items-center justify-between gap-2">
            <div class="flex items-center gap-2">
                <User v-if="isUser" :size="19" class="text-emerald-700 dark:text-emerald-400" aria-hidden="true" />

                <Bot v-else :size="19" class="text-neutral-600 dark:text-neutral-400" aria-hidden="true" />

                <span class="text-sm font-semibold" :class="isUser
                    ? 'text-emerald-900 dark:text-emerald-200'
                    : 'text-neutral-900 dark:text-neutral-200'
                    ">
                    {{ isUser ? '用户' : '助手' }}
                </span>
            </div>

            <time v-if="messageTime" :datetime="message.createdAt"
                class="text-xs text-neutral-500 dark:text-neutral-400">
                {{ messageTime }}
            </time>
        </header>

        <ThinkingPanel v-if="!isUser && message.thinking" :content="message.thinking" class="mb-4" />

        <div v-if="isUser" class="whitespace-pre-wrap break-words text-sm leading-7">
            {{ message.content }}
        </div>

        <div v-else class="min-w-0">
            <SearchSources v-if="message.toolEvents?.length" :events="message.toolEvents" />

            <GeneratedImageResult v-for="image in generatedImages" :key="`${message.id}-${image.url}`" :image="image" />

            <MarkdownContent v-if="message.content" :content="message.content" class="text-sm" />
        </div>
    </article>
</template>