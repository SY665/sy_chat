<script setup lang="ts">
import { computed } from 'vue'
import { BrainCircuit } from '@lucide/vue'
import { storeToRefs } from 'pinia'

import { useModelStore } from '../stores/model'

const props = defineProps<{
    disabled?: boolean
}>()

const modelStore = useModelStore()
const {
    effectiveThinkingEnabled,
    selectedModelSupportsThinking,
} = storeToRefs(modelStore)

// 页面状态和模型能力都会影响开关是否允许操作。
const isDisabled = computed(() => {
    return props.disabled || !selectedModelSupportsThinking.value
})

function toggleThinking() {
    if (isDisabled.value) {
        return
    }

    modelStore.setThinkingEnabled(!effectiveThinkingEnabled.value)
}
</script>

<template>
    <button type="button" role="switch"
        class="inline-flex h-8 items-center gap-2 rounded-md px-2 text-xs transition-colors hover:bg-neutral-100 disabled:cursor-not-allowed disabled:opacity-50 dark:hover:bg-neutral-800"
        :aria-checked="effectiveThinkingEnabled" :disabled="isDisabled" :title="selectedModelSupportsThinking
            ? '开启或关闭深度思考'
            : '当前模型不支持深度思考'" @click="toggleThinking">
        <BrainCircuit :size="15" />

        <span>深度思考</span>

        <span aria-hidden="true" class="relative h-4 w-7 rounded-full transition-colors" :class="effectiveThinkingEnabled
            ? 'bg-neutral-950 dark:bg-white'
            : 'bg-neutral-300 dark:bg-neutral-600'">
            <span class="absolute top-0.5 h-3 w-3 rounded-full transition-transform" :class="effectiveThinkingEnabled
                ? 'translate-x-3 bg-white dark:bg-neutral-950'
                : 'translate-x-0.5 bg-white'" />
        </span>
    </button>
</template>