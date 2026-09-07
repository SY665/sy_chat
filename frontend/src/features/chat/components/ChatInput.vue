<script setup lang="ts">
import { computed, ref } from 'vue'
import {
    FileText,
    FileUp,
    LoaderCircle,
    Send,
    Square,
    X,
} from '@lucide/vue'

import type { FileAttachment } from '../types'
import { uploadAttachment } from '@/lib/api/attachment'

const emit = defineEmits<{
    send: [content: string, attachments: FileAttachment[]]
    stop: []
}>()

const props = withDefaults(
    defineProps<{
        disabled?: boolean
        isGenerating?: boolean
    }>(),
    {
        disabled: false,
        isGenerating: false,
    },
)

const content = ref('')
const fileInput = ref<HTMLInputElement | null>(null)
const uploadedFiles = ref<FileAttachment[]>([])
const isUploading = ref(false)
const isDragging = ref(false)
const uploadError = ref<string | null>(null)

let dragDepth = 0

// 标记中文输入法是否正在选字，避免按 Enter 时误发送。
const isComposing = ref(false)

const canAddFile = computed(() => {
    return (
        !props.disabled &&
        !props.isGenerating &&
        !isUploading.value &&
        uploadedFiles.value.length < 5
    )
})

const canSend = computed(() => {
    return (
        !props.disabled &&
        !props.isGenerating &&
        !isUploading.value &&
        content.value.trim().length > 0
    )
})

function openFilePicker() {
    if (canAddFile.value) {
        fileInput.value?.click()
    }
}

async function uploadFiles(files: File[]) {
    if (
        files.length === 0 ||
        props.disabled ||
        props.isGenerating ||
        isUploading.value
    ) {
        return
    }

    const remainingCount = 5 - uploadedFiles.value.length
    if (remainingCount <= 0) {
        uploadError.value = '每条消息最多添加 5 个附件'
        return
    }

    const selectedFiles = files.slice(0, remainingCount)
    uploadError.value =
        files.length > remainingCount
            ? '每条消息最多添加 5 个附件'
            : null
    isUploading.value = true

    try {
        // 逐个上传，避免一次选择多个文件时瞬间产生大量请求。
        for (const file of selectedFiles) {
            try {
                const attachment = await uploadAttachment(file)
                uploadedFiles.value.push(attachment)
            } catch (error) {
                const message =
                    error instanceof Error ? error.message : '文件上传失败'
                uploadError.value = `${file.name}：${message}`
            }
        }
    } finally {
        isUploading.value = false
    }
}

async function handleFileChange(event: Event) {
    const input = event.target as HTMLInputElement
    const files = Array.from(input.files ?? [])

    // 允许删除附件后再次选择相同文件。
    input.value = ''

    await uploadFiles(files)
}

function hasDraggedFiles(event: DragEvent): boolean {
    return event.dataTransfer?.types.includes('Files') ?? false
}

function handleDragEnter(event: DragEvent) {
    if (!hasDraggedFiles(event) || !canAddFile.value) {
        return
    }

    dragDepth += 1
    isDragging.value = true
}

function handleDragOver(event: DragEvent) {
    if (!hasDraggedFiles(event) || !canAddFile.value) {
        return
    }

    if (event.dataTransfer) {
        event.dataTransfer.dropEffect = 'copy'
    }
}

function handleDragLeave() {
    dragDepth = Math.max(0, dragDepth - 1)

    if (dragDepth === 0) {
        isDragging.value = false
    }
}

async function handleDrop(event: DragEvent) {
    dragDepth = 0
    isDragging.value = false

    if (!canAddFile.value) {
        return
    }

    const files = Array.from(event.dataTransfer?.files ?? [])
    await uploadFiles(files)
}

function removeFile(index: number) {
    uploadedFiles.value.splice(index, 1)
    uploadError.value = null
}

function formatFileSize(size: number): string {
    return `${(size / 1024).toFixed(1)} KB`
}

function submitMessage() {
    if (!canSend.value) {
        return
    }

    // 使用副本发送，避免清空输入框状态时影响正在发送的消息。
    emit(
        'send',
        content.value.trim(),
        [...uploadedFiles.value],
    )

    content.value = ''
    uploadedFiles.value = []
    uploadError.value = null
}

function handleEnter() {
    if (!isComposing.value) {
        submitMessage()
    }
}
</script>

