import { defineStore } from "pinia";

import type { ChatMessage, FileAttachment } from "../types";
import { ApiRequestError, isAbortError } from "@/lib/api/client";
import {
    requestChatStream,
    truncateMessagesFromUserMessage,
} from "@/lib/api/chat";
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
            modelId = '',
            enableThinking = false,
            attachments: FileAttachment[] = [],
            enableWebSearch = false,
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
                attachments: attachments.length > 0 ? [...attachments] : undefined,
                createdAt: new Date().toISOString(),
                status: 'sending',
            }

            this.errorMessage = null
            this.isGenerating = true
            this.messages.push(pendingMessage)

            const getPendingAssistantMessage = (): ChatMessage | undefined => {
                let assistantMessage = this.messages.find(
                    (message) => message.id === pendingAssistantID,
                )

                if (assistantMessage) {
                    return assistantMessage
                }

                const userMessageStillVisible = this.messages.some(
                    (message) => message.id === pendingUserID,
                )
                if (!userMessageStillVisible) {
                    return undefined
                }

                assistantMessage = {
                    id: pendingAssistantID,
                    role: 'assistant',
                    content: '',
                    createdAt: new Date().toISOString(),
                }

                this.messages.push(assistantMessage)
                return assistantMessage
            }

            try {
                //调用后端
                const response = await requestChatStream(
                    conversationID,
                    value,
                    (chunk) => {
                        const assistantMessage = getPendingAssistantMessage()

                        if (assistantMessage) {
                            assistantMessage.content += chunk
                        }
                    },
                    controller.signal,
                    modelId,
                    (chunk) => {
                        const assistantMessage = getPendingAssistantMessage()

                        if (assistantMessage) {
                            assistantMessage.thinking =
                                (assistantMessage.thinking ?? '') + chunk
                        }
                    },
                    enableThinking,
                    attachments,
                    enableWebSearch,
                    (toolEvent) => {
                        const assistantMessage =
                            getPendingAssistantMessage()

                        if (!assistantMessage) {
                            return
                        }

                        const toolEvents =
                            assistantMessage.toolEvents ?? []

                        const existingIndex = toolEvents.findIndex(
                            (event) =>
                                event.toolCallId ===
                                toolEvent.toolCallId,
                        )

                        if (existingIndex === -1) {
                            toolEvents.push(toolEvent)
                        } else {
                            toolEvents.splice(
                                existingIndex,
                                1,
                                toolEvent,
                            )
                        }

                        assistantMessage.toolEvents = toolEvents
                    },
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
                    const streamedToolEvents = this.messages[assistantIndex]?.toolEvents

                    this.messages.splice(
                        assistantIndex,
                        1,
                        {
                            ...response.assistantMessage,
                            toolEvents:
                                response.assistantMessage.toolEvents ??
                                streamedToolEvents,
                        },
                    )
                } else if (userIndex !== -1) {
                    this.messages.push(response.assistantMessage)
                }
            } catch (error) {
                const pendingUserMessage = this.messages.find(
                    (message) => message.id === pendingUserID,
                )
                const pendingAssistantMessage = this.messages.find(
                    (message) => message.id === pendingAssistantID,
                )

                if (isAbortError(error)) {
                    // 后端会持久化用户消息和已生成的部分回复，
                    // 前端同时保留当前内容，避免点击停止后消息突然消失。
                    if (pendingUserMessage) {
                        pendingUserMessage.status = 'stopped'
                    }
                    if (pendingAssistantMessage) {
                        pendingAssistantMessage.status = 'stopped'
                    }

                    this.errorMessage = null
                    throw error
                }

                // 普通生成失败不会持久化未完成的 AI 回复。
                this.messages = this.messages.filter(
                    (message) => message.id !== pendingAssistantID,
                )

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

        /**
         * 删除数据库中的旧消息分支，并使用指定内容重新生成回复。
         *
         * 该方法只处理已经持久化的用户消息；replacementContent
         * 可以是原始内容，也可以是用户编辑后的内容。
         */
        async truncateAndResend(
            conversationID: string,
            message: ChatMessage,
            replacementContent = message.content,
            modelId = '',
            enableThinking = false,
            enableWebSearch = false,
        ) {
            const content = replacementContent.trim()

            if (
                this.isGenerating ||
                !conversationID ||
                !content ||
                message.role !== 'user' ||
                message.status !== undefined
            ) {
                return
            }

            const messageIndex = this.messages.findIndex(
                (item) => item.id === message.id,
            )
            if (messageIndex === -1) {
                return
            }

            this.errorMessage = null

            try {
                // 先由后端删除目标用户消息及其后的持久化分支。
                await truncateMessagesFromUserMessage(
                    conversationID,
                    message.id,
                )

                // 后端成功后再同步本地状态，避免请求失败时页面提前丢失消息。
                this.messages.splice(messageIndex)

                await this.sendMessage(
                    conversationID,
                    content,
                    modelId,
                    enableThinking,
                    message.attachments ?? [],
                    enableWebSearch,
                )
            } catch (error) {
                this.errorMessage = getErrorMessage(error)
                throw error
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