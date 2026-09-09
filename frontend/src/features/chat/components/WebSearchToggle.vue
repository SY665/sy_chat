<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { computed } from 'vue'

import { useModelStore } from '@/features/model/stores/model'

const props = defineProps<{
    disabled?: boolean
}>()

const modelStore = useModelStore()
const {
    effectiveWebSearchEnabled,
    webSearchAvailable,
} = storeToRefs(modelStore)

const isDisabled = computed(() => {
    return props.disabled || !webSearchAvailable.value
})

function toggleWebSearch() {
    if (props.disabled) {
        return
    }

    modelStore.setWebSearchEnabled(!effectiveWebSearchEnabled.value)
}
</script>

<template>
    <button type="button" role="switch"
        class="inline-flex h-8 items-center gap-2 rounded-md px-2 text-xs transition-colors hover:bg-neutral-100 disabled:cursor-not-allowed disabled:opacity-50 dark:hover:bg-neutral-800"
        :aria-checked="effectiveWebSearchEnabled" :disabled="isDisabled" :title="webSearchAvailable
            ? '开启或关闭联网搜索'
            : '服务端尚未配置联网搜索'" @click="toggleWebSearch">
        <span>联网搜索</span>

        <span aria-hidden="true" class="relative h-4 w-7 shrink-0 overflow-hidden rounded-full transition-colors"
            :class="effectiveWebSearchEnabled
                ? 'bg-blue-600 dark:bg-blue-500'
                : 'bg-neutral-300 dark:bg-neutral-600'">
            <span class="absolute left-0 top-0.5 h-3 w-3 rounded-full bg-white transition-transform duration-200"
                :class="effectiveWebSearchEnabled
                    ? 'translate-x-3.5'
                    : 'translate-x-0.5'" />
        </span>
    </button>
</template>