<script setup lang="ts">
import { computed } from 'vue';
import { Bot } from '@lucide/vue';

import type { ChatMessage } from '../types';

const props = defineProps<{
    message: ChatMessage
}>()

const isUser = computed(()=> props.message.role === 'user')
</script>

<template>
 <article
    class="flex w-full gap-3"
    :class="isUser? 'justify-end':'justify-start'"
>
    <div
        v-if="!isUser"
        class="flex h-8 w-8 shrink-0 items-center justify-center rounded-md bg-neutral-950 text-white dark:bg-white dark:text-neutral-950"
    >
        <Bot :size="17"/>
    </div>

    <div
        class="max-w-[80%] whitespace-pre-wrap break-words py-3 text-sm leading-6"
        :class="
            isUser
            ? 'rounded-lg bg-neutral-100 px-4 dark:bg-neutral-800'
            : 'px-0'
        "
    >
    {{ message.content }}
    </div>
</article>
</template>

<style>
 
</style>