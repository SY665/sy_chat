<script setup lang="ts">
import { computed } from 'vue'
import { LoaderCircle } from '@lucide/vue'

import type { ToolStreamEvent } from '../types'

const props = defineProps<{
    events: ToolStreamEvent[]
}>()

// 图片地址返回前，通过工具运行事件展示等待状态。
const isGenerating = computed(() => {
    return props.events.some((event) => {
        return (
            event.name === 'generate_image' &&
            event.status === 'running'
        )
    })
})
</script>

<template>
    <div v-if="isGenerating" class="my-3 flex items-center gap-2 text-sm text-neutral-500 dark:text-neutral-400"
        role="status" aria-live="polite">
        <LoaderCircle :size="18" class="shrink-0 animate-spin" aria-hidden="true" />

        <span>正在生成图片...</span>
    </div>
</template>