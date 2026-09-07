<script setup lang="ts">
import { computed, ref } from 'vue'
import { useClipboard } from '@/composables/useClipboard'
import {
    Bot,
    Check,
    Copy,
    CircleAlert,
    CircleStop,
    FileText,
    RotateCcw,
    Pencil,
} from '@lucide/vue'

import type { ChatMessage } from '../types'
import ThinkingPanel from './ThinkingPanel.vue'
import MessageEditor from './MessageEditor.vue'
import MarkdownContent from './MarkdownContent.vue'

const props = withDefaults(
    defineProps<{
        message: ChatMessage
        retryDisabled?: boolean
    }>(),
    {
        retryDisabled: false,
    },
)

const emit = defineEmits<{
    retry: [message: ChatMessage]
    edit: [message: ChatMessage, content: string]
}>()

const isUser = computed(() => props.message.role === 'user')
const hasFailed = computed(
    () => props.message.status === 'failed',
)
const hasStopped = computed(
    () => props.message.status === 'stopped',
)

const canRetryPersistedAssistant = computed(
    () =>
        !isUser.value &&
        props.message.status === undefined,
)

const isEditing = ref(false)

const canEditPersistedUser = computed(
    () =>
        isUser.value &&
        props.message.status === undefined,
)

function startEditing() {
    if (
        props.retryDisabled ||
        !canEditPersistedUser.value
    ) {
        return
    }

    isEditing.value = true
}

function cancelEditing() {
    isEditing.value = false
}

function submitEditing(content: string) {
    isEditing.value = false
    emit('edit', props.message, content)
}

function formatFileSize(size: number): string {
    return `${(size / 1024).toFixed(1)} KB`
}

const { isCopying, copyFeedback, copyText } = useClipboard()

function copyMessage() {
    return copyText(props.message.content)
}
</script>

<template>
    <article class="flex w-full gap-3" :class="isUser ? 'justify-end' : 'justify-start'">
        <div v-if="!isUser"
            class="flex h-8 w-8 shrink-0 items-center justify-center rounded-md bg-neutral-950 text-white dark:bg-white dark:text-neutral-950"
            aria-label="SY Chat">
            <Bot :size="17" aria-hidden="true" />
        </div>

        <div class="flex min-w-0 max-w-[80%] flex-col gap-1.5" :class="isUser ? 'items-end' : 'items-start'">
            <div v-if="isUser && message.attachments?.length" class="flex max-w-full flex-wrap justify-end gap-2">
                <div v-for="(file, index) in message.attachments" :key="`${file.name}-${index}`"
                    class="flex min-w-0 items-center gap-1.5 rounded-md border px-2.5 py-1.5 text-xs" :class="file.type === 'md'
                            ? 'border-orange-200 bg-orange-50 text-orange-700 dark:border-orange-800 dark:bg-orange-950 dark:text-orange-300'
                            : 'border-blue-200 bg-blue-50 text-blue-700 dark:border-blue-800 dark:bg-blue-950 dark:text-blue-300'
                        " :title="file.name">
                    <FileText :size="14" class="shrink-0" aria-hidden="true" />

                    <span class="max-w-48 truncate font-medium">
                        {{ file.name }}
                    </span>

                    <span class="shrink-0 opacity-70">
                        {{ formatFileSize(file.size) }}
                    </span>
                </div>
            </div>

            <MessageEditor v-if="isUser && isEditing" :original-content="message.content" :disabled="retryDisabled"
                @cancel="cancelEditing" @submit="submitEditing" />

            <!-- 用户消息保留原始文本，AI 回复按 Markdown 展示。 -->
            <div v-else-if="isUser"
                class="max-w-full whitespace-pre-wrap break-words rounded-lg bg-neutral-100 px-4 py-3 text-sm leading-6 dark:bg-neutral-800">
                {{ message.content }}
            </div>

            <div v-else class="w-full min-w-0">
                <ThinkingPanel v-if="message.thinking" :content="message.thinking" :default-open="retryDisabled" />

                <MarkdownContent v-if="message.content" :content="message.content"
                    class="w-full min-w-0 py-3 text-sm" />
            </div>

            <div v-if="message.content && !isEditing" class="flex max-w-full items-center gap-2">
                <button type="button"
                    class="flex h-7 w-7 shrink-0 items-center justify-center rounded-md text-neutral-500 hover:bg-neutral-100 disabled:opacity-50 dark:hover:bg-neutral-800"
                    :disabled="isCopying" title="复制消息" aria-label="复制消息" @click="copyMessage">
                    <Check v-if="copyFeedback === '已复制'" :size="14" />
                    <Copy v-else :size="14" />
                </button>

                <button v-if="canEditPersistedUser" type="button"
                    class="flex h-7 w-7 shrink-0 items-center justify-center rounded-md text-neutral-500 hover:bg-neutral-100 disabled:opacity-50 dark:hover:bg-neutral-800"
                    :disabled="retryDisabled" title="编辑消息" aria-label="编辑用户消息" @click="startEditing">
                    <Pencil :size="14" aria-hidden="true" />
                </button>

                <button v-if="canRetryPersistedAssistant" type="button"
                    class="flex h-7 w-7 shrink-0 items-center justify-center rounded-md text-neutral-500 hover:bg-neutral-100 disabled:opacity-50 dark:hover:bg-neutral-800"
                    :disabled="retryDisabled" title="重新生成" aria-label="重新生成 AI 回复" @click="emit('retry', message)">
                    <RotateCcw :size="14" aria-hidden="true" />
                </button>
                <span class="text-xs text-neutral-500" role="status">
                    {{ copyFeedback }}
                </span>
            </div>

            <div v-if="isUser && (hasFailed || hasStopped)" class="flex items-center gap-2 text-xs" :class="hasFailed
                ? 'text-red-600 dark:text-red-400'
                : 'text-neutral-500 dark:text-neutral-400'
                ">
                <CircleAlert v-if="hasFailed" :size="14" aria-hidden="true" />
                <CircleStop v-else :size="14" aria-hidden="true" />

                <span>
                    {{ hasFailed ? '发送失败' : '已停止' }}
                </span>

                <button type="button" class="flex items-center gap-1 hover:text-neutral-950 dark:hover:text-white"
                    :disabled="retryDisabled" aria-label="重新发送消息" @click="emit('retry', message)">
                    <RotateCcw :size="13" aria-hidden="true" />
                    <span>重试</span>
                </button>
            </div>
        </div>
    </article>
</template>

<style></style>