import type {
  ConversationDetail,
  ConversationListData,
  ConversationSummary,
  DeleteConversationData,
  UpdateConversationData,
  UpdateConversationPinnedData,
} from '@/features/conversation/types'

import { apiRequest } from './client'

export function listConversations(): Promise<ConversationListData> {
  return apiRequest<ConversationListData>('/conversations')
}

export function createConversation(
  title = '',
): Promise<ConversationSummary> {
  return apiRequest<ConversationSummary>('/conversations', {
    method: 'POST',
    body: JSON.stringify({ title }),
  })
}

export function getConversation(
  id: string,
): Promise<ConversationDetail> {
  return apiRequest<ConversationDetail>(
    `/conversations/${encodeURIComponent(id)}`,
  )
}

export function updateConversationTitle(
  id: string,
  title: string,
): Promise<UpdateConversationData> {
  return apiRequest<UpdateConversationData>(
    `/conversations/${encodeURIComponent(id)}`,
    {
      method: 'PATCH',
      body: JSON.stringify({ title }),
    },
  )
}

export function updateConversationPinned(
  id: string,
  isPinned: boolean,
): Promise<UpdateConversationPinnedData> {
  return apiRequest<UpdateConversationPinnedData>(
    `/conversations/${encodeURIComponent(id)}/pin`,
    {
      method: 'PATCH',
      body: JSON.stringify({ isPinned }),
    },
  )
}

export function deleteConversation(
  id: string,
): Promise<DeleteConversationData> {
  return apiRequest<DeleteConversationData>(
    `/conversations/${encodeURIComponent(id)}`,
    {
      method: 'DELETE',
    },
  )
}

