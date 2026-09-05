<script setup lang="ts">
import { onBeforeUnmount, onMounted } from 'vue'
import { storeToRefs } from 'pinia'
import { Monitor, Moon, Sun, X } from '@lucide/vue'

import { useAppStore } from '@/stores/app'
import type { ThemePreference } from '@/stores/app'
import ModelSelect from '@/features/model/components/ModelSelect.vue'

defineProps<{
    open: boolean
}>()

const emit = defineEmits<{
    close: []
}>()

const appStore = useAppStore()
const { themePreference } = storeToRefs(appStore)

const themeOptions: Array<{
    value: ThemePreference
    label: string
}> = [
        { value: 'light', label: '浅色' },
        { value: 'dark', label: '深色' },
        { value: 'system', label: '跟随系统' },
    ]

function handleKeydown(event: KeyboardEvent) {
    if (event.key === 'Escape') {
        emit('close')
    }
}

onMounted(() => window.addEventListener('keydown', handleKeydown))
onBeforeUnmount(() => window.removeEventListener('keydown', handleKeydown))
</script>

<template>
    <Teleport to="body">
        <div v-if="open" class="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4"
            @click.self="emit('close')">
            <section class="w-full max-w-md rounded-lg bg-white shadow-xl dark:bg-neutral-900" role="dialog"
                aria-modal="true" aria-labelledby="settings-title">
                <header class="flex h-14 items-center border-b border-neutral-200 px-4 dark:border-neutral-800">
                    <h2 id="settings-title" class="flex-1 text-base font-semibold">
                        设置
                    </h2>

                    <button type="button"
                        class="flex h-8 w-8 items-center justify-center rounded-md hover:bg-neutral-100 dark:hover:bg-neutral-800"
                        title="关闭设置" aria-label="关闭设置" @click="emit('close')">
                        <X :size="18" />
                    </button>
                </header>

                <div class="space-y-5 p-4">
                    <section>
                        <h3 class="mb-2 text-sm font-medium">主题</h3>

                        <div class="grid grid-cols-3 rounded-lg bg-neutral-100 p-1 dark:bg-neutral-800">
                            <button v-for="option in themeOptions" :key="option.value" type="button"
                                class="flex h-9 items-center justify-center gap-1.5 rounded-md text-xs transition-colors"
                                :class="themePreference === option.value
                                    ? 'bg-white text-neutral-950 shadow-sm dark:bg-neutral-700 dark:text-white'
                                    : 'text-neutral-500 hover:text-neutral-950 dark:hover:text-white'"
                                :aria-pressed="themePreference === option.value"
                                @click="appStore.setThemePreference(option.value)">
                                <Sun v-if="option.value === 'light'" :size="15" />
                                <Moon v-else-if="option.value === 'dark'" :size="15" />
                                <Monitor v-else :size="15" />
                                <span>{{ option.label }}</span>
                            </button>
                        </div>
                    </section>

                    <section>
                        <h3 class="mb-2 text-sm font-medium">默认模型</h3>
                        <ModelSelect full-width />
                    </section>
                </div>
            </section>
        </div>
    </Teleport>
</template>