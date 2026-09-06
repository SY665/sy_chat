<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted } from 'vue'
import { storeToRefs } from 'pinia'
import {
    Check,
    Copy,
    Link2,
    LoaderCircle,
    Unlink,
    X,
} from '@lucide/vue'

import { useClipboard } from '@/composables/useClipboard'
import { useConversationStore } from '@/features/conversation/stores/conversation'

const props = defineProps<{
    open: boolean
    conversationId: string
}>()

const emit = defineEmits<{
    close: []
}>()

const conversationStore = useConversationStore()
const { currentConversation, isSharing, shareError } =
    storeToRefs(conversationStore)
const { isCopying, copyFeedback, copyText } = useClipboard()

const shareToken = computed(() => {
    if (currentConversation.value?.id !== props.conversationId) {
        return ''
    }

    return currentConversation.value.shareToken ?? ''
})

const shareUrl = computed(() => {
    if (!shareToken.value) return ''

    return `${window.location.origin}/share/${encodeURIComponent(shareToken.value)}`
})

async function handleShare() {
    if (!props.conversationId) return

    try {
        await conversationStore.share(props.conversationId)
    } catch {
        // Store 保存错误，弹窗负责展示。
    }
}

async function handleUnshare() {
    if (!props.conversationId) return

    try {
        await conversationStore.unshare(props.conversationId)
    } catch {
        // Store 保存错误，弹窗负责展示。
    }
}

function handleKeydown(event: KeyboardEvent) {
    if (props.open && event.key === 'Escape') {
        emit('close')
    }
}

onMounted(() => window.addEventListener('keydown', handleKeydown))
onBeforeUnmount(() => window.removeEventListener('keydown', handleKeydown))
</script>

<template>
    <Teleport to="body">
        <div v-if="open" class="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4"
            @click.self="emit('close')">
            <section class="w-full max-w-md rounded-lg bg-white shadow-xl dark:bg-neutral-900" role="dialog"
                aria-modal="true" aria-labelledby="share-title">
                <header class="flex h-14 items-center border-b border-neutral-200 px-4 dark:border-neutral-800">
                    <h2 id="share-title" class="flex-1 text-base font-semibold">
                        分享会话
                    </h2>

                    <button type="button"
                        class="flex h-8 w-8 items-center justify-center rounded-md hover:bg-neutral-100 dark:hover:bg-neutral-800"
                        title="关闭" aria-label="关闭分享弹窗" @click="emit('close')">
                        <X :size="18" />
                    </button>
                </header>

                <div class="space-y-4 p-4">
                    <template v-if="shareUrl">
                        <div class="flex items-center gap-2">
                            <input :value="shareUrl" readonly
                                class="h-9 min-w-0 flex-1 rounded-md border border-neutral-300 bg-neutral-50 px-3 text-xs dark:border-neutral-700 dark:bg-neutral-800"
                                aria-label="分享链接">

                            <button type="button"
                                class="flex h-9 w-9 shrink-0 items-center justify-center rounded-md bg-neutral-950 text-white disabled:opacity-50 dark:bg-white dark:text-neutral-950"
                                :disabled="isCopying" title="复制分享链接" aria-label="复制分享链接" @click="copyText(shareUrl)">
                                <Check v-if="copyFeedback === '已复制'" :size="16" />
                                <Copy v-else :size="16" />
                            </button>
                        </div>

                        <div class="flex items-center justify-between gap-3">
                            <span class="text-xs text-neutral-500" role="status">
                                {{ copyFeedback || '分享链接已启用' }}
                            </span>

                            <button type="button"
                                class="flex h-9 items-center gap-2 rounded-md px-3 text-sm text-red-600 hover:bg-red-50 disabled:opacity-50 dark:text-red-400 dark:hover:bg-red-950"
                                :disabled="isSharing" @click="handleUnshare">
                                <LoaderCircle v-if="isSharing" class="animate-spin" :size="16" />
                                <Unlink v-else :size="16" />
                                取消分享
                            </button>
                        </div>
                    </template>

                    <button v-else type="button"
                        class="flex h-10 w-full items-center justify-center gap-2 rounded-md bg-neutral-950 px-4 text-sm text-white disabled:opacity-50 dark:bg-white dark:text-neutral-950"
                        :disabled="isSharing" @click="handleShare">
                        <LoaderCircle v-if="isSharing" class="animate-spin" :size="16" />
                        <Link2 v-else :size="16" />
                        生成分享链接
                    </button>

                    <p v-if="shareError" class="text-sm text-red-600 dark:text-red-400" role="alert">
                        {{ shareError }}
                    </p>
                </div>
            </section>
        </div>
    </Teleport>
</template>