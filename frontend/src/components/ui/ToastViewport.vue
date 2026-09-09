<script setup lang="ts">
import type { Component } from 'vue'
import { storeToRefs } from 'pinia'
import {
    CircleAlert,
    CircleCheck,
    Info,
    X,
} from '@lucide/vue'

import type { ToastTone } from '@/stores/toast'
import { useToastStore } from '@/stores/toast'

const toastStore = useToastStore()
const { toasts } = storeToRefs(toastStore)

const iconByTone: Record<ToastTone, Component> = {
    success: CircleCheck,
    error: CircleAlert,
    info: Info,
}

const iconClassByTone: Record<ToastTone, string> = {
    success: 'text-emerald-600 dark:text-emerald-400',
    error: 'text-red-600 dark:text-red-400',
    info: 'text-blue-600 dark:text-blue-400',
}

const borderClassByTone: Record<ToastTone, string> = {
    success: 'border-l-emerald-500',
    error: 'border-l-red-500',
    info: 'border-l-blue-500',
}
</script>

<template>
    <Teleport to="body">
        <div class="pointer-events-none fixed right-4 top-4 z-[70] flex w-[calc(100%-2rem)] max-w-sm flex-col gap-2 sm:right-5 sm:top-5"
            aria-live="polite" aria-atomic="false">
            <TransitionGroup name="toast">
                <article v-for="toast in toasts" :key="toast.id"
                    class="pointer-events-auto flex min-h-14 items-start gap-3 rounded-lg border border-l-4 border-neutral-200 bg-white px-4 py-3 text-neutral-950 shadow-lg dark:border-neutral-700 dark:bg-neutral-900 dark:text-white"
                    :class="borderClassByTone[toast.tone]" :role="toast.tone === 'error' ? 'alert' : 'status'">
                    <component :is="iconByTone[toast.tone]" :size="18" class="mt-0.5 shrink-0"
                        :class="iconClassByTone[toast.tone]" aria-hidden="true" />

                    <p class="min-w-0 flex-1 text-sm leading-5">
                        {{ toast.message }}
                    </p>

                    <button type="button"
                        class="flex h-7 w-7 shrink-0 items-center justify-center rounded-md text-neutral-500 hover:bg-neutral-100 hover:text-neutral-950 dark:hover:bg-neutral-800 dark:hover:text-white"
                        title="关闭提示" aria-label="关闭提示" @click="toastStore.remove(toast.id)">
                        <X :size="15" />
                    </button>
                </article>
            </TransitionGroup>
        </div>
    </Teleport>
</template>

<style scoped>
.toast-enter-active,
.toast-leave-active {
    transition:
        opacity 160ms ease,
        transform 160ms ease;
}

.toast-enter-from,
.toast-leave-to {
    opacity: 0;
    transform: translateY(-8px);
}

.toast-move {
    transition: transform 160ms ease;
}
</style>