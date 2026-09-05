<script setup lang="ts">
import { nextTick, ref, watch,computed } from 'vue';

import MessageItem from './MessageItem.vue';
import type { ChatMessage } from '../types';
import { LoaderCircle } from '@lucide/vue';

const props = withDefaults(
    defineProps<{
        messages:ChatMessage[]
        isGenerating?:boolean
    }>(),
    {
        isGenerating:false
    }
)

const emit = defineEmits<{
  retry: [message: ChatMessage]
}>()

const bottomElement = ref<HTMLElement | null>(null)

const showThinking = computed(() => {
    const lastMessage = props.messages[
        props.messages.length - 1
    ]

    // 首个 AI 片段到达后会出现 assistant 消息，此时隐藏转圈。
    return props.isGenerating &&
        lastMessage?.role !== 'assistant'
})

watch(
    () => [
        props.messages.length,
        props.messages[props.messages.length - 1]?.content,
        props.isGenerating,
    ],
    async () => {
        // 等待新消息渲染完成后再滚动。
        await nextTick()
        bottomElement.value?.scrollIntoView({
            behavior:props.isGenerating ? 'auto' : 'smooth',
            block:'end',
        })
    },
)
</script>

<template>
<div class="mx-auto flex w-full max-w-3xl flex-col gap-6 py-8">
    <MessageItem
        v-for="message in messages"
        :key="message.id"
        :message="message"
        @retry="emit('retry', $event)"
    />

    <div
        v-if="showThinking"
        class="flex items-center gap-3 text-sm text-neutral-500"
    >
        <div
            class="flex h-8 w-8 items-center justify-center rounded-md bg-neutral-950 text-white dark:bg-white dark:text-neutral-950"
        >
            <LoaderCircle :size="17" class="animate-spin" />
            
        </div>
        <span>正在思考...</span>
    </div>

    <div ref="bottomElement" aria-hidden="true"/>
</div>
</template>

<style></style>