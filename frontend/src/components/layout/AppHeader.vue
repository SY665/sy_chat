<script setup lang="ts">
withDefaults(
    defineProps<{
        title?: string
    }>(),
    {
        title: '新对话',
    },
)
</script>

<template>
    <header class="flex h-14 shrink-0 items-center border-b border-neutral-200 px-4 dark:border-neutral-800">
        <h1 class="truncate text-sm font-medium">
            {{ title }}
        </h1>
    </header>
</template>