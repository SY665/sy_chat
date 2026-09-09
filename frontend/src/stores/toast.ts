import { defineStore } from 'pinia'

export type ToastTone = 'success' | 'error' | 'info'

export interface ToastItem {
    id: number
    message: string
    tone: ToastTone
}

interface ShowToastOptions {
    message: string
    tone?: ToastTone
    duration?: number
}

const defaultDuration = 4000
const errorDuration = 6000
const maximumToastCount = 4

let nextToastID = 0

const removalTimers = new Map<
    number,
    ReturnType<typeof window.setTimeout>
>()

export const useToastStore = defineStore('toast', {
    state: () => ({
        toasts: [] as ToastItem[],
    }),

    actions: {
        show(options: ShowToastOptions): number {
            const id = ++nextToastID
            const duration = options.duration ?? defaultDuration

            // 避免大量提示同时占满页面。
            if (this.toasts.length >= maximumToastCount) {
                const oldestToast = this.toasts[0]

                if (oldestToast) {
                    this.remove(oldestToast.id)
                }
            }

            this.toasts.push({
                id,
                message: options.message,
                tone: options.tone ?? 'info',
            })

            if (duration > 0) {
                const timer = window.setTimeout(() => {
                    this.remove(id)
                }, duration)

                removalTimers.set(id, timer)
            }

            return id
        },

        success(message: string, duration = defaultDuration): number {
            return this.show({
                message,
                tone: 'success',
                duration,
            })
        },

        error(message: string, duration = errorDuration): number {
            return this.show({
                message,
                tone: 'error',
                duration,
            })
        },

        info(message: string, duration = defaultDuration): number {
            return this.show({
                message,
                tone: 'info',
                duration,
            })
        },

        remove(id: number) {
            this.toasts = this.toasts.filter((toast) => toast.id !== id)

            const timer = removalTimers.get(id)

            if (timer !== undefined) {
                window.clearTimeout(timer)
                removalTimers.delete(id)
            }
        },

        clear() {
            for (const timer of removalTimers.values()) {
                window.clearTimeout(timer)
            }

            removalTimers.clear()
            this.toasts = []
        },
    },
})