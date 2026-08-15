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
    updateConversationTitle,
} from '@/lib/api/conversation'
import { ApiRequestError } from '@/lib/api/client'

export const useConversationStore = defineStore('conversation', {
    state: () => ({
        conversations: [] as ConversationSummary[],
        currentConversation: null as ConversationDetail | null,
        isLoading: false,
        isMutating: false,
        error: null as string | null,
    }),

    getters: {
        hasConversations: (state) => state.conversations.length > 0,
    },

    actions: {
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
            this.isLoading = true
            this.error = null

            try {
                const conversation = await getConversation(id)
                this.currentConversation = conversation

                return conversation
            } catch (error) {
                this.error = getErrorMessage(error)
                throw error
            } finally {
                this.isLoading = false
            }
        },

        async create(title = '') {
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
                this.isMutating = false
            }
        },

        async rename(id: string, title: string) {
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
                this.isMutating = false
            }
        },

        async remove(id: string) {
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
                this.isMutating = false
            }
        },

        clearCurrentConversation() {
            this.currentConversation = null
        },

        reset() {
            this.conversations = []
            this.currentConversation = null
            this.isLoading = false
            this.isMutating = false
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
