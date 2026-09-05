import { defineStore } from "pinia";

import type { ChatMessage } from "../types";
import { ApiRequestError, isAbortError } from "@/lib/api/client";
import { requestChatStream } from "@/lib/api/chat";
import { markRaw } from 'vue'

export const useChatStore = defineStore('chat', {
    state: () => ({
        messages: [] as ChatMessage[],
        isGenerating: false,
        errorMessage: null as string | null,
        abortController: null as AbortController | null,
    }),

    getters: {
        hasMessages: (state) => state.messages.length > 0,
    },

    actions: {
        async sendMessage(
            conversationID: string,
            content: string,
            modelId = ''
        ) {
            const value = content.trim()

            if (!value || this.isGenerating || !conversationID) {
                return
            }

            const controller = markRaw(new AbortController())
            this.abortController = controller

            const requestID = Date.now()
            const pendingUserID = `pending-user-${requestID}`
            const pendingAssistantID = `pending-assistant-${requestID}`

            // 后端返回前先显示用户消息，让聊天交互得到即时反馈。
            const pendingMessage: ChatMessage = {
                id: pendingUserID,
                role: 'user',
                content: value,
                createdAt: new Date().toISOString(),
                status: 'sending',
            }

            this.errorMessage = null
            this.isGenerating = true
            this.messages.push(pendingMessage)

            try {
                //调用后端
                const response = await requestChatStream(
                    conversationID,
                    value,
                    (chunk) => {
                        let assistantMessage = this.messages.find(
                            (message) => message.id === pendingAssistantID,
                        )

                        if (!assistantMessage) {
                            const userMessageStillVisible = this.messages.some(
                                (message) => message.id === pendingUserID,
                            )
                            if (!userMessageStillVisible) {
                                return
                            }

                            this.messages.push({
                                id: pendingAssistantID,
                                role: 'assistant',
                                content: '',
                                createdAt: new Date().toISOString(),
                            })

                            assistantMessage = this.messages.find(
                                (message) => message.id === pendingAssistantID,
                            )
                        }

                        if (assistantMessage) {
                            assistantMessage.content += chunk
                        }
                    },
                    controller.signal,
                    modelId,
                )

                const userIndex = this.messages.findIndex(
                    (message) => message.id === pendingUserID,
                )
                if (userIndex !== -1) {
                    this.messages.splice(
                        userIndex,
                        1,
                        response.userMessage,
                    )
                }

                const assistantIndex = this.messages.findIndex(
                    (message) => message.id === pendingAssistantID,
                )
                if (assistantIndex !== -1) {
                    this.messages.splice(
                        assistantIndex,
                        1,
                        response.assistantMessage,
                    )
                } else if (userIndex !== -1) {
                    this.messages.push(response.assistantMessage)
                }
            } catch (error) {
                // 未完成的 AI 回复不会写入数据库，因此从页面移除。
                this.messages = this.messages.filter(
                    (message) => message.id !== pendingAssistantID,
                )

                const pendingUserMessage = this.messages.find(
                    (message) => message.id === pendingUserID,
                )

                if (isAbortError(error)) {
                    if (pendingUserMessage) {
                        pendingUserMessage.status = 'stopped'
                    }

                    this.errorMessage = null
                    throw error
                }

                // 保留原始输入，下一步允许用户直接重试。
                if (pendingUserMessage) {
                    pendingUserMessage.status = 'failed'
                }

                this.errorMessage = getErrorMessage(error)
                throw error
            } finally {
                // 只清除属于当前请求的控制器，避免误伤后续请求。
                if (this.abortController === controller) {
                    this.abortController = null
                }

                this.isGenerating = false
            }
        },

        removeMessage(id: string) {
            this.messages = this.messages.filter(
                (message) => message.id !== id,
            )
        },

        setMessages(messages: ChatMessage[]) {
            this.messages = messages
            this.errorMessage = null
        },

        stopGenerating() {
            this.abortController?.abort()
        },

        clearMessages() {
            // 切换或删除会话时，停止仍在进行的模型请求。
            this.stopGenerating()

            this.messages = []
            this.errorMessage = null
        },
    },
})

function getErrorMessage(error: unknown): string {
    if (error instanceof ApiRequestError) {
        return error.message
    }

    return '发送消息失败，请稍后重试'
}