<template>
    <form class="w-full" @submit.prevent="submitMessage">
        <div class="relative rounded-lg border border-neutral-300 bg-neutral-100 p-2 shadow-sm dark:border-neutral-700 dark:bg-neutral-800"
            @dragenter.prevent="handleDragEnter" @dragover.prevent="handleDragOver" @dragleave.prevent="handleDragLeave"
            @drop.prevent="handleDrop">

            <div v-if="isDragging"
                class="pointer-events-none absolute inset-0 z-10 flex items-center justify-center rounded-lg border-2 border-dashed border-blue-500 bg-blue-50/95 text-blue-700 dark:bg-neutral-900/95 dark:text-blue-300">
                <div class="flex items-center gap-2 text-sm font-medium">
                    <FileUp :size="20" aria-hidden="true" />
                    <span>拖放 .txt 或 .md 文件到这里</span>
                </div>
            </div>

            <div v-if="uploadedFiles.length > 0" class="mb-2 flex flex-wrap gap-2 px-1">
                <div v-for="(file, index) in uploadedFiles" :key="`${file.name}-${index}`"
                    class="flex min-w-0 items-center gap-1.5 rounded-md border px-2 py-1 text-xs" :class="file.type === 'md'
                        ? 'border-orange-200 bg-orange-50 text-orange-700 dark:border-orange-800 dark:bg-orange-950 dark:text-orange-300'
                        : 'border-blue-200 bg-blue-50 text-blue-700 dark:border-blue-800 dark:bg-blue-950 dark:text-blue-300'
                        ">
                    <FileText :size="14" class="shrink-0" />

                    <span class="max-w-40 truncate font-medium">
                        {{ file.name }}
                    </span>

                    <span class="shrink-0 opacity-70">
                        {{ formatFileSize(file.size) }}
                    </span>

                    <button type="button"
                        class="ml-0.5 flex h-5 w-5 shrink-0 items-center justify-center rounded hover:bg-black/10 dark:hover:bg-white/10"
                        :aria-label="`移除 ${file.name}`" :title="`移除 ${file.name}`" @click="removeFile(index)">
                        <X :size="13" />
                    </button>
                </div>
            </div>

            <p v-if="uploadError" class="mb-2 px-2 text-xs text-red-600 dark:text-red-400" role="alert">
                {{ uploadError }}
            </p>

            <div class="flex min-h-10 items-end gap-2">
                <input ref="fileInput" type="file" class="hidden" accept=".txt,.md,text/plain,text/markdown" multiple
                    @change="handleFileChange" />

                <button type="button"
                    class="flex h-9 w-9 shrink-0 items-center justify-center rounded-md text-neutral-600 hover:bg-neutral-200 disabled:cursor-not-allowed disabled:opacity-40 dark:text-neutral-300 dark:hover:bg-neutral-700"
                    :disabled="!canAddFile" aria-label="添加文本附件" :title="uploadedFiles.length >= 5
                        ? '每条消息最多添加 5 个附件'
                        : '添加 .txt 或 .md 文件'
                        " @click="openFilePicker">
                    <LoaderCircle v-if="isUploading" :size="17" class="animate-spin" />
                    <FileUp v-else :size="17" />
                </button>

                <textarea v-model="content" rows="1"
                    class="max-h-40 min-h-10 flex-1 resize-none bg-transparent px-2 py-2 text-sm outline-none placeholder:text-neutral-500"
                    :disabled="disabled || isGenerating" :placeholder="isGenerating
                        ? 'SY Chat 正在回复...'
                        : '给 SY Chat 发送消息'
                        " @compositionstart="isComposing = true" @compositionend="isComposing = false"
                    @keydown.enter.exact.prevent="handleEnter" />

                <button v-if="isGenerating" type="button"
                    class="flex h-9 w-9 shrink-0 items-center justify-center rounded-md bg-neutral-950 text-white hover:opacity-80 dark:bg-white dark:text-neutral-950"
                    aria-label="停止生成" title="停止生成" @click="emit('stop')">
                    <Square :size="15" fill="currentColor" />
                </button>

                <button v-else type="submit"
                    class="flex h-9 w-9 shrink-0 items-center justify-center rounded-md bg-neutral-950 text-white transition-opacity hover:opacity-80 disabled:cursor-not-allowed disabled:opacity-40 dark:bg-white dark:text-neutral-950"
                    :disabled="!canSend" aria-label="发送消息" title="发送消息">
                    <Send :size="17" />
                </button>
            </div>
        </div>
    </form>
</template>