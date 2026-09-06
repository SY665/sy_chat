<script setup lang="ts">
import { computed, nextTick, onMounted, ref } from 'vue'
import { Check, X } from '@lucide/vue'

const props = withDefaults(
    defineProps<{
        originalContent: string
        disabled?: boolean
    }>(),
    {
        disabled: false,
    },
)

const emit = defineEmits<{
    cancel: []
    submit: [content: string]
}>()

const content = ref(props.originalContent)
const textareaElement = ref<HTMLTextAreaElement | null>(null)

const canSubmit = computed(() => {
    const value = content.value.trim()

    return (
        !props.disabled &&
        value.length > 0 &&
        value !== props.originalContent.trim()
    )
})

onMounted(async () => {
    await nextTick()

    textareaElement.value?.focus()
    textareaElement.value?.setSelectionRange(
        content.value.length,
        content.value.length,
    )
})

function submit() {
    if (!canSubmit.value) {
        return
    }

    emit('submit', content.value.trim())
}

function handleKeydown(event: KeyboardEvent) {
    if (event.key === 'Escape') {
        event.preventDefault()
        emit('cancel')
        return
    }

    if (event.key === 'Enter' && !event.shiftKey) {
        event.preventDefault()
        submit()
    }
}
</script>

<template>
    <div
        class="w-full min-w-64 rounded-lg border border-neutral-300 bg-white p-3 shadow-sm dark:border-neutral-700 dark:bg-neutral-900">
        <textarea ref="textareaElement" v-model="content" :disabled="disabled" :maxlength="20000" rows="4"
            class="block min-h-24 w-full resize-y bg-transparent text-sm leading-6 outline-none placeholder:text-neutral-400 disabled:cursor-not-allowed disabled:opacity-60"
            aria-label="编辑消息内容" @keydown="handleKeydown" />

        <div class="mt-3 flex justify-end gap-2">
            <button type="button"
                class="flex h-8 items-center gap-1.5 rounded-md px-3 text-sm text-neutral-600 hover:bg-neutral-100 disabled:opacity-50 dark:text-neutral-300 dark:hover:bg-neutral-800"
                :disabled="disabled" @click="emit('cancel')">
                <X :size="15" aria-hidden="true" />
                <span>取消</span>
            </button>

            <button type="button"
                class="flex h-8 items-center gap-1.5 rounded-md bg-neutral-950 px-3 text-sm text-white hover:bg-neutral-800 disabled:cursor-not-allowed disabled:opacity-40 dark:bg-white dark:text-neutral-950 dark:hover:bg-neutral-200"
                :disabled="!canSubmit" @click="submit">
                <Check :size="15" aria-hidden="true" />
                <span>发送</span>
            </button>
        </div>
    </div>
</template>