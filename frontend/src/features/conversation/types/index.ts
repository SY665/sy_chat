import type { FileAttachment } from '@/features/chat/types'

export type ConversationMessageRole =
  | 'user'
  | 'assistant'
  | 'system'
  | 'tool'

export interface ConversationMessage {
  id: string
  role: ConversationMessageRole
  content: string
  thinking?: string
  attachments?: FileAttachment[]
  createdAt: string
}

export interface ConversationSummary {
  id: string
  title: string
  isPinned: boolean
  createdAt: string
  updatedAt: string
}

export interface ConversationDetail extends ConversationSummary {
  isShared: boolean
  shareToken?: string
  sharedAt?: string
  messages: ConversationMessage[]
}

export interface ConversationListData {
  conversations: ConversationSummary[]
}

export interface UpdateConversationData {
  id: string
  title: string
}

export interface UpdateConversationPinnedData {
  id: string
  isPinned: boolean
}

export interface DeleteConversationData {
  id: string
  message: string
}

export interface ShareConversationData {
  shareToken: string
  sharedAt: string
}

export interface UnshareConversationData {
  id: string
  isShared: false
}

export interface PublicShareMessage {
  id: string
  role: 'user' | 'assistant'
  content: string
  createdAt: string
}

export interface PublicShareData {
  title: string
  ownerName: string
  sharedAt?: string
  viewCount: number
  messages: PublicShareMessage[]
}