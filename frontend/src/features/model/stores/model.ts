import { defineStore } from 'pinia'
import type { AIModel } from '@/features/model/types'
import { listModels } from '@/lib/api/model'
import { ApiRequestError } from '@/lib/api/client'

const selectedModelStorageKey = 'sy-chat:selected-model'
const thinkingEnabledStorageKey = 'sy-chat:thinking-enabled'
const webSearchEnabledStorageKey = 'sy-chat:web-search-enabled'

function readStoredModelId(): string {
    try {
        return window.localStorage.getItem(selectedModelStorageKey) ?? ''
    } catch {
        // 浏览器禁用本地存储时，仍允许模型选择在当前页面工作。
        return ''
    }
}

function writeStoredModelId(modelId: string) {
    try {
        window.localStorage.setItem(selectedModelStorageKey, modelId)
    } catch {
        // 存储失败不应阻止用户继续聊天。
    }
}

function readStoredThinkingEnabled(): boolean {
    try {
        return window.localStorage.getItem(thinkingEnabledStorageKey) === 'true'
    } catch {
        return false
    }
}

function writeStoredThinkingEnabled(enabled: boolean) {
    try {
        window.localStorage.setItem(
            thinkingEnabledStorageKey,
            String(enabled),
        )
    } catch {
        // 存储失败时仅失去跨页面持久化，不影响当前聊天。
    }
}

function readStoredWebSearchEnabled(): boolean {
    try {
        return window.localStorage.getItem(
            webSearchEnabledStorageKey,
        ) === 'true'
    } catch {
        return false
    }
}

function writeStoredWebSearchEnabled(enabled: boolean) {
    try {
        window.localStorage.setItem(
            webSearchEnabledStorageKey,
            String(enabled),
        )
    } catch {
        // 存储失败时仅影响跨页面持久化。
    }
}

export const useModelStore = defineStore('model', {
    state: () => ({
        models: [] as AIModel[],
        selectedModelId: readStoredModelId(),
        thinkingEnabled: readStoredThinkingEnabled(),
        webSearchEnabled: readStoredWebSearchEnabled(),
        webSearchAvailable: false,
        isLoading: false,
        error: null as string | null
    }),

    getters: {
        selectedModel: (state) => {
            return state.models.find((model) => {
                return model.id === state.selectedModelId
            })
        },
        selectedModelSupportsThinking: (state) => {
            return state.models.find((model) => {
                return model.id === state.selectedModelId
            })?.supportsThinking ?? false
        },
        effectiveThinkingEnabled: (state) => {
            const selectedModel = state.models.find((model) => {
                return model.id === state.selectedModelId
            })

            return state.thinkingEnabled &&
                (selectedModel?.supportsThinking ?? false)
        },
        effectiveWebSearchEnabled: (state) => {
            return state.webSearchEnabled && state.webSearchAvailable
        },
    },

    actions: {
        async loadModels() {
            if (this.isLoading) return
            this.isLoading = true
            this.error = null

            try {
                const data = await listModels()
                this.models = data.models
                this.webSearchAvailable = data.webSearchAvailable

                // 刷新列表时保留有效选择，否则使用默认模型或首个模型。
                const hasSelectedModel = this.models.some((model) => {
                    return model.id === this.selectedModelId
                })

                if (!hasSelectedModel) {
                    const defaultModel = this.models.find((model) => {
                        return model.id === data.defaultModelId
                    })

                    this.selectedModelId =
                        defaultModel?.id ?? this.models[0]?.id ?? ''
                }

                writeStoredModelId(this.selectedModelId)

            } catch (error) {
                this.error = error instanceof ApiRequestError
                    ? error.message
                    : '模型列表加载失败，请稍后重试'
                throw error
            } finally {
                this.isLoading = false
            }
        },
        selectModel(id: string) {
            // 仅允许选择后端返回的模型。
            if (!this.models.some((model) => model.id === id)) {
                return
            }

            this.selectedModelId = id
            writeStoredModelId(id)
        },
        setThinkingEnabled(enabled: boolean) {
            // 保存用户偏好，实际请求还会通过模型能力进行限制。
            this.thinkingEnabled = enabled
            writeStoredThinkingEnabled(enabled)
        },
        setWebSearchEnabled(enabled: boolean) {
            this.webSearchEnabled = enabled
            writeStoredWebSearchEnabled(enabled)
        },
    }
})
