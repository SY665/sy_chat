<script setup lang="ts">
import { computed } from 'vue'
import { LoaderCircle, Search } from '@lucide/vue'

import type {
    SearchSource,
    ToolStreamEvent,
} from '../types'

const props = defineProps<{
    events: ToolStreamEvent[]
}>()

const searchEvents = computed(() => {
    return props.events.filter((event) => {
        return event.name === 'web_search'
    })
})

const isRunning = computed(() => {
    return searchEvents.value.some((event) => {
        return event.status === 'running'
    })
})

const sources = computed(() => {
    const uniqueSources = new Map<string, SearchSource>()

    for (const event of searchEvents.value) {
        for (const source of event.sources) {
            if (!isSafeSourceURL(source.url)) {
                continue
            }

            uniqueSources.set(source.url, source)
        }
    }

    return [...uniqueSources.values()]
})

function isSafeSourceURL(value: string): boolean {
    try {
        const url = new URL(value)
        return url.protocol === 'http:' || url.protocol === 'https:'
    } catch {
        return false
    }
}

function sourceHost(value: string): string {
    try {
        return new URL(value).hostname.replace(/^www\./, '')
    } catch {
        return ''
    }
}
</script>

<template>
    <section v-if="searchEvents.length > 0"
        class="my-2 w-full rounded-md border border-neutral-200 bg-neutral-50 p-3 dark:border-neutral-700 dark:bg-neutral-900">
        <div class="flex items-center gap-2 text-xs font-medium text-neutral-600 dark:text-neutral-300" role="status"
            aria-live="polite">
            <LoaderCircle v-if="isRunning" :size="15" class="animate-spin" aria-hidden="true" />
            <Search v-else :size="15" aria-hidden="true" />

            <span>
                {{
                    isRunning
                        ? '正在联网搜索...'
                        : sources.length > 0
                            ? `已检索 ${sources.length} 个来源`
                            : '联网搜索已完成'
                }}
            </span>
        </div>

        <div v-if="sources.length > 0" class="mt-2 grid gap-1.5">
            <a v-for="source in sources" :key="source.url" :href="source.url" target="_blank" rel="noopener noreferrer"
                class="min-w-0 rounded px-2 py-1.5 text-xs transition-colors hover:bg-neutral-200 dark:hover:bg-neutral-800">
                <span class="block truncate font-medium text-neutral-800 dark:text-neutral-100">
                    {{ source.title || source.url }}
                </span>

                <span class="block truncate text-neutral-500 dark:text-neutral-400">
                    {{ sourceHost(source.url) }}
                </span>
            </a>
        </div>
    </section>
</template>