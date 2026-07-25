<script setup lang="ts">
import {
    MessageSquare,
    Moon,
    PanelLeftClose,
    PanelLeftOpen,
    Plus,
    Sun,
} from '@lucide/vue'
import { storeToRefs } from 'pinia'

import { useAppStore } from '@/stores/app'

defineProps<{
    collapsed: boolean
}>()

const emit = defineEmits<{
    toggle: []
}>()

const appStore = useAppStore()
const { isDark } = storeToRefs(appStore)
</script>

<template>
    <aside
        class="hidden shrink-0 flex-col border-r border-neutral-200 bg-neutral-50 transition-[width] duration-200 dark:border-neutral-800 dark:bg-neutral-900 md:flex"
        :class="collapsed ? 'w-16' : 'w-64'">
        <div class="flex h-14 shrink-0 items-center border-b border-neutral-200 px-3 dark:border-neutral-800"
            :class="collapsed ? 'justify-center' : 'gap-3'">
            <div
                class="flex h-8 w-8 shrink-0 items-center justify-center rounded-md bg-neutral-950 text-white dark:bg-white dark:text-neutral-950">
                <MessageSquare :size="18" />
            </div>

            <span v-if="!collapsed" class="min-w-0 flex-1 truncate text-sm font-semibold">
                SY Chat
            </span>

            <button v-if="!collapsed" type="button"
                class="flex h-8 w-8 shrink-0 items-center justify-center rounded-md text-neutral-500 hover:bg-neutral-200 hover:text-neutral-950 dark:hover:bg-neutral-800 dark:hover:text-white"
                title="收起侧边栏" @click="emit('toggle')">
                <PanelLeftClose :size="17" />
            </button>
        </div>

        <div class="p-3">
            <button type="button"
                class="flex h-10 w-full items-center rounded-md text-sm transition-colors hover:bg-neutral-200 dark:hover:bg-neutral-800"
                :class="collapsed ? 'justify-center' : 'gap-3 px-3'" title="新对话">
                <Plus :size="18" />
                <span v-if="!collapsed">新对话</span>
            </button>
        </div>

        <div v-if="collapsed" class="px-3">
            <button type="button"
                class="flex h-10 w-full items-center justify-center rounded-md text-neutral-500 hover:bg-neutral-200 hover:text-neutral-950 dark:hover:bg-neutral-800 dark:hover:text-white"
                title="展开侧边栏" @click="emit('toggle')">
                <PanelLeftOpen :size="18" />
            </button>
        </div>

        <div v-if="!collapsed" class="min-h-0 flex-1 overflow-y-auto px-3 py-2">
            <p class="px-3 pb-2 text-xs font-medium text-neutral-500">
                最近
            </p>

            <button type="button"
                class="w-full truncate rounded-md px-3 py-2 text-left text-sm hover:bg-neutral-200 dark:hover:bg-neutral-800">
                欢迎使用 SY Chat
            </button>
        </div>

        <div class="mt-auto border-t border-neutral-200 p-3 dark:border-neutral-800">
            <button type="button"
                class="flex h-10 w-full items-center rounded-md text-sm transition-colors hover:bg-neutral-200 dark:hover:bg-neutral-800"
                :class="collapsed ? 'justify-center' : 'gap-3 px-3'" :title="isDark ? '切换到浅色模式' : '切换到深色模式'"
                @click="appStore.toggleTheme">
                <Sun v-if="isDark" :size="18" />
                <Moon v-else :size="18" />

                <span v-if="!collapsed">
                    {{ isDark ? '浅色模式' : '深色模式' }}
                </span>
            </button>
        </div>
    </aside>
</template>