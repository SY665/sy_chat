<script setup lang="ts">
import { computed, ref } from 'vue';
import { Send } from '@lucide/vue';

const emit = defineEmits<{
    send: [content: string]
}>()

const content = ref('')

// 标记中文输入法是否正在选字，避免按 Enter 时误发送。
const isComposing = ref(false)

// 只有输入非空内容时，发送按钮才可用。
const canSend = computed(() => content.value.trim().length > 0)

function submitMessage() {
    const value = content.value.trim()

    if (!value) {
        return
    }
    emit('send', value)
    content.value = ''
}

function handleEnter() {
    if (!isComposing.value) {
        submitMessage()
    }
}
</script>

<template>
    <form class="w-full" @submit.prevent="submitMessage">
        <div
            class="flex min-h-14 items-end gap-2 rounded-lg border border-neutral-300 bg-neutral-100 p-2 shadow-sm
            dark:border-neutral-700 dark:bg-neutral-800"
            >
        <textarea 
            v-model="content" 
            rows="1"
            class="max-h-40 min-h-10 flex-1 resize-none bg-transparent px-2 py-2 text-sm outline-none placeholder:text-neutral-500"
            placeholder="给 SY Chat 发送消息" 
            @compositionstart="isComposing = true" 
            @compositionend="isComposing = false"
            @keydown.enter.exact.prevent="handleEnter" 
            />
        <button type="submit"
            class="flex h-9 w-9 shrink-0 items-center justify-center rounded-md bg-neutral-950 text-white transition-opacity hover:opacity-80 disabled:cursor-not-allowed disabled:opacity-40 dark:bg-white dark:text-neutral-950"
            :disabled="!canSend" 
            aria-label="发送消息"
            >
            <Send :size="17" />
        </button>
        </div>
    </form>
</template>