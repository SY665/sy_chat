<script setup lang="ts">
import { computed } from 'vue'
import { useClipboard } from '@/composables/useClipboard'
import {
    Bot,
    Check,
    Copy,
    CircleAlert,
    CircleStop,
    RotateCcw,
} from '@lucide/vue'

import type { ChatMessage } from '../types'
import MarkdownContent from './MarkdownContent.vue'

const props = defineProps<{
    message: ChatMessage
}>()

const emit = defineEmits<{
    retry: [message: ChatMessage]
}>()

const isUser = computed(() => props.message.role === 'user')
const hasFailed = computed(
    () => props.message.status === 'failed',
)
const hasStopped = computed(
    () => props.message.status === 'stopped',
)

const canRetry = computed(
    () => hasFailed.value || hasStopped.value,
)

const {isCopying,copyFeedback,copyText} = useClipboard()

function copyMessage(){
    return copyText(props.message.content)
}
</script>

<template>
    <article class="flex w-full gap-3" :class="isUser ? 'justify-end' : 'justify-start'">
        <div v-if="!isUser"
            class="flex h-8 w-8 shrink-0 items-center justify-center rounded-md bg-neutral-950 text-white dark:bg-white dark:text-neutral-950">
            <Bot :size="17" />
        </div>

        <div class="flex min-w-0 max-w-[80%] flex-col gap-1.5" :class="isUser ? 'items-end' : 'items-start'">
            <!-- 用户消息保留原始文本，AI 回复按 Markdown 展示。 -->
            <div v-if="isUser"
                class="max-w-full whitespace-pre-wrap break-words rounded-lg bg-neutral-100 px-4 py-3 text-sm leading-6 dark:bg-neutral-800">
                {{ message.content }}
            </div>

            <MarkdownContent v-else :content="message.content" class="w-full min-w-0 py-3 text-sm" />

            <div v-if="message.content" class="flex max-w-full items-center gap-2">
                <button type="button"
                    class="flex h-7 w-7 shrink-0 items-center justify-center rounded-md text-neutral-500 hover:bg-neutral-100 disabled:opacity-50 dark:hover:bg-neutral-800"
                    :disabled="isCopying" title="复制消息" aria-label="复制消息" @click="copyMessage">
                    <Check v-if="copyFeedback === '已复制'" :size="14" />
                    <Copy v-else :size="14" />
                </button>
                <span class="text-xs text-neutral-500" role="status">
                    {{ copyFeedback }}
                </span>
            </div>

            <div v-if="canRetry" class="flex items-center gap-2 text-xs" :class="hasFailed
                ? 'text-red-600 dark:text-red-400'
                : 'text-neutral-500 dark:text-neutral-400'
                ">
                <CircleAlert v-if="hasFailed" :size="14" aria-hidden="true" />
                <CircleStop v-else :size="14" aria-hidden="true" />

                <span>
                    {{ hasFailed ? '发送失败' : '已停止' }}
                </span>

                <button type="button" class="flex items-center gap-1 hover:text-neutral-950 dark:hover:text-white"
                    aria-label="重新发送消息" @click="emit('retry', message)">
                    <RotateCcw :size="13" aria-hidden="true" />
                    <span>重试</span>
                </button>
            </div>
        </div>
    </article>
</template>

<style></style>