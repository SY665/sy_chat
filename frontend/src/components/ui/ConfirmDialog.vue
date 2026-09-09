<script setup lang="ts">
import {
    nextTick,
    onBeforeUnmount,
    onMounted,
    ref,
    watch,
} from 'vue'
import {
    AlertTriangle,
    LoaderCircle,
    X,
} from '@lucide/vue'

const props = withDefaults(
    defineProps<{
        open: boolean
        title: string
        description?: string
        confirmText?: string
        cancelText?: string
        loading?: boolean
        danger?: boolean
    }>(),
    {
        description: '',
        confirmText: '确认',
        cancelText: '取消',
        loading: false,
        danger: false,
    },
)

const emit = defineEmits<{
    confirm: []
    cancel: []
}>()

const cancelButton = ref<HTMLButtonElement | null>(null)

watch(
    () => props.open,
    async (open) => {
        if (!open) {
            return
        }

        // 默认聚焦取消按钮，避免用户误触破坏性操作。
        await nextTick()
        cancelButton.value?.focus()
    },
)

function cancel() {
    if (!props.loading) {
        emit('cancel')
    }
}

function confirm() {
    if (!props.loading) {
        emit('confirm')
    }
}

function handleKeydown(event: KeyboardEvent) {
    if (props.open && event.key === 'Escape') {
        cancel()
    }
}

onMounted(() => {
    window.addEventListener('keydown', handleKeydown)
})

onBeforeUnmount(() => {
    window.removeEventListener('keydown', handleKeydown)
})
</script>

<template>
    <Teleport to="body">
        <div v-if="open" class="fixed inset-0 z-[60] flex items-center justify-center bg-black/50 p-4"
            role="presentation" @click.self="cancel">
            <section
                class="w-full max-w-sm rounded-lg border border-neutral-200 bg-white shadow-xl dark:border-neutral-800 dark:bg-neutral-900"
                role="alertdialog" aria-modal="true" aria-labelledby="confirm-dialog-title"
                aria-describedby="confirm-dialog-description">
                <header class="flex items-start gap-3 border-b border-neutral-200 p-4 dark:border-neutral-800">
                    <div class="flex h-9 w-9 shrink-0 items-center justify-center rounded-md" :class="danger
                            ? 'bg-red-50 text-red-600 dark:bg-red-950/40 dark:text-red-400'
                            : 'bg-neutral-100 text-neutral-600 dark:bg-neutral-800 dark:text-neutral-300'
                        ">
                        <AlertTriangle :size="19" aria-hidden="true" />
                    </div>

                    <div class="min-w-0 flex-1">
                        <h2 id="confirm-dialog-title" class="text-base font-semibold text-neutral-950 dark:text-white">
                            {{ title }}
                        </h2>

                        <p v-if="description" id="confirm-dialog-description"
                            class="mt-1 text-sm leading-6 text-neutral-500 dark:text-neutral-400">
                            {{ description }}
                        </p>
                    </div>

                    <button type="button"
                        class="flex h-8 w-8 shrink-0 items-center justify-center rounded-md text-neutral-500 hover:bg-neutral-100 disabled:opacity-40 dark:hover:bg-neutral-800"
                        :disabled="loading" title="关闭" aria-label="关闭确认弹窗" @click="cancel">
                        <X :size="17" />
                    </button>
                </header>

                <footer class="flex justify-end gap-2 p-4">
                    <button ref="cancelButton" type="button"
                        class="h-9 rounded-md border border-neutral-300 px-4 text-sm font-medium text-neutral-700 hover:bg-neutral-100 disabled:cursor-not-allowed disabled:opacity-40 dark:border-neutral-700 dark:text-neutral-200 dark:hover:bg-neutral-800"
                        :disabled="loading" @click="cancel">
                        {{ cancelText }}
                    </button>

                    <button type="button"
                        class="flex h-9 min-w-20 items-center justify-center gap-2 rounded-md px-4 text-sm font-medium text-white disabled:cursor-not-allowed disabled:opacity-50"
                        :class="danger
                                ? 'bg-red-600 hover:bg-red-700'
                                : 'bg-neutral-950 hover:bg-neutral-800 dark:bg-white dark:text-neutral-950'
                            " :disabled="loading" @click="confirm">
                        <LoaderCircle v-if="loading" :size="15" class="animate-spin" aria-hidden="true" />
                        {{ confirmText }}
                    </button>
                </footer>
            </section>
        </div>
    </Teleport>
</template>