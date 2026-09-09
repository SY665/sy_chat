<script setup lang="ts">
import { computed, ref } from 'vue'
import { Download, ImageOff, LoaderCircle } from '@lucide/vue'

import type { GeneratedImage } from '../types'

const props = defineProps<{
    image: GeneratedImage
}>()

const isLoaded = ref(false)
const loadFailed = ref(false)

const safeImageURL = computed(() => {
    const value = props.image.url.trim()

    // 生成图片只允许使用当前站点下的相对资源地址。
    if (!value.startsWith('/') || value.startsWith('//')) {
        return ''
    }

    const parsedURL = new URL(value, window.location.origin)
    if (parsedURL.origin !== window.location.origin) {
        return ''
    }

    return `${parsedURL.pathname}${parsedURL.search}`
})

const aspectRatio = computed(() => {
    return `${props.image.width} / ${props.image.height}`
})

const hasError = computed(() => {
    return loadFailed.value || safeImageURL.value === ''
})
</script>

<template>
    <figure
        class="relative my-3 w-full max-w-2xl overflow-hidden rounded-lg border border-neutral-200 bg-neutral-100 dark:border-neutral-700 dark:bg-neutral-900"
        :style="{ aspectRatio }">
        <div v-if="!isLoaded && !hasError" class="absolute inset-0 flex items-center justify-center text-neutral-500">
            <LoaderCircle :size="24" class="animate-spin" aria-hidden="true" />
            <span class="sr-only">图片加载中</span>
        </div>

        <div v-if="hasError"
            class="absolute inset-0 flex flex-col items-center justify-center gap-2 text-sm text-neutral-500">
            <ImageOff :size="24" aria-hidden="true" />
            <span>图片加载失败</span>
        </div>

        <img v-if="safeImageURL && !loadFailed" :src="safeImageURL" alt="AI 生成图片" class="h-full w-full object-contain"
            :class="{ 'opacity-0': !isLoaded }" loading="lazy" @load="isLoaded = true" @error="loadFailed = true">

        <a v-if="safeImageURL && !hasError" :href="safeImageURL" download
            class="absolute right-2 top-2 flex h-8 w-8 items-center justify-center rounded-md bg-white/90 text-neutral-700 shadow-sm hover:bg-white dark:bg-neutral-900/90 dark:text-neutral-200"
            title="下载图片" aria-label="下载生成图片">
            <Download :size="16" aria-hidden="true" />
        </a>
    </figure>
</template>