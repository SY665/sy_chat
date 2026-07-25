<script setup lang="ts">
import { nextTick, ref, watch } from 'vue';

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

const bottomElement = ref<HTMLElement | null>(null)

watch(
    () => [props.messages.length,props.isGenerating],
    async () => {
        // 等待新消息渲染完成后再滚动。
        await nextTick()
        bottomElement.value?.scrollIntoView({
            behavior:'smooth',
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
    />

    <div
        v-if="isGenerating"
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