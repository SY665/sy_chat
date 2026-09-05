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