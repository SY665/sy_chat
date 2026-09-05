import { defineStore } from 'pinia'

import type {
    ConversationDetail,
    ConversationSummary,
} from '@/features/conversation/types'
import {
    createConversation,
    deleteConversation,
    getConversation,
    listConversations,
    updateConversationPinned,
    updateConversationTitle,
} from '@/lib/api/conversation'
import { ApiRequestError } from '@/lib/api/client'

let latestConversationLoadID = 0

export const useConversationStore = defineStore('conversation', {
    state: () => ({
        conversations: [] as ConversationSummary[],
        searchQuery: '',
        currentConversation: null as ConversationDetail | null,
        isLoading: false,
        isCreating: false,
        isMutating: false,
        mutatingConversationId: null as string | null,
        error: null as string | null,
    }),

    getters: {
        hasConversations: (state) => state.conversations.length > 0,
        filteredConversations: (state) => {
            const query = state.searchQuery.trim().toLowerCase()

            if (!query) {
                return state.conversations
            }

            return state.conversations.filter((conversation) =>
                conversation.title.toLowerCase().includes(query),
            )
        },
    },

    actions: {
        setSearchQuery(query: string) {
            this.searchQuery = query
        },

        async loadConversations() {
            this.isLoading = true
            this.error = null

            try {
                const data = await listConversations()
                this.conversations = data.conversations
            } catch (error) {
                this.error = getErrorMessage(error)
                throw error
            } finally {
                this.isLoading = false
            }
        },

        async loadConversation(id: string) {
            const requestID = ++latestConversationLoadID

            this.isLoading = true
            this.error = null

            try {
                const conversation = await getConversation(id)

                // 较早请求晚返回时，不允许覆盖后来选择的会话。
                if (requestID === latestConversationLoadID) {
                    this.currentConversation = conversation
                }

                return conversation
            } catch (error) {
                if (requestID === latestConversationLoadID) {
                    this.error = getErrorMessage(error)
                }

                throw error
            } finally {
                // 旧请求不能关闭新请求的加载状态。
                if (requestID === latestConversationLoadID) {
                    this.isLoading = false
                }
            }
        },

        async create(title = '') {
            this.isCreating = true
            this.isMutating = true
            this.error = null

            try {
                const conversation = await createConversation(title)

                // 新建的普通对话放在已有置顶对话之后。
                const firstNormalIndex = this.conversations.findIndex(
                    (item) => !item.isPinned,
                )

                if (firstNormalIndex === -1) {
                    this.conversations.push(conversation)
                } else {
                    this.conversations.splice(firstNormalIndex, 0, conversation)
                }

                return conversation
            } catch (error) {
                this.error = getErrorMessage(error)
                throw error
            } finally {
                this.isCreating = false
                this.isMutating = false
            }
        },

        async rename(id: string, title: string) {
            this.mutatingConversationId = id
            this.isMutating = true
            this.error = null

            try {
                const updated = await updateConversationTitle(id, title)
                const conversation = this.conversations.find(
                    (item) => item.id === id,
                )

                if (conversation) {
                    conversation.title = updated.title
                    conversation.updatedAt = new Date().toISOString()
                }

                if (this.currentConversation?.id === id) {
                    this.currentConversation.title = updated.title
                    this.currentConversation.updatedAt = new Date().toISOString()
                }

                return updated
            } catch (error) {
                this.error = getErrorMessage(error)
                throw error
            } finally {
                this.mutatingConversationId = null
                this.isMutating = false
            }
        },

        async setPinned(id: string, isPinned: boolean) {
            this.mutatingConversationId = id
            this.isMutating = true
            this.error = null

            try {
                const updated = await updateConversationPinned(id, isPinned)
                const index = this.conversations.findIndex((item) => item.id === id)

                if (index !== -1) {
                    const [conversation] = this.conversations.splice(index, 1)

                    if (conversation) {
                        conversation.isPinned = updated.isPinned
                        conversation.updatedAt = new Date().toISOString()

                        if (updated.isPinned) {
                            // 最新置顶的会话放在整个列表最前面。
                            this.conversations.unshift(conversation)
                        } else {
                            // 取消置顶后，放在所有置顶会话之后。
                            const firstNormalIndex = this.conversations.findIndex(
                                (item) => !item.isPinned,
                            )

                            if (firstNormalIndex === -1) {
                                this.conversations.push(conversation)
                            } else {
                                this.conversations.splice(firstNormalIndex, 0, conversation)
                            }
                        }
                    }
                }

                if (this.currentConversation?.id === id) {
                    this.currentConversation.isPinned = updated.isPinned
                    this.currentConversation.updatedAt = new Date().toISOString()
                }

                return updated
            } catch (error) {
                this.error = getErrorMessage(error)
                throw error
            } finally {
                this.mutatingConversationId = null
                this.isMutating = false
            }
        },

        async remove(id: string) {
            this.mutatingConversationId = id
            this.isMutating = true
            this.error = null

            try {
                await deleteConversation(id)

                this.conversations = this.conversations.filter(
                    (item) => item.id !== id,
                )

                if (this.currentConversation?.id === id) {
                    this.currentConversation = null
                }
            } catch (error) {
                this.error = getErrorMessage(error)
                throw error
            } finally {
                this.mutatingConversationId = null
                this.isMutating = false
            }
        },

        clearCurrentConversation() {
            latestConversationLoadID += 1

            this.currentConversation = null
            this.isLoading = false
        },

        reset() {
            latestConversationLoadID += 1

            this.conversations = []
            this.searchQuery = ''
            this.currentConversation = null
            this.isLoading = false
            this.isCreating = false
            this.isMutating = false
            this.mutatingConversationId = null
            this.error = null
        },
    },
})

function getErrorMessage(error: unknown): string {
    if (error instanceof ApiRequestError) {
        return error.message
    }

    return '对话操作失败，请稍后重试'
}
