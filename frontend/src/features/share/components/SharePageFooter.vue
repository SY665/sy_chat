<script setup lang="ts">
import {
    Code,
    Globe,
} from '@lucide/vue'

const repositoryURL = 'https://github.com/SY665/sy_chat'
</script>

<template>
    <footer class="border-t border-neutral-200 bg-white/95 dark:border-neutral-800 dark:bg-neutral-950/95">
        <div
            class="mx-auto flex w-full max-w-4xl flex-col items-center justify-between gap-4 px-4 py-6 sm:flex-row sm:px-6">
            <div class="text-center sm:text-left">
                <p class="text-sm text-neutral-500 dark:text-neutral-400">
                    由
                    <span class="font-medium text-neutral-700 dark:text-neutral-200">
                        SY Chat
                    </span>
                    提供支持
                </p>

                <p class="mt-1 text-xs text-neutral-500 dark:text-neutral-400">
                    这是一个只读分享页面，原始对话可能已经更新
                </p>
            </div>

            <nav class="flex items-center gap-4" aria-label="分享页相关链接">
                <RouterLink to="/"
                    class="flex items-center gap-1.5 text-sm text-neutral-500 transition-colors hover:text-neutral-950 dark:text-neutral-400 dark:hover:text-white">
                    <Globe :size="16" aria-hidden="true" />
                    <span>访问 SY Chat</span>
                </RouterLink>

                <a :href="repositoryURL" target="_blank" rel="noopener noreferrer"
                    class="flex items-center gap-1.5 text-sm text-neutral-500 transition-colors hover:text-neutral-950 dark:text-neutral-400 dark:hover:text-white">
                    <Code :size="16" aria-hidden="true" />
                    <span>GitHub</span>
                </a>
            </nav>
        </div>
    </footer>
</template>