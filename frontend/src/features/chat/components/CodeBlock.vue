<script setup lang="ts">
import { computed } from 'vue'
import { Check, Copy } from '@lucide/vue'
import hljs from 'highlight.js/lib/common'
import 'highlight.js/styles/github-dark.css'
import { useClipboard } from '@/composables/useClipboard'

const props = defineProps<{
    code: string,
    language: string
}>()

const { isCopying, copyFeedback, copyText } = useClipboard()
const languageName = computed(
    () => props.language.trim().split(/\s+/)[0] || 'plaintext',
)

const highlightedCode = computed(() => {
    if (!hljs.getLanguage(languageName.value))
        return null

    return hljs.highlight(props.code, {
        language: languageName.value,
        ignoreIllegals: true
    }).value
})
</script>

<template>
    <div class="my-3 min-w-0 overflow-hidden rounded-md border border-neutral-700">
        <div class="flex min-h-9 items-center gap-2 bg-neutral-800 px-3 text-xs text-neutral-300">
            <span class="min-w-0 flex-1 truncate" :title="languageName">
                {{ languageName }}
            </span>
            <span class="shrink-0" role="status">{{ copyFeedback }}</span>
            <button type="button"
                class="flex h-7 w-7 shrink-0 items-center justify-center rounded hover:bg-neutral-700 disabled:opacity-50"
                :disabled="isCopying || !code" title="复制代码" aria-label="复制代码" @click="copyText(code)">
                <Check v-if="copyFeedback === '已复制'" :size="14" />
                <Copy v-else :size="14" />
            </button>
        </div>
        <!-- 未识别的语言使用文本插值，确保代码中的 HTML 不被执行。 -->
        <pre
            class="m-0 overflow-x-auto bg-[#0d1117] p-4 text-sm leading-6 text-[#e6edf3]"><code v-if="highlightedCode !== null" v-html="highlightedCode"></code><code v-else>{{ code }}</code></pre>
    </div>
</template>
