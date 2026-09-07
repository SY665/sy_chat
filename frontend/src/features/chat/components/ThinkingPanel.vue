<script setup lang="ts">
import { ref } from 'vue'
import { BrainCircuit, ChevronDown } from '@lucide/vue'

import MarkdownContent from './MarkdownContent.vue'

const props = withDefaults(
    defineProps<{
        content: string
        defaultOpen?: boolean
    }>(),
    {
        defaultOpen: false,
    },
)

const isOpen = ref(props.defaultOpen)
</script>

<template>
    <section class="w-full min-w-0 text-neutral-600 dark:text-neutral-400">
        <button type="button" class="flex h-8 items-center gap-2 text-xs hover:text-neutral-950 dark:hover:text-white"
            :aria-expanded="isOpen" @click="isOpen = !isOpen">
            <BrainCircuit :size="14" aria-hidden="true" />

            <span>
                {{ isOpen ? '收起思考过程' : '查看思考过程' }}
            </span>

            <ChevronDown :size="14" class="transition-transform" :class="{ 'rotate-180': isOpen }" aria-hidden="true" />
        </button>

        <div v-show="isOpen" class="border-l border-neutral-300 pl-3 dark:border-neutral-700">
            <MarkdownContent :content="content" class="text-xs leading-5" />
        </div>
    </section>
</template>