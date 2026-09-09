import { createApp } from 'vue'
import { createPinia } from 'pinia'

import App from './App.vue'
import router from './router'
import { useToastStore } from './stores/toast'
import './styles/index.css'

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)
app.use(router)

const toastStore = useToastStore(pinia)

function reportUnexpectedError(
    source: string,
    error: unknown,
) {
    // 控制台保留错误详情，界面只显示不包含内部信息的提示。
    console.error(source, error)
    toastStore.error('应用发生了未预期的错误，请稍后重试')
}

function isAbortReason(reason: unknown): boolean {
    return (
        reason instanceof Error &&
        reason.name === 'AbortError'
    )
}

app.config.errorHandler = (error, _instance, info) => {
    reportUnexpectedError(
        `Unhandled Vue error: ${info}`,
        error,
    )
}

window.addEventListener('error', (event) => {
    reportUnexpectedError(
        'Unhandled window error',
        event.error ?? event.message,
    )
})

window.addEventListener('unhandledrejection', (event) => {
    // 主动停止流式响应属于正常行为，不应作为系统错误提示。
    if (isAbortReason(event.reason)) {
        return
    }

    reportUnexpectedError(
        'Unhandled promise rejection',
        event.reason,
    )
})

app.mount('#app')