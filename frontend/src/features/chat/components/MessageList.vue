<script setup lang="ts">
import { nextTick, ref, watch } from 'vue';

import MessageItem from './MessageItem.vue';
import type { ChatMessage } from '../types';

const props = defineProps<{
    messages: ChatMessage[]
}>()

const bottomElement = ref<HTMLElement | null>(null)

watch(
    () => props.messages.length,
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

    <div ref="bottomElement" aria-hidden="true"/>
</div>
</template>

<style></style>