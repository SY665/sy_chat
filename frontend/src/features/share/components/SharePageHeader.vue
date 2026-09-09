<script setup lang="ts">
import { computed } from 'vue'
import {
    Calendar,
    Clock,
    Eye,
    MessageSquareText,
    User,
} from '@lucide/vue'

const props = defineProps<{
    title: string
    ownerName: string
    sharedAt?: string
    viewCount: number
}>()

const sharedDate = computed(() => {
    if (!props.sharedAt) {
        return ''
    }

    const timestamp = Date.parse(props.sharedAt)

    if (Number.isNaN(timestamp)) {
        return ''
    }

    return new Intl.DateTimeFormat('zh-CN', {
        year: 'numeric',
        month: 'long',
        day: 'numeric',
    }).format(timestamp)
})

const relativeTime = computed(() => {
    if (!props.sharedAt) {
        return ''
    }

    const timestamp = Date.parse(props.sharedAt)

    if (Number.isNaN(timestamp)) {
        return ''
    }

    // 服务端时间略快于客户端时也按“刚刚”处理。
    const difference = Math.max(0, Date.now() - timestamp)
    const minutes = Math.floor(difference / 60_000)
    const hours = Math.floor(difference / 3_600_000)
    const days = Math.floor(difference / 86_400_000)

    if (minutes < 1) return '刚刚'
    if (minutes < 60) return `${minutes} 分钟前`
    if (hours < 24) return `${hours} 小时前`
    if (days < 30) return `${days} 天前`
    if (days < 365) return `${Math.floor(days / 30)} 个月前`

    return `${Math.floor(days / 365)} 年前`
})
</script>

<template>
    <header
        class="border-b border-neutral-200 bg-white/95 backdrop-blur dark:border-neutral-800 dark:bg-neutral-950/95">
        <div class="mx-auto w-full max-w-4xl px-4 py-4 sm:px-6">
            <div class="flex items-center justify-between gap-4">
                <RouterLink to="/" class="flex items-center gap-2 text-sm font-semibold">
                    <MessageSquareText :size="19" class="text-emerald-600 dark:text-emerald-400" aria-hidden="true" />
                    <span>SY Chat</span>
                </RouterLink>

                <span
                    class="rounded-md bg-blue-50 px-2.5 py-1 text-xs font-medium text-blue-700 dark:bg-blue-950/40 dark:text-blue-300">
                    只读模式
                </span>
            </div>

            <h1 class="mt-5 break-words text-2xl font-semibold text-neutral-950 dark:text-white">
                {{ title }}
            </h1>

            <div
                class="mt-3 flex flex-wrap items-center gap-x-4 gap-y-2 text-sm text-neutral-500 dark:text-neutral-400">
                <span class="flex items-center gap-1.5">
                    <User :size="15" aria-hidden="true" />
                    {{ ownerName }}
                </span>

                <span v-if="sharedDate" class="flex items-center gap-1.5">
                    <Calendar :size="15" aria-hidden="true" />
                    分享于 {{ sharedDate }}
                </span>

                <span v-if="relativeTime" class="flex items-center gap-1.5">
                    <Clock :size="15" aria-hidden="true" />
                    {{ relativeTime }}
                </span>

                <span class="flex items-center gap-1.5">
                    <Eye :size="15" aria-hidden="true" />
                    {{ viewCount }} 次查看
                </span>
            </div>
        </div>
    </header>
</template>