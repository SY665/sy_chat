<script setup lang="ts">
import {
    nextTick,
    onErrorCaptured,
    ref,
} from 'vue'
import {
    RefreshCw,
    TriangleAlert,
} from '@lucide/vue'

const hasError = ref(false)
const heading = ref<HTMLHeadingElement | null>(null)

onErrorCaptured((error, _instance, info) => {
    hasError.value = true

    // 详细错误只写入控制台，避免向用户暴露内部实现信息。
    console.error('Unhandled Vue component error:', {
        error,
        info,
    })

    void nextTick(() => {
        heading.value?.focus()
    })

    // 阻止错误继续向父组件传播。
    return false
})

function retry() {
    hasError.value = false
}

function reloadPage() {
    window.location.reload()
}
</script>

<template>
    <slot v-if="!hasError" />

    <main v-else
        class="flex min-h-screen items-center justify-center bg-white px-6 py-12 text-neutral-950 dark:bg-neutral-950 dark:text-white"
        role="alert">
        <div class="w-full max-w-md text-center">
            <TriangleAlert :size="42" class="mx-auto text-red-600 dark:text-red-400" aria-hidden="true" />

            <h1 ref="heading" class="mt-5 text-xl font-semibold outline-none" tabindex="-1">
                页面出现了一些问题
            </h1>

            <p class="mt-3 text-sm leading-6 text-neutral-500 dark:text-neutral-400">
                当前页面无法继续正常显示。你可以先尝试重新渲染，仍未恢复时再刷新页面。
            </p>

            <div class="mt-6 flex justify-center gap-3">
                <button type="button"
                    class="h-10 rounded-md border border-neutral-300 px-4 text-sm font-medium hover:bg-neutral-100 dark:border-neutral-700 dark:hover:bg-neutral-800"
                    @click="retry">
                    再试一次
                </button>

                <button type="button"
                    class="flex h-10 items-center gap-2 rounded-md bg-neutral-950 px-4 text-sm font-medium text-white hover:bg-neutral-800 dark:bg-white dark:text-neutral-950 dark:hover:bg-neutral-200"
                    @click="reloadPage">
                    <RefreshCw :size="16" aria-hidden="true" />
                    刷新页面
                </button>
            </div>
        </div>
    </main>
</template>