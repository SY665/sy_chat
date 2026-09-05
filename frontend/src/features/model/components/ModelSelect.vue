<script setup lang="ts">
import { onMounted } from 'vue'
import { storeToRefs } from 'pinia'
import { RotateCcw } from '@lucide/vue'
import { useModelStore } from '../stores/model'

defineProps<{
    disabled?: boolean
    fullWidth?: boolean
}>()

const modelStore = useModelStore()
const { models, selectedModelId, isLoading, error } = storeToRefs(modelStore)

async function loadModels() {
    try {
        await modelStore.loadModels()
    } catch {
        // Store 保存错误，模板负责显示并提供重试入口。
    }
}

function handleChange(event: Event) {
    modelStore.selectModel((event.target as HTMLSelectElement).value)
}

onMounted(() => {
    if (models.value.length === 0) {
        void loadModels()
    }
})
</script>

<template>
    <div class="flex min-w-0 flex-col gap-1">
        <div class="flex min-w-0 items-center gap-2">
            <select :value="selectedModelId" :disabled="disabled || isLoading || !!error || models.length === 0"
                class="h-8 min-w-0 rounded-md border border-neutral-300 bg-white px-2 text-xs disabled:opacity-50 dark:border-neutral-700 dark:bg-neutral-900"
                :class="fullWidth ? 'w-full' : 'w-full max-w-64'" aria-label="选择模型" @change="handleChange">
                <option v-if="isLoading" :value="selectedModelId">加载模型中...</option>
                <option v-else-if="error" :value="selectedModelId">模型加载失败</option>
                <option v-else-if="models.length === 0" value="">暂无可用模型</option>
                <template v-else>
                    <option v-for="model in models" :key="model.id" :value="model.id">
                        {{ model.name }}
                    </option>
                </template>
            </select>
            <button v-if="error" type="button" :disabled="disabled || isLoading"
                class="flex h-8 w-8 shrink-0 items-center justify-center rounded-md hover:bg-neutral-100 disabled:opacity-50 dark:hover:bg-neutral-800"
                title="重新加载模型" aria-label="重新加载模型" @click="loadModels">
                <RotateCcw :size="14" />
            </button>
        </div>
        <p v-if="error" class="text-xs text-red-600 dark:text-red-400" role="alert">{{ error }}</p>
    </div>
</template>