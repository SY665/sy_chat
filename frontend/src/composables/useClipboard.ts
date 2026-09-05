import { onBeforeUnmount, ref } from 'vue'

export function useClipboard() {
    const isCopying = ref(false)
    const copyFeedback = ref('')
    let timer: ReturnType<typeof setTimeout> | undefined
    let disposed = false

    async function copyText(text: string) {
        if (disposed || isCopying.value || !text) return

        clearTimeout(timer)
        isCopying.value = true
        copyFeedback.value = ''

        try {
            await navigator.clipboard.writeText(text)
            if (!disposed) copyFeedback.value = '已复制'
        } catch {
            if (!disposed) copyFeedback.value = '复制失败，请重试'
        } finally {
            // 请求结束时组件可能已卸载，避免重新创建定时器。
            if (!disposed) {
                isCopying.value = false
                timer = setTimeout(() => {
                    copyFeedback.value = ''
                    timer = undefined
                }, 2000)
            }
        }
    }

    onBeforeUnmount(() => {
        disposed = true
        clearTimeout(timer)
    })

    return { isCopying, copyFeedback, copyText }
